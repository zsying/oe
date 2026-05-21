package buffer

import (
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	b := New()
	if b.Lines() != 1 {
		t.Fatalf("expected 1 line, got %d", b.Lines())
	}
	if b.Line(0) != "" {
		t.Fatalf("expected empty line, got %q", b.Line(0))
	}
	if b.Modified() {
		t.Fatal("new buffer should not be modified")
	}
}

func TestInsert(t *testing.T) {
	b := New()
	b.Insert('a', 0, 0)
	if b.Line(0) != "a" {
		t.Fatalf("expected 'a', got %q", b.Line(0))
	}
	if !b.Modified() {
		t.Fatal("buffer should be modified after insert")
	}

	// Insert at end
	b.Insert('b', 1, 0)
	if b.Line(0) != "ab" {
		t.Fatalf("expected 'ab', got %q", b.Line(0))
	}

	// Insert in middle
	b.Insert('X', 1, 0)
	if b.Line(0) != "aXb" {
		t.Fatalf("expected 'aXb', got %q", b.Line(0))
	}

	// Insert at beginning
	b.Insert('0', 0, 0)
	if b.Line(0) != "0aXb" {
		t.Fatalf("expected '0aXb', got %q", b.Line(0))
	}
}

func TestInsertOutOfBounds(t *testing.T) {
	b := New()
	// y out of bounds — should be no-op
	b.Insert('x', 0, 999)
	if b.Line(0) != "" {
		t.Fatal("insert at out-of-bounds y should be no-op")
	}

	// x beyond line — clamps to end
	b.Insert('a', 100, 0)
	if b.Line(0) != "a" {
		t.Fatalf("expected 'a' (clamped), got %q", b.Line(0))
	}

	// negative x — clamps to 0
	b2 := New()
	b2.Insert('x', -1, 0)
	if b2.Line(0) != "x" {
		t.Fatalf("expected 'x' (clamped from -1), got %q", b2.Line(0))
	}
}

func TestDeleteBackward(t *testing.T) {
	b := New()
	b.Insert('a', 0, 0)
	b.Insert('b', 1, 0)
	b.Insert('c', 2, 0) // "abc"

	// Delete middle
	x, y := b.DeleteBackward(2, 0) // delete 'b'
	if b.Line(0) != "ac" {
		t.Fatalf("expected 'ac', got %q", b.Line(0))
	}
	if x != 1 || y != 0 {
		t.Fatalf("expected (1,0), got (%d,%d)", x, y)
	}

	// Delete at start of line — should join with previous
	b2 := New()
	b2.Insert('a', 0, 0)
	b2.NewLine(1, 0) // split into ["a", ""] move cursor to (0,1)
	b2.Insert('b', 0, 1) // "b" on second line
	// Now lines are ["a", "b"]
	x, y = b2.DeleteBackward(0, 1) // delete at start of "b" — joins with "a"
	if b2.Lines() != 1 {
		t.Fatalf("expected 1 line after join, got %d", b2.Lines())
	}
	if b2.Line(0) != "ab" {
		t.Fatalf("expected 'ab', got %q", b2.Line(0))
	}
	if x != 1 || y != 0 {
		t.Fatalf("expected (1,0) (cursor at join point), got (%d,%d)", x, y)
	}

	// Delete at start of first line — no-op
	beforeLines := b2.Lines()
	x, y = b2.DeleteBackward(0, 0)
	if b2.Lines() != beforeLines {
		t.Fatal("delete at buffer start should not change line count")
	}
	if x != 0 || y != 0 {
		t.Fatalf("expected (0,0), got (%d,%d)", x, y)
	}
}

func TestDeleteForward(t *testing.T) {
	b := New()
	b.Insert('a', 0, 0)
	b.Insert('b', 1, 0)
	b.Insert('c', 2, 0) // "abc"

	b.DeleteForward(1, 0) // delete 'b'
	if b.Line(0) != "ac" {
		t.Fatalf("expected 'ac', got %q", b.Line(0))
	}

	// Delete at end of line — join with next line
	b2 := New()
	b2.Insert('h', 0, 0)
	b2.NewLine(1, 0) // ["h", ""]
	b2.Insert('i', 0, 1) // ["h", "i"]
	b2.DeleteForward(1, 0) // at end of "h" — join with "i"
	if b2.Lines() != 1 {
		t.Fatalf("expected 1 line, got %d", b2.Lines())
	}
	if b2.Line(0) != "hi" {
		t.Fatalf("expected 'hi', got %q", b2.Line(0))
	}

	// Delete at end of last line — no-op
	beforeLines := b.Lines()
	b.DeleteForward(2, 0) // at end of "ac"
	if b.Lines() != beforeLines {
		t.Fatal("delete at end of last line should be no-op")
	}
}

func TestNewLine(t *testing.T) {
	b := New()
	b.Insert('a', 0, 0)
	b.Insert('b', 1, 0)
	b.Insert('c', 2, 0) // "abc"

	// Split in middle
	x, y := b.NewLine(1, 0) // split "abc" -> "a" + "bc"
	if b.Line(0) != "a" {
		t.Fatalf("expected 'a', got %q", b.Line(0))
	}
	if b.Line(1) != "bc" {
		t.Fatalf("expected 'bc', got %q", b.Line(1))
	}
	if x != 0 || y != 1 {
		t.Fatalf("expected (0,1), got (%d,%d)", x, y)
	}

	// Split at end — creates empty line
	x, y = b.NewLine(2, 1) // at end of "bc" -> "bc" + ""
	if b.Line(2) != "" {
		t.Fatalf("expected empty line, got %q", b.Line(2))
	}

	// Negative x — clamps to 0
	b2 := New()
	b2.Insert('x', 0, 0)
	b2.NewLine(-1, 0)
	if b2.Line(0) != "" || b2.Line(1) != "x" {
		t.Fatalf("expected ['', 'x'], got [%q, %q]", b2.Line(0), b2.Line(1))
	}
}

