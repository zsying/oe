package buffer

import "testing"

func TestCursorBasic(t *testing.T) {
	b := New()
	b.InsertText("hello\nworld\nfoo", 0, 0)

	c := &Cursor{}

	// Move right on first line
	c.MoveRight(b)
	if c.X != 1 || c.Y != 0 {
		t.Fatalf("expected (1,0), got (%d,%d)", c.X, c.Y)
	}

	// Moving right past the end of a line wraps to the next line's start
	for c.Y == 0 {
		c.MoveRight(b)
	}
	if c.X != 0 || c.Y != 1 {
		t.Fatalf("expected wrap to (0,1), got (%d,%d)", c.X, c.Y)
	}

	// Moving right repeatedly reaches the last line's start, then its end
	for c.Y < 2 {
		c.MoveRight(b)
	}
	if c.X != 0 || c.Y != 2 {
		t.Fatalf("expected (0,2) at start of last line, got (%d,%d)", c.X, c.Y)
	}
	for c.X < 3 {
		c.MoveRight(b)
	}
	if c.X != 3 || c.Y != 2 {
		t.Fatalf("expected (3,2) at end of last line, got (%d,%d)", c.X, c.Y)
	}

	// Move down clamps X to the shorter line length
	c.Y = 0
	c.X = 5 // end of "hello"
	c.MoveDown(b)
	if c.Y != 1 {
		t.Fatalf("expected y=1, got %d", c.Y)
	}
	if c.X != 5 {
		t.Fatalf("X should be clamped; expected 5, got %d (world has length 5)", c.X)
	}

	c.MoveDown(b) // to "foo" (len 3)
	if c.Y != 2 {
		t.Fatalf("expected y=2, got %d", c.Y)
	}
	if c.X != 3 {
		t.Fatalf("X should clamp to 3 (foo length), got %d", c.X)
	}

	// Move left stays within the line
	c.MoveLeft(b)
	if c.X != 2 {
		t.Fatalf("expected x=2, got %d", c.X)
	}
}

func TestCursorCrossLine(t *testing.T) {
	b := New()
	b.InsertText("hello\nworld\nfoo", 0, 0)

	// Left at the start of line 1 wraps to the end of line 0 ("hello" len 5)
	c := &Cursor{Y: 1, X: 0}
	c.MoveLeft(b)
	if c.X != 5 || c.Y != 0 {
		t.Fatalf("expected (5,0), got (%d,%d)", c.X, c.Y)
	}

	// Right at the end of line 0 wraps to the start of line 1
	c = &Cursor{Y: 0, X: 5}
	c.MoveRight(b)
	if c.X != 0 || c.Y != 1 {
		t.Fatalf("expected (0,1), got (%d,%d)", c.X, c.Y)
	}

	// Left at the very top stays put
	c = &Cursor{Y: 0, X: 0}
	c.MoveLeft(b)
	if c.X != 0 || c.Y != 0 {
		t.Fatalf("expected (0,0), got (%d,%d)", c.X, c.Y)
	}

	// Right at the very bottom stays put
	c = &Cursor{Y: 2, X: 3}
	c.MoveRight(b)
	if c.X != 3 || c.Y != 2 {
		t.Fatalf("expected (3,2), got (%d,%d)", c.X, c.Y)
	}
}

func TestCursorMoveToLineBounds(t *testing.T) {
	b := New()
	b.InsertText("abc\nde\nfghij", 0, 0)

	c := &Cursor{}

	// Move to start
	c.MoveToStartOfLine()
	if c.X != 0 {
		t.Fatalf("expected x=0, got %d", c.X)
	}

	// Move to end
	c.MoveToEndOfLine(b)
	if c.X != 3 {
		t.Fatalf("expected x=3 (abc), got %d", c.X)
	}

	// Move down (now at end of "abc")
	c.MoveDown(b) // to "de"
	if c.X != 2 {
		t.Fatalf("X should clamp to 2 (de length), got %d", c.X)
	}

	c.MoveDown(b) // to "fghij"
	if c.X != 2 {
		t.Fatalf("X should stay at 2, got %d (fghij has 5, so 2 is valid)", c.X)
	}
}

