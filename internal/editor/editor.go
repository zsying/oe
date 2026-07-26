package editor

import (
	"github.com/zsying/oe/internal/buffer"
	"github.com/zsying/oe/internal/clipboard"
)

const MaxUndo = 100

// OpType classifies the mutation a snapshot was taken before, so that
// consecutive rune insertions can be coalesced into a single undo step.
type OpType int

const (
	OpGeneric OpType = iota
	OpInsert
)

// Editor is the core editor state.
type Editor struct {
	Buffer      *buffer.Buffer
	Cursor      *buffer.Cursor
	Selection   *buffer.Selection
	Mode        Mode
	Commands    *CommandRegistry
	Clip        clipboard.Clipboard
	ShowLineNum bool
	SearchQuery string

	undoStack []undoSnapshot
	redoStack []undoSnapshot

	// coalesceX/Y is the cursor position where the next rune insert would
	// continue the current typing burst (-1 means no active burst).
	coalesceX int
	coalesceY int
}

// undoSnapshot captures the buffer state immediately before a mutation so it
// can be restored by undo. It also records the cursor and the modified flag.
type undoSnapshot struct {
	lines    []string
	cursorX  int
	cursorY  int
	modified bool
	op       OpType
}

// New creates a new Editor with default state.
func New() *Editor {
	ed := &Editor{
		Buffer:      buffer.New(),
		Cursor:      &buffer.Cursor{},
		Selection:   &buffer.Selection{},
		Mode:        ModeView,
		ShowLineNum: true,
		Clip:        clipboard.New(),
		coalesceX:   -1,
		coalesceY:   -1,
	}
	ed.Commands = NewCommandRegistry(ed)
	return ed
}

// --- Exported wrappers for screen package ---

// ToggleMode switches between View and Edit modes.
func (e *Editor) ToggleMode() { e.cmdToggleMode() }

// Save persists the current buffer to disk.
func (e *Editor) Save() error { return e.cmdSave() }

// --- Command stubs (implemented in later tasks) ---

func (e *Editor) cmdOpen() error {
	// Overridden by screen to show file dialog
	return nil
}

func (e *Editor) cmdSave() error {
	if e.Buffer.Filename() == "" {
		// Use the overridden handler from the command registry
		if cmd := e.Commands.Find("file.saveAs"); cmd != nil {
			return cmd.Handler()
		}
		return nil
	}
	return e.SaveFile()
}

func (e *Editor) cmdSaveAs() error {
	// Overridden by screen to show file dialog
	return nil
}

func (e *Editor) cmdNew() error {
	// Overridden by screen to confirm unsaved changes
	e.SaveSnapshot(OpGeneric)
	e.Buffer = buffer.New()
	e.Cursor = &buffer.Cursor{}
	return nil
}

func (e *Editor) cmdQuit() error {
	// Overridden by screen to confirm unsaved changes
	return nil
}

func (e *Editor) cmdUndo() error {
	e.RestoreSnapshot()
	return nil
}
func (e *Editor) cmdRedo() error {
	e.Redo()
	return nil
}
func (e *Editor) cmdCut() error {
	if e.Mode != ModeEdit || !e.Selection.Active() {
		return nil
	}
	text := e.Selection.Text(e.Buffer)
	sx, sy := e.Selection.StartX, e.Selection.StartY
	// Place cursor at the selection start, then snapshot so undo restores both
	// the removed text and the cursor position.
	e.Cursor.X, e.Cursor.Y = sx, sy
	e.SaveSnapshot(OpGeneric)
	if err := e.Clip.Write(text); err != nil {
		return err
	}
	e.Selection.Delete(e.Buffer)
	return nil
}

func (e *Editor) cmdCopy() error {
	if !e.Selection.Active() {
		return nil
	}
	text := e.Selection.Text(e.Buffer)
	return e.Clip.Write(text)
}

func (e *Editor) cmdPaste() error {
	if e.Mode != ModeEdit {
		return nil
	}
	text, err := e.Clip.Read()
	if err != nil || text == "" {
		return err
	}
	// Delete selection if active before pasting
	if e.Selection.Active() {
		sx, sy := e.Selection.StartX, e.Selection.StartY
		e.Selection.Delete(e.Buffer)
		e.Cursor.X, e.Cursor.Y = sx, sy
	}
	e.SaveSnapshot(OpGeneric)
	x, y := e.Buffer.InsertText(text, e.Cursor.X, e.Cursor.Y)
	e.Cursor.X, e.Cursor.Y = x, y
	return nil
}

func (e *Editor) cmdDelete() error {
	if e.Mode != ModeEdit {
		return nil
	}
	// When a selection is active, Delete removes the whole selection.
	if e.DeleteSelection() {
		return nil
	}
	// Forward delete is a no-op at the very end of the buffer
	if e.Cursor.X >= e.Buffer.LineLen(e.Cursor.Y) && e.Cursor.Y >= e.Buffer.Lines()-1 {
		return nil
	}
	e.SaveSnapshot(OpGeneric)
	e.Buffer.DeleteForward(e.Cursor.X, e.Cursor.Y)
	return nil
}

