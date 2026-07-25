package screen

import (
	"testing"

	"github.com/mattn/go-runewidth"
	"github.com/zsying/oe/internal/buffer"
	"github.com/zsying/oe/internal/editor"
)

// newTestScreen builds a minimal Screen sufficient for the wrap helpers,
// which only read sc.ed and sc.width.
func newTestScreen(width int) *Screen {
	return &Screen{ed: editor.New(), width: width}
}

func TestWrapSegments(t *testing.T) {
	// Empty line -> exactly one (zero-length) segment.
	seg := wrapSegments("", 10)
	if len(seg) != 1 || seg[0] != [2]int{0, 0} {
		t.Fatalf("empty line: got %v", seg)
	}

	// Short line fits in a single segment.
	seg = wrapSegments("hello", 10)
	if len(seg) != 1 || seg[0] != [2]int{0, 5} {
		t.Fatalf("short line: got %v", seg)
	}

	// A line longer than width must wrap into multiple segments, each
	// within the allowed width.
	line := "abcdefghij" // 10 chars
	seg = wrapSegments(line, 4)
	if len(seg) != 3 {
		t.Fatalf("expected 3 segments, got %d: %v", len(seg), seg)
	}
	for i, s := range seg {
		w := runewidth.StringWidth(string([]rune(line)[s[0]:s[1]]))
		if w > 4 {
			t.Fatalf("segment %d width %d exceeds 4: %v", i, w, s)
		}
	}

	// CJK characters are double-width and must wrap correctly.
	cjk := "中文测试换行显示效果验证" // 10 double-width runes
	seg = wrapSegments(cjk, 6)      // 3 CJK per row
	if len(seg) != 4 {
		t.Fatalf("expected 4 segments for cjk, got %d: %v", len(seg), seg)
	}
	for i, s := range seg {
		w := runewidth.StringWidth(string([]rune(cjk)[s[0]:s[1]]))
		if w > 6 {
			t.Fatalf("cjk segment %d width %d exceeds 6: %v", i, w, s)
		}
	}
}

func TestCursorVisualRow(t *testing.T) {
	sc := newTestScreen(10)
	sc.ed.ShowLineNum = false // disable gutter so contentWidth == width
	sc.ed.Buffer.SetLines([]string{"abcdefghij", "second line here"})
	sc.vm = sc.buildWrapMap()

	// First logical line is 10 wide; with width 10 and no gutter it fits on
	// one visual row (content width = 10). Cursor at rune 5 -> row 0.
	sc.ed.Cursor.X, sc.ed.Cursor.Y = 5, 0
	if got := sc.cursorVisualRow(); got != 0 {
		t.Fatalf("cursor (5,0): expected row 0, got %d", got)
	}

	// Force wrapping with a very narrow content width.
	sc.width = 4 // content width 4 (no gutter)
	sc.vm = sc.buildWrapMap()
	// "abcdefghij" wraps into rows of 4: [0,4),[4,8),[8,10); cursor 5 -> row 1
	sc.ed.Cursor.X, sc.ed.Cursor.Y = 5, 0
	if got := sc.cursorVisualRow(); got != 1 {
		t.Fatalf("cursor (5,0) narrow: expected row 1, got %d", got)
	}
	// Last rune (10) of the line -> last segment (row 2).
	sc.ed.Cursor.X, sc.ed.Cursor.Y = 10, 0
	if got := sc.cursorVisualRow(); got != 2 {
		t.Fatalf("cursor (10,0) narrow: expected row 2, got %d", got)
	}
}

func TestMoveWithSelectionShiftExtend(t *testing.T) {
	sc := newTestScreen(80)
	sc.ed.Buffer.InsertText("hello world", 0, 0)
	sc.ed.Cursor = &buffer.Cursor{} // (0,0)

	// First Shift+Right: begins selection at (0,0), extends to (1,0)
	sc.moveWithSelection(true, func() { sc.ed.Cursor.MoveRight(sc.ed.Buffer) })
	if !sc.ed.Selection.Active() {
		t.Fatal("selection should be active")
	}
	if sc.ed.Selection.StartX != 0 || sc.ed.Selection.StartY != 0 {
		t.Fatalf("anchor should stay at (0,0), got (%d,%d)", sc.ed.Selection.StartX, sc.ed.Selection.StartY)
	}
	if sc.ed.Selection.EndX != 1 || sc.ed.Selection.EndY != 0 {
		t.Fatalf("end should be (1,0), got (%d,%d)", sc.ed.Selection.EndX, sc.ed.Selection.EndY)
	}

	// Second Shift+Right: anchor MUST remain (0,0); end extends to (2,0)
	sc.moveWithSelection(true, func() { sc.ed.Cursor.MoveRight(sc.ed.Buffer) })
	if sc.ed.Selection.StartX != 0 || sc.ed.Selection.StartY != 0 {
		t.Fatalf("anchor must NOT move with the cursor; got (%d,%d)", sc.ed.Selection.StartX, sc.ed.Selection.StartY)
	}
	if sc.ed.Selection.EndX != 2 || sc.ed.Selection.EndY != 0 {
		t.Fatalf("end should be (2,0), got (%d,%d)", sc.ed.Selection.EndX, sc.ed.Selection.EndY)
	}

	// Non-shift move clears the selection
	sc.moveWithSelection(false, func() { sc.ed.Cursor.MoveRight(sc.ed.Buffer) })
	if sc.ed.Selection.Active() {
		t.Fatal("selection should be cleared after a non-shift move")
	}
}
