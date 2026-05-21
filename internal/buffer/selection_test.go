package buffer

import "testing"

func TestSelectionBasic(t *testing.T) {
	b := New()
	b.InsertText("hello world", 0, 0)

	s := &Selection{}
	if s.Active() {
		t.Fatal("new selection should not be active")
	}

	s.Begin(0, 0)
	if !s.Active() {
		t.Fatal("selection should be active after Begin")
	}

	s.Extend(5, 0)
	text := s.Text(b)
	if text != "hello" {
		t.Fatalf("expected 'hello', got %q", text)
	}

	s.Clear()
	if s.Active() {
		t.Fatal("selection should not be active after Clear")
	}
}

func TestSelectionNormalized(t *testing.T) {
	s := &Selection{}
	s.Begin(5, 1)
	s.Extend(2, 0)

	startX, startY, endX, endY := s.Normalized()
	if startX != 2 || startY != 0 || endX != 5 || endY != 1 {
		t.Fatalf("expected (2,0)→(5,1), got (%d,%d)→(%d,%d)",
			startX, startY, endX, endY)
	}
}

func TestSelectionMultiLine(t *testing.T) {
	b := New()
	b.lines = []string{"abc", "def", "ghi"}

	s := &Selection{}
	s.Begin(1, 0) // 'b' in "abc"
	s.Extend(2, 2) // up to 'h' in "ghi"

	text := s.Text(b)
	// Expected: "bc" + "\n" + "def" + "\n" + "gh"
	expected := "bc\ndef\ngh"
	if text != expected {
		t.Fatalf("expected %q, got %q", expected, text)
	}
}

func TestSelectionDeleteSingleLine(t *testing.T) {
	b := New()
	b.InsertText("hello world", 0, 0)

	s := &Selection{}
	s.Begin(0, 0)
	s.Extend(5, 0) // select "hello"

	s.Delete(b)
	if b.Line(0) != " world" {
		t.Fatalf("expected ' world', got %q", b.Line(0))
	}
	if !b.Modified() {
		t.Fatal("buffer should be modified after delete")
	}
	if s.Active() {
		t.Fatal("selection should be cleared after delete")
	}
}



func TestSelectionContains(t *testing.T) {
	b := New()
	b.lines = []string{"abcdef", "ghijkl"}

	s := &Selection{}
	s.Begin(1, 0) // 'b'
	s.Extend(3, 1) // up to 'j' in "ghijkl"

	// Test points inside
	if !s.Contains(2, 0) {
		t.Fatal("(2,0) should be in selection")
	}
	if !s.Contains(0, 1) {
		t.Fatal("(0,1) should be in selection")
	}
	if !s.Contains(2, 1) {
		t.Fatal("(2,1) should be in selection")
	}

	// Test points outside
	if s.Contains(0, 0) {
		t.Fatal("(0,0) should NOT be in selection (before startX on startY)")
	}
	if s.Contains(3, 1) {
		t.Fatal("(3,1) should NOT be in selection (endX is exclusive)")
	}
	if s.Contains(0, 2) {
		t.Fatal("(0,2) should NOT be in selection (after endY)")
	}
}

// TestSelectionDeleteMultiLineFixed corrects the wrong test above
func TestSelectionDeleteMultiLineFixed(t *testing.T) {
	b := New()
	b.lines = []string{"abc", "def", "ghi"}

	s := &Selection{}
	s.Begin(1, 0) // 'b' in "abc"
	s.Extend(2, 2) // up to 'g' in "ghi" (exclusive of position 2 on line 2)

	s.Delete(b)
	// Remaining: "a" (line 0 start) + "hi" (line 2 from position 2 to end)
	// Wait, that's not right either. Let me trace through:
	// Normalized: start=(1,0), end=(2,2)
	// topLine = "abc"[:1] = "a"
	// bottomLine = "ghi"[2:] = "i"  (wait, position 2 on "ghi" is 'i', so [2:] = "i"? No...
	// Actually "ghi" has runes: g(0), h(1), i(2). So [2:] = "i" wait no: 
	// If endX=2, bottomLine[:endX] would be "gh" and bottomLine[endX:] would be "i"
	// Wait let me re-read the code:
	// bottomLine := []rune(buf.Line(endY))  -- line "ghi" → []rune{'g','h','i'}
	// buf.lines[startY] = string(append(topLine[:startX], bottomLine[endX:]...))
	// = string(append("a"[:1], "ghi"[2:]...)) -- but "ghi"[2:] as runes is "i"
	// = "a" + "i" = "ai"
	// Then lines are [startY+1:endY+1] which is lines[1:3] = ["def", "ghi"]
	// But wait, startY=0, endY=2, so:
	// buf.lines = append(buf.lines[:0+1], buf.lines[2+1:]...)
	// = append(["ai"], buf.lines[3:]...)
	// = ["ai"]
	// Hmm, that removes too much! lines[3:] on ["abc", "def", "ghi"] is empty.
	// So result is just ["ai"]

	b2 := New()
	b2.lines = []string{"abc", "def", "ghi"}
	s2 := &Selection{}
	s2.Begin(1, 0)
	s2.Extend(2, 2)
	s2.Delete(b2)

	if b2.Lines() != 1 {
		t.Fatalf("expected 1 line, got %d lines: %v", b2.Lines(), b2.lines)
	}
	if b2.Line(0) != "ai" {
		t.Fatalf("expected 'ai', got %q", b2.Line(0))
	}
}
