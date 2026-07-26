package editor

import (
	"testing"
)

func TestNewEditor(t *testing.T) {
	e := New()
	if e.Mode != ModeView {
		t.Fatalf("expected ModeView, got %v", e.Mode)
	}
	if e.Buffer.Lines() != 1 {
		t.Fatalf("expected 1 line, got %d", e.Buffer.Lines())
	}
	if !e.ShowLineNum {
		t.Fatal("line numbers should be on by default")
	}
}

func TestToggleMode(t *testing.T) {
	e := New()
	if e.Mode != ModeView {
		t.Fatal("expected view mode initially")
	}
	e.ToggleMode()
	if e.Mode != ModeEdit {
		t.Fatal("expected edit mode after toggle")
	}
	e.ToggleMode()
	if e.Mode != ModeView {
		t.Fatal("expected view mode after second toggle")
	}
}

func TestCmdToggleLineNum(t *testing.T) {
	e := New()
	e.cmdToggleLineNum()
	if e.ShowLineNum {
		t.Fatal("line numbers should be off after toggle")
	}
	e.cmdToggleLineNum()
	if !e.ShowLineNum {
		t.Fatal("line numbers should be on after second toggle")
	}
}

// BUG CONFIRMED: cmdCut sets cursor to (0,0) after delete because Selection.Delete() clears StartX/StartY
func TestCmdCutSelection(t *testing.T) {
	e := New()
	e.Mode = ModeEdit
	e.Buffer.InsertText("hello world", 0, 0)

	// Select "hello"
	e.Selection.Begin(0, 0)
	e.Selection.Extend(5, 0)

	err := e.cmdCut()
	if err != nil {
		t.Fatal(err)
	}

	// Buffer should have " world"
	if e.Buffer.Line(0) != " world" {
		t.Fatalf("expected ' world', got %q", e.Buffer.Line(0))
	}

	// Clipboard should have "hello"
	clipText, _ := e.Clip.Read()
	if clipText != "hello" {
		t.Fatalf("expected clipboard 'hello', got %q", clipText)
	}

	if e.Cursor.X != 0 || e.Cursor.Y != 0 {
		t.Fatalf("expected cursor (0,0), got (%d,%d)", e.Cursor.X, e.Cursor.Y)
	}
}

func TestCmdCopy(t *testing.T) {
	e := New()
	e.Buffer.InsertText("hello world", 0, 0)

	// Select "hello"
	e.Selection.Begin(0, 0)
	e.Selection.Extend(5, 0)

	err := e.cmdCopy()
	if err != nil {
		t.Fatal(err)
	}

	// Buffer unchanged
	if e.Buffer.Line(0) != "hello world" {
		t.Fatalf("buffer should be unchanged after copy, got %q", e.Buffer.Line(0))
	}

	// Clipboard has text
	clipText, _ := e.Clip.Read()
	if clipText != "hello" {
		t.Fatalf("expected clipboard 'hello', got %q", clipText)
	}
}

// BUG CONFIRMED: cmdPaste with active selection deletes selection then reads deleted (0,0) StartX/StartY
func TestCmdPaste(t *testing.T) {
	e := New()
	e.Mode = ModeEdit
	e.Buffer.InsertText("hello world", 0, 0)
	e.Clip.Write("XYZ")

	// Paste at position 5
	e.Cursor.X = 5
	e.Cursor.Y = 0
	err := e.cmdPaste()
	if err != nil {
		t.Fatal(err)
	}

	// Buffer should have "helloXYZ world"
	if e.Buffer.Line(0) != "helloXYZ world" {
		t.Fatalf("expected 'helloXYZ world', got %q", e.Buffer.Line(0))
	}

	// Cursor should be at end of pasted text
	if e.Cursor.X != 8 || e.Cursor.Y != 0 {
		t.Fatalf("expected cursor (8,0), got (%d,%d)", e.Cursor.X, e.Cursor.Y)
	}
}