func TestInsertText(t *testing.T) {
	b := New()
	b.InsertText("hello", 0, 0)
	if b.Line(0) != "hello" {
		t.Fatalf("expected 'hello', got %q", b.Line(0))
	}

	// Multi-line paste in middle
	b2 := New()
	b2.InsertText("ab", 0, 0) // "ab"
	x, y := b2.InsertText("1\n2\n3", 1, 0) // insert between a and b
	if b2.Line(0) != "a1b" {
		t.Fatalf("expected 'a1b', got %q", b2.Line(0))
	}
	if b2.Line(1) != "2" {
		t.Fatalf("expected '2', got %q", b2.Line(1))
	}
	if b2.Line(2) != "3" {
		t.Fatalf("expected '3', got %q", b2.Line(2))
	}
	if x != 1 || y != 2 {
		t.Fatalf("expected (1,2), got (%d,%d)", x, y)
	}
}

func TestInsertChinese(t *testing.T) {
	b := New()
	b.InsertText("你好世界", 0, 0)
	if b.Line(0) != "你好世界" {
		t.Fatalf("expected '你好世界', got %q", b.Line(0))
	}
	if b.LineLen(0) != 4 {
		t.Fatalf("expected 4 runes, got %d", b.LineLen(0))
	}

	// Insert between Chinese chars
	b.InsertText("!", 2, 0) // insert between "你好" and "世界"
	if b.Line(0) != "你好!世界" {
		t.Fatalf("expected '你好!世界', got %q", b.Line(0))
	}
	if b.LineLen(0) != 5 {
		t.Fatalf("expected 5 runes, got %d", b.LineLen(0))
	}
}

func TestLoadAndSave(t *testing.T) {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "editor-test-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	content := "line1\nline2\nline3"
	if err := os.WriteFile(tmpPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	b := New()
	if err := b.Load(tmpPath); err != nil {
		t.Fatal(err)
	}

	if b.Lines() != 3 {
		t.Fatalf("expected 3 lines, got %d", b.Lines())
	}
	if b.Line(0) != "line1" || b.Line(1) != "line2" || b.Line(2) != "line3" {
		t.Fatalf("unexpected content: [%q, %q, %q]", b.Line(0), b.Line(1), b.Line(2))
	}
	if b.Filename() != tmpPath {
		t.Fatalf("filename not set correctly")
	}
	if b.Modified() {
		t.Fatal("buffer should not be modified after load")
	}

	// Modify and save
	b.Insert('X', 5, 0)
	b.Insert('Y', 5, 1)
	if err := b.Save(); err != nil {
		t.Fatal(err)
	}

	saved, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatal(err)
	}
	expected := "line1X\nline2Y\nline3"
	if string(saved) != expected {
		t.Fatalf("expected %q, got %q", expected, string(saved))
	}
	if b.Modified() {
		t.Fatal("buffer should not be modified after save")
	}
}

func TestLineLen(t *testing.T) {
	b := New()
	b.InsertText("你好世界", 0, 0)
	// Each Chinese character is 1 rune
	if b.LineLen(0) != 4 {
		t.Fatalf("expected 4 (runes), got %d", b.LineLen(0))
	}

	// Out of bounds
	if b.LineLen(-1) != 0 {
		t.Fatal("expected 0 for negative index")
	}
	if b.LineLen(999) != 0 {
		t.Fatal("expected 0 for out-of-bounds index")
	}
}

func TestSnapshot(t *testing.T) {
	b := New()
	b.InsertText("original", 0, 0)

	lines := b.LinesSlice()
	if len(lines) != 1 || lines[0] != "original" {
		t.Fatalf("snapshot mismatch: %v", lines)
	}

	b.Insert('!', 8, 0)

	b.SetLines(lines)
	if b.Line(0) != "original" {
		t.Fatalf("restore failed: got %q", b.Line(0))
	}
}

func TestCRLFHandling(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "editor-test-crlf-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Write CRLF content
	if err := os.WriteFile(tmpPath, []byte("line1\r\nline2\r\nline3"), 0644); err != nil {
		t.Fatal(err)
	}

	b := New()
	if err := b.Load(tmpPath); err != nil {
		t.Fatal(err)
	}
	if b.Lines() != 3 {
		t.Fatalf("expected 3 lines, got %d (CRLF not handled)", b.Lines())
	}
	if b.Line(0) != "line1" || b.Line(1) != "line2" || b.Line(2) != "line3" {
		t.Fatalf("CRLF not stripped correctly: %v", []string{b.Line(0), b.Line(1), b.Line(2)})
	}
}

func TestInsertAndDeleteChineseCharacters(t *testing.T) {
	b := New()
	b.InsertText("Hello世界", 0, 0)
	if b.LineLen(0) != 7 {
		t.Fatalf("expected 7 runes (H,e,l,l,o,世,界), got %d", b.LineLen(0))
	}

	// DeleteForward at position 5 (before 世)
	b.SetModified(false)
	b.DeleteForward(5, 0)
	// Should delete 世
	if b.Line(0) != "Hello界" {
		t.Fatalf("expected 'Hello界', got %q", b.Line(0))
	}

	// DeleteBackward at position 6 (after 界)
	b2 := New()
	b2.InsertText("Hello世界", 0, 0)
	x, y := b2.DeleteBackward(7, 0)
	// Should delete 界
	if b2.Line(0) != "Hello世" {
		t.Fatalf("expected 'Hello世', got %q", b2.Line(0))
	}
	if x != 6 || y != 0 {
		t.Fatalf("expected cursor (6,0), got (%d,%d)", x, y)
	}
}
