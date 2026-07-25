package screen

import (
	"github.com/mattn/go-runewidth"
)

// visualRow describes one wrapped visual row of a logical buffer line.
// A single logical line may produce several visual rows when it is wider
// than the available content area; soft-wrapping folds it instead of
// truncating or requiring horizontal scrolling.
type visualRow struct {
	line  int // logical line index in the buffer
	start int // rune index (inclusive) where this visual row begins
	end   int // rune index (exclusive) where this visual row ends
}

// gutterWidth returns the number of columns reserved for the line-number gutter.
func (sc *Screen) gutterWidth() int {
	if !sc.ed.ShowLineNum {
		return 0
	}
	n := sc.ed.Buffer.Lines()
	w := 0
	for n > 0 {
		n /= 10
		w++
	}
	if w < 2 {
		w = 2
	}
	return w + 2 // digits + trailing space + separator '│'
}

// contentWidth returns the number of columns available for text after the gutter.
func (sc *Screen) contentWidth() int {
	cw := sc.width - sc.gutterWidth()
	if cw < 1 {
		cw = 1
	}
	return cw
}

// wrapSegments splits a single logical line into visual-row segments, each of
// which fits within contentWidth display columns (measured via runewidth, so
// CJK/emoji are handled correctly). A line always yields at least one segment.
func wrapSegments(line string, contentWidth int) [][2]int {
	runes := []rune(line)
	if len(runes) == 0 {
		return [][2]int{{0, 0}}
	}
	var segs [][2]int
	start := 0
	col := 0
	for i, ch := range runes {
		w := runewidth.RuneWidth(ch)
		if w == 0 {
			// Zero-width rune (combining mark): attach to the current segment.
			continue
		}
		if col+w > contentWidth && i > start {
			segs = append(segs, [2]int{start, i})
			start = i
			col = w
		} else {
			col += w
		}
	}
	segs = append(segs, [2]int{start, len(runes)})
	return segs
}

// buildWrapMap returns the ordered list of visual rows for the whole buffer.
// Index 0 is the first visual row; the cursor, scroll offset and mouse mapping
// all operate in this visual-row space so wrapping is transparent everywhere.
func (sc *Screen) buildWrapMap() []visualRow {
	cw := sc.contentWidth()
	buf := sc.ed.Buffer
	vm := make([]visualRow, 0, buf.Lines())
	for y := 0; y < buf.Lines(); y++ {
		for _, seg := range wrapSegments(buf.Line(y), cw) {
			vm = append(vm, visualRow{line: y, start: seg[0], end: seg[1]})
		}
	}
	return vm
}

// cursorVisualRow returns the visual-row index that contains the cursor.
// It prefers a segment the cursor is strictly inside, then any segment that
// merely contains it (boundaries), then the first visual row of the line.
func (sc *Screen) cursorVisualRow() int {
	ed := sc.ed
	for i, vr := range sc.vm {
		if vr.line == ed.Cursor.Y && ed.Cursor.X >= vr.start && ed.Cursor.X < vr.end {
			return i
		}
	}
	for i, vr := range sc.vm {
		if vr.line == ed.Cursor.Y && ed.Cursor.X >= vr.start && ed.Cursor.X <= vr.end {
			return i
		}
	}
	for i, vr := range sc.vm {
		if vr.line == ed.Cursor.Y {
			return i
		}
	}
	return -1
}