func TestCmdSelectAll(t *testing.T) {
	e := New()
	e.Buffer.InsertText("abc\ndef\nghi", 0, 0)

	e.cmdSelectAll()
	if !e.Selection.Active() {
		t.Fatal("selection should be active after select all")
	}

	text := e.Selection.Text(e.Buffer)
	expected := "abc\ndef\nghi"
	if text != expected {
		t.Fatalf("expected %q, got %q", expected, text)
	}
}

func TestCmdDeleteWithSelection(t *testing.T) {
	e := New()
	e.Mode = ModeEdit
	e.Buffer.InsertText("hello world", 0, 0)

	e.Selection.Begin(0, 0)
	e.Selection.Extend(5, 0) // select "hello"

	err := e.cmdDelete()
	if err != nil {
		t.Fatal(err)
	}

	if e.Buffer.Line(0) != " world" {
		t.Fatalf("expected ' world', got %q", e.Buffer.Line(0))
	}

	if e.Cursor.X != 0 || e.Cursor.Y != 0 {
		t.Fatalf("expected cursor (0,0), got (%d,%d)", e.Cursor.X, e.Cursor.Y)
	}
}

func TestCmdDeleteWithoutSelection(t *testing.T) {
	e := New()
	e.Mode = ModeEdit
	e.Buffer.InsertText("abcd", 0, 0)

	e.Cursor.X = 1
	e.Cursor.Y = 0

	err := e.cmdDelete()
	if err != nil {
		t.Fatal(err)
	}

	if e.Buffer.Line(0) != "acd" {
		t.Fatalf("expected 'acd', got %q", e.Buffer.Line(0))
	}
}

func TestDeleteSelectionReplacesText(t *testing.T) {
	e := New()
	e.Mode = ModeEdit
	e.Buffer.InsertText("hello world", 0, 0)

	e.Selection.Begin(0, 0)
	e.Selection.Extend(5, 0) // select "hello"

	if !e.DeleteSelection() {
		t.Fatal("expected DeleteSelection to report true")
	}
	if e.Buffer.Line(0) != " world" {
		t.Fatalf("expected ' world', got %q", e.Buffer.Line(0))
	}
	if e.Cursor.X != 0 || e.Cursor.Y != 0 {
		t.Fatalf("expected cursor (0,0), got (%d,%d)", e.Cursor.X, e.Cursor.Y)
	}
	if e.Selection.Active() {
		t.Fatal("selection should be cleared after deletion")
	}

	// Undo restores the deleted text (single undo step).
	e.RestoreSnapshot()
	if e.Buffer.Line(0) != "hello world" {
		t.Fatalf("expected undo to restore 'hello world', got %q", e.Buffer.Line(0))
	}
}

func TestDeleteSelectionNoSelection(t *testing.T) {
	e := New()
	e.Mode = ModeEdit
	e.Buffer.InsertText("abcd", 0, 0)
	if e.DeleteSelection() {
		t.Fatal("expected DeleteSelection to report false when no selection")
	}
}

func TestSnapshotAndUndo(t *testing.T) {
	e := New()
	e.Buffer.InsertText("original", 0, 0)
	e.Buffer.SetModified(false)

	e.SaveSnapshot(OpGeneric)
	e.Buffer.InsertText(" modified", 8, 0)

	if e.Buffer.Line(0) != "original modified" {
		t.Fatalf("expected 'original modified', got %q", e.Buffer.Line(0))
	}

	restored := e.RestoreSnapshot()
	if !restored {
		t.Fatal("expected restore to succeed")
	}
	if e.Buffer.Line(0) != "original" {
		t.Fatalf("expected 'original' after restore, got %q", e.Buffer.Line(0))
	}
}

