package editor

import (
	"github.com/user/editor/internal/buffer"
	"github.com/user/editor/internal/clipboard"
)

const MaxUndo = 100

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

	undoStack []*buffer.Buffer
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
		return e.cmdSaveAs()
	}
	return e.SaveFile()
}

func (e *Editor) cmdSaveAs() error {
	// Overridden by screen to show file dialog
	return nil
}

func (e *Editor) cmdNew() error {
	// Overridden by screen to confirm unsaved changes
	e.Buffer = buffer.New()
	e.Cursor = &buffer.Cursor{}
	return nil
}

func (e *Editor) cmdQuit() error {
	// Overridden by screen to confirm unsaved changes
	return nil
}

func (e *Editor) cmdUndo() error { return nil }
func (e *Editor) cmdCut() error {
	if e.Mode != ModeEdit || !e.Selection.Active() {
		return nil
	}
	text := e.Selection.Text(e.Buffer)
	if err := e.Clip.Write(text); err != nil {
		return err
	}
	e.Selection.Delete(e.Buffer)
	e.Cursor.X, e.Cursor.Y = e.Selection.StartX, e.Selection.StartY
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
		e.Selection.Delete(e.Buffer)
		e.Cursor.X, e.Cursor.Y = e.Selection.StartX, e.Selection.StartY
		e.Selection.Clear()
	}
	e.SaveSnapshot()
	x, y := e.Buffer.InsertText(text, e.Cursor.X, e.Cursor.Y)
	e.Cursor.X, e.Cursor.Y = x, y
	return nil
}

func (e *Editor) cmdDelete() error {
	if e.Mode != ModeEdit {
		return nil
	}
	if e.Selection.Active() {
		e.Selection.Delete(e.Buffer)
		e.Cursor.X, e.Cursor.Y = e.Selection.StartX, e.Selection.StartY
		e.Selection.Clear()
		return nil
	}
	e.Buffer.DeleteForward(e.Cursor.X, e.Cursor.Y)
	return nil
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

// --- Snapshot helpers ---

func (e *Editor) SaveSnapshot() {
	if len(e.undoStack) >= MaxUndo {
		e.undoStack = e.undoStack[1:]
	}
	snap := buffer.New()
	snap.SetLines(e.Buffer.LinesSlice())
	e.undoStack = append(e.undoStack, snap)
}

func (e *Editor) RestoreSnapshot() bool {
	if len(e.undoStack) == 0 {
		return false
	}
	snap := e.undoStack[len(e.undoStack)-1]
	e.undoStack = e.undoStack[:len(e.undoStack)-1]
	e.Buffer.SetLines(snap.LinesSlice())
	e.Buffer.SetModified(true)
	return true
}