func TestCursorPageUpDown(t *testing.T) {
	b := New()
	// Create 20 lines
	for i := 0; i < 20; i++ {
		if i == 0 {
			b.InsertText("line0", 0, 0)
		} else {
			b.NewLine(5, i-1) // split, cursor goes to next line
			b.InsertText("line", 0, i) // will be wrong since we're inserting at line i with wrong content
		}
	}

	// Simpler: build multi-line buffer
	b2 := New()
	lines := []string{"line0", "line1", "line2", "line3", "line4", "line5", "line6", "line7", "line8", "line9"}
	b2.lines = lines

	c := &Cursor{Y: 5, X: 3}

	c.MovePageUp(b2, 3) // up 3 lines
	if c.Y != 2 {
		t.Fatalf("expected y=2, got %d", c.Y)
	}

	c.MovePageDown(b2, 3) // down 3 lines
	if c.Y != 5 {
		t.Fatalf("expected y=5, got %d", c.Y)
	}

	// Don't go past start
	c.Y = 1
	c.MovePageUp(b2, 5)
	if c.Y != 0 {
		t.Fatalf("expected y=0 (clamped), got %d", c.Y)
	}

	// Don't go past end
	c.Y = 8
	c.MovePageDown(b2, 5)
	if c.Y != 9 {
		t.Fatalf("expected y=9 (clamped to last line), got %d", c.Y)
	}
}

func TestCursorBoundaries(t *testing.T) {
	b := New() // single empty line
	c := &Cursor{}

	// All movements on empty buffer should stay at (0,0)
	c.MoveLeft(b)
	c.MoveRight(b)
	c.MoveUp(b)
	c.MoveDown(b)
	if c.X != 0 || c.Y != 0 {
		t.Fatalf("expected (0,0) on empty buffer, got (%d,%d)", c.X, c.Y)
	}
}

func TestCursorWordMove(t *testing.T) {
	b := New()
	b.InsertText("hello world\nfoo bar baz", 0, 0)

	// Ctrl+Right from start of "hello" jumps to start of "world" (x=6)
	c := &Cursor{Y: 0, X: 0}
	c.MoveWordRight(b)
	if c.X != 6 || c.Y != 0 {
		t.Fatalf("expected (6,0), got (%d,%d)", c.X, c.Y)
	}

	// Another Ctrl+Right: end of "world" (x=11)
	c.MoveWordRight(b)
	if c.X != 11 || c.Y != 0 {
		t.Fatalf("expected (11,0), got (%d,%d)", c.X, c.Y)
	}

	// Another Ctrl+Right crosses into next line start of "foo" (0,1)
	c.MoveWordRight(b)
	if c.X != 0 || c.Y != 1 {
		t.Fatalf("expected (0,1), got (%d,%d)", c.X, c.Y)
	}

	// Ctrl+Left from middle of "bar" (x=7 within "foo bar baz") -> start of "bar" (x=4)
	c = &Cursor{Y: 1, X: 7}
	c.MoveWordLeft(b)
	if c.X != 4 || c.Y != 1 {
		t.Fatalf("expected (4,1), got (%d,%d)", c.X, c.Y)
	}

	// Ctrl+Left from start of "foo" wraps to end of previous line (11,0)
	c = &Cursor{Y: 1, X: 0}
	c.MoveWordLeft(b)
	if c.X != 11 || c.Y != 0 {
		t.Fatalf("expected (11,0), got (%d,%d)", c.X, c.Y)
	}

	// Ctrl+Right at very bottom stays put
	c = &Cursor{Y: 1, X: 11}
	c.MoveWordRight(b)
	if c.X != 11 || c.Y != 1 {
		t.Fatalf("expected (11,1), got (%d,%d)", c.X, c.Y)
	}

	// Ctrl+Left at very top stays put
	c = &Cursor{Y: 0, X: 0}
	c.MoveWordLeft(b)
	if c.X != 0 || c.Y != 0 {
		t.Fatalf("expected (0,0), got (%d,%d)", c.X, c.Y)
	}
}

func TestCursorWordMoveSeparators(t *testing.T) {
	b := New()
	b.InsertText("foo   bar_baz.qux", 0, 0)

	// From start, skip "foo" then the three spaces -> start of "bar_baz" (x=6)
	c := &Cursor{Y: 0, X: 0}
	c.MoveWordRight(b)
	if c.X != 6 || c.Y != 0 {
		t.Fatalf("expected (6,0), got (%d,%d)", c.X, c.Y)
	}

	// "bar_baz" is one word (underscore counts); skip it then the "." ->
	// cursor lands at the start of the next word "qux" (x=14)
	c.MoveWordRight(b)
	if c.X != 14 || c.Y != 0 {
		t.Fatalf("expected (14,0), got (%d,%d)", c.X, c.Y)
	}

	// Skip "qux" -> end of line (x=17)
	c.MoveWordRight(b)
	if c.X != 17 || c.Y != 0 {
		t.Fatalf("expected (17,0), got (%d,%d)", c.X, c.Y)
	}
}