// DeleteSelection removes the active selection (if any), moving the cursor to
// the selection start, and records a single undo snapshot. Returns true if a
// selection was present and removed. Used so that typing, Backspace, Delete,
// Enter, or Tab over a selection replaces/deletes the whole selection.
func (e *Editor) DeleteSelection() bool {
	if !e.Selection.Active() {
		return false
	}
	sx, sy := e.Selection.StartX, e.Selection.StartY
	e.Cursor.X, e.Cursor.Y = sx, sy
	e.SaveSnapshot(OpGeneric)
	e.Selection.Delete(e.Buffer)
	return true
}

func (e *Editor) cmdSelectAll() error {
	buf := e.Buffer
	lastLine := buf.Lines() - 1
	e.Selection.Begin(0, 0)
	e.Selection.Extend(buf.LineLen(lastLine), lastLine)
	return nil
}
func (e *Editor) cmdToggleLineNum() error {
	e.ShowLineNum = !e.ShowLineNum
	return nil
}
func (e *Editor) cmdToggleMode() error {
	if e.Mode == ModeView {
		e.Mode = ModeEdit
	} else {
		e.Mode = ModeView
	}
	return nil
}
func (e *Editor) cmdFind() error         { return nil }
func (e *Editor) cmdFindNext() error     { return nil }
func (e *Editor) cmdReplace() error      { return nil }
func (e *Editor) cmdHelpKeyboard() error { return nil }

// --- Snapshot helpers ---

// captureSnapshot returns the current buffer/cursor/modified state as a
// snapshot, for undo/redo bookkeeping.
func (e *Editor) captureSnapshot() undoSnapshot {
	return undoSnapshot{
		lines:    e.Buffer.LinesSlice(),
		cursorX:  e.Cursor.X,
		cursorY:  e.Cursor.Y,
		modified: e.Buffer.Modified(),
		op:       OpGeneric,
	}
}

// SaveSnapshot records the current buffer and cursor state so it can be
// restored by undo. Consecutive rune insertions (OpInsert) at the same
// location coalesce into the previous snapshot, so a typing burst becomes a
// single undo step. Any new edit invalidates the redo history.
func (e *Editor) SaveSnapshot(op OpType) {
	e.redoStack = nil
	if op == OpInsert && e.coalesceX == e.Cursor.X && e.coalesceY == e.Cursor.Y {
		// Continue the current typing burst: advance the coalesce point and
		// skip pushing a new snapshot (the existing one already covers it).
		e.coalesceX++
		return
	}
	// Any non-insert op breaks the current burst.
	if op != OpInsert {
		e.coalesceX, e.coalesceY = -1, -1
	}
	if len(e.undoStack) >= MaxUndo {
		e.undoStack = e.undoStack[1:]
	}
	e.undoStack = append(e.undoStack, undoSnapshot{
		lines:    e.Buffer.LinesSlice(),
		cursorX:  e.Cursor.X,
		cursorY:  e.Cursor.Y,
		modified: e.Buffer.Modified(),
		op:       op,
	})
	if op == OpInsert {
		// The coalesce point is the cursor right after this single insert.
		e.coalesceX = e.Cursor.X + 1
		e.coalesceY = e.Cursor.Y
	}
}

// restoreFrom applies a snapshot's buffer/cursor/modified state to the editor.
func (e *Editor) restoreFrom(snap undoSnapshot) {
	e.Buffer.SetLines(snap.lines)
	e.Buffer.SetModified(snap.modified)
	e.Cursor.X = snap.cursorX
	e.Cursor.Y = snap.cursorY
	e.Selection.Clear()
	e.coalesceX, e.coalesceY = -1, -1
}

// RestoreSnapshot (undo) reverts the buffer to the most recent snapshot and
// pushes the current state onto the redo stack.
func (e *Editor) RestoreSnapshot() bool {
	if len(e.undoStack) == 0 {
		return false
	}
	cur := e.captureSnapshot()
	snap := e.undoStack[len(e.undoStack)-1]
	e.undoStack = e.undoStack[:len(e.undoStack)-1]
	e.restoreFrom(snap)
	e.redoStack = append(e.redoStack, cur)
	return true
}

// Redo re-applies the most recently undone change.
func (e *Editor) Redo() bool {
	if len(e.redoStack) == 0 {
		return false
	}
	cur := e.captureSnapshot()
	snap := e.redoStack[len(e.redoStack)-1]
	e.redoStack = e.redoStack[:len(e.redoStack)-1]
	e.restoreFrom(snap)
	e.undoStack = append(e.undoStack, cur)
	return true
}