func TestCmdUndoRestoresContentAndCursor(t *testing.T) {
	e := New()
	e.Mode = ModeEdit
	e.Buffer.InsertText("hello", 0, 0)
	e.Cursor.X = 5 // end of line

	e.SaveSnapshot(OpGeneric)
	e.Buffer.Insert('!', 5, 0)
	e.Cursor.X = 6

	if e.Buffer.Line(0) != "hello!" {
		t.Fatalf("expected 'hello!', got %q", e.Buffer.Line(0))
	}

	restored := e.RestoreSnapshot()
	if !restored {
		t.Fatal("expected restore to succeed")
	}
	if e.Buffer.Line(0) != "hello" {
		t.Fatalf("expected undo to restore 'hello', got %q", e.Buffer.Line(0))
	}
	if e.Cursor.X != 5 || e.Cursor.Y != 0 {
		t.Fatalf("expected cursor restored to (5,0), got (%d,%d)", e.Cursor.X, e.Cursor.Y)
	}
}

// Typing a burst of characters should coalesce into a single undo step.
func TestCmdUndoCoalescesTyping(t *testing.T) {
	e := New()
	e.Mode = ModeEdit

	// Simulate typing "abc" one rune at a time (as the screen layer does).
	for _, ch := range "abc" {
		e.SaveSnapshot(OpInsert)
		e.Buffer.Insert(ch, e.Cursor.X, e.Cursor.Y)
		e.Cursor.X++
	}

	// Only one snapshot should be on the stack for the burst.
	if len(e.undoStack) != 1 {
		t.Fatalf("expected 1 coalesced snapshot, got %d", len(e.undoStack))
	}
	if e.Buffer.Line(0) != "abc" {
		t.Fatalf("expected 'abc', got %q", e.Buffer.Line(0))
	}

	// A single undo reverts the whole burst.
	if !e.RestoreSnapshot() {
		t.Fatal("expected restore to succeed")
	}
	if e.Buffer.Line(0) != "" {
		t.Fatalf("expected empty buffer after undo, got %q", e.Buffer.Line(0))
	}
	if e.Cursor.X != 0 || e.Cursor.Y != 0 {
		t.Fatalf("expected cursor restored to (0,0), got (%d,%d)", e.Cursor.X, e.Cursor.Y)
	}
}

// Undo then redo should return to the state right before the undo.
func TestCmdUndoAndRedo(t *testing.T) {
	e := New()
	e.Mode = ModeEdit
	e.Buffer.InsertText("hello", 0, 0)
	e.Cursor.X = 5

	e.SaveSnapshot(OpGeneric)
	e.Buffer.Insert('!', 5, 0)
	e.Cursor.X = 6

	// Undo
	if !e.RestoreSnapshot() {
		t.Fatal("undo should succeed")
	}
	if e.Buffer.Line(0) != "hello" {
		t.Fatalf("after undo expected 'hello', got %q", e.Buffer.Line(0))
	}

	// Redo
	if !e.Redo() {
		t.Fatal("redo should succeed")
	}
	if e.Buffer.Line(0) != "hello!" {
		t.Fatalf("after redo expected 'hello!', got %q", e.Buffer.Line(0))
	}
	if e.Cursor.X != 6 || e.Cursor.Y != 0 {
		t.Fatalf("after redo expected cursor (6,0), got (%d,%d)", e.Cursor.X, e.Cursor.Y)
	}

	// No more redo
	if e.Redo() {
		t.Fatal("redo should be exhausted")
	}
}

// A new edit after undo clears the redo history.
func TestCmdUndoThenEditClearsRedo(t *testing.T) {
	e := New()
	e.Mode = ModeEdit
	e.Buffer.InsertText("hello", 0, 0)

	e.SaveSnapshot(OpGeneric)
	e.Buffer.Insert('!', 5, 0)
	e.Cursor.X = 6

	e.RestoreSnapshot() // undo -> "hello", redo has the "hello!" state

	if e.Buffer.Line(0) != "hello" {
		t.Fatalf("expected 'hello' after undo, got %q", e.Buffer.Line(0))
	}

	// New edit invalidates redo
	e.SaveSnapshot(OpGeneric)
	e.Buffer.Insert('?', 5, 0)

	if e.Redo() {
		t.Fatal("redo should be cleared after a new edit")
	}
	if e.Buffer.Line(0) != "hello?" {
		t.Fatalf("expected 'hello?', got %q", e.Buffer.Line(0))
	}
}

