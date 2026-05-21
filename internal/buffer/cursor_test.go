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

	// Move to end and beyond
	for i := 0; i < 10; i++ {
		c.MoveRight(b)
	}
	if c.X != 5 || c.Y != 0 {
		t.Fatalf("expected (5,0) clamped, got (%d,%d)", c.X, c.Y)
	}

	// Move down to shorter line
	c.Y = 0
	c.X = 5 // past end of line 2 (world = 5)
	c.MoveDown(b)
	if c.Y != 1 {
		t.Fatalf("expected y=1, got %d", c.Y)
	}
	if c.X != 5 {
		t.Fatalf("X should be clamped; expected 5, got %d (world has length 5)", c.X)
	}
	// Actually world has 5 runes (w-o-r-l-d), so X=5 is at end, which is fine

	// Move down to shorter line (foo = 3)
	c.MoveDown(b)
	if c.Y != 2 {
		t.Fatalf("expected y=2, got %d", c.Y)
	}
	if c.X != 3 {
		t.Fatalf("X should clamp to 3 (foo length), got %d", c.X)
	}

	// Move left
	c.MoveLeft(b)
	if c.X != 2 {
		t.Fatalf("expected x=2, got %d", c.X)
	}

	// Move left to 0
	for i := 0; i < 5; i++ {
		c.MoveLeft(b)
	}
	if c.X != 0 {
		t.Fatalf("expected x=0, got %d", c.X)
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