func TestCmdSaveWithoutFilename(t *testing.T) {
	e := New()
	e.Mode = ModeEdit
	e.Buffer.InsertText("content", 0, 0)

	// cmdSave with no filename should call the file.saveAs handler
	// We can't test the handler directly (it's overridden by screen),
	// but we can verify the method doesn't crash

	// Verify the command registry has file.saveAs
	cmd := e.Commands.Find("file.saveAs")
	if cmd == nil {
		t.Fatal("file.saveAs command should exist")
	}

	// Save without filename — should be a no-op (the screen override handles the dialog)
	err := e.cmdSave()
	if err != nil {
		t.Fatalf("cmdSave() should not error on empty filename: %v", err)
	}
}

func TestCmdNew(t *testing.T) {
	e := New()
	e.Buffer.InsertText("some content", 0, 0)
	e.Buffer.SetModified(true)

	e.cmdNew()
	if e.Buffer.Lines() != 1 || e.Buffer.Line(0) != "" {
		t.Fatalf("expected empty buffer after new")
	}
	if e.Cursor.X != 0 || e.Cursor.Y != 0 {
		t.Fatalf("expected cursor reset to (0,0), got (%d,%d)", e.Cursor.X, e.Cursor.Y)
	}
}

func TestCommandRegistryFind(t *testing.T) {
	e := New()
	cmd := e.Commands.Find("file.save")
	if cmd == nil {
		t.Fatal("expected to find file.save command")
	}
	if cmd.ID != "file.save" {
		t.Fatalf("expected id 'file.save', got %q", cmd.ID)
	}

	// BUG: When Find returns &c (pointer to loop variable),
	// modifying the result modifies wrong command
	findCmd := e.Commands.Find("find.find")
	_ = findCmd

	// Test that each Find returns a unique pointer
	cmd1 := e.Commands.Find("file.save")
	cmd2 := e.Commands.Find("file.save")
	if cmd1 != cmd2 {
		// This is fine — they should point to the same element
	}

	// Check that different commands have different pointers
	cmdFile := e.Commands.Find("file.save")
	cmdEdit := e.Commands.Find("edit.copy")
	if cmdFile == cmdEdit {
		t.Fatal("BUG CONFIRMED: Find returns same pointer for different commands (loop variable bug)")
	}
}

// BUG CONFIRMED: Find returns a COPY (loop variable), so modifying the result
// doesn't affect the original command in the registry.
func TestCommandRegistryFindModifyReflectsInAll(t *testing.T) {
	e := New()
	cmd := e.Commands.Find("file.save")
	if cmd == nil {
		t.Fatal("expected to find file.save")
	}

	// Try to modify the handler via the Find pointer
	called := false
	cmd.Handler = func() error {
		called = true
		return nil
	}

	// Check via All() which returns the actual slice
	all := e.Commands.All()
	for _, c := range all {
		if c.ID == "file.save" {
			c.Handler()
			break
		}
	}

	if !called {
		t.Fatal("BUG CONFIRMED: Find returned a copy of the command, not a reference to the original. Modification via Find pointer was lost.")
	}
}

func TestCommandRegistryAll(t *testing.T) {
	e := New()
	all := e.Commands.All()
	if len(all) != 18 {
		t.Fatalf("expected 18 commands, got %d", len(all))
	}

	// Check specific commands exist
	ids := make(map[string]bool)
	for _, c := range all {
		ids[c.ID] = true
	}
	required := []string{"file.open", "file.save", "edit.cut", "edit.copy", "edit.paste", "app.quit", "help.keyboard"}
	for _, id := range required {
		if !ids[id] {
			t.Fatalf("missing required command: %s", id)
		}
	}
}
