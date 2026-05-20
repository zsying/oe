package buffer

import "strings"

// Selection represents a text selection range.
type Selection struct {
	StartX, StartY int
	EndX, EndY     int
	active         bool
}

// Begin starts a new selection at (x, y).
func (s *Selection) Begin(x, y int) {
	s.StartX, s.StartY = x, y
	s.EndX, s.EndY = x, y
	s.active = true
}

// Extend extends the selection to (x, y).
func (s *Selection) Extend(x, y int) {
	s.EndX, s.EndY = x, y
}

// Active returns whether a selection is active.
func (s *Selection) Active() bool { return s.active }

// Clear clears the selection.
func (s *Selection) Clear() {
	s.active = false
	s.StartX, s.StartY = 0, 0
	s.EndX, s.EndY = 0, 0
}

// Normalized returns the selection boundaries in normalized order (start ≤ end).
func (s *Selection) Normalized() (startX, startY, endX, endY int) {
	startX, startY = s.StartX, s.StartY
	endX, endY = s.EndX, s.EndY
	if startY > endY || (startY == endY && startX > endX) {
		startX, startY, endX, endY = endX, endY, startX, startY
	}
	return
}

// Text returns the selected text.
func (s *Selection) Text(buf *Buffer) string {
	if !s.active {
		return ""
	}
	startX, startY, endX, endY := s.Normalized()
	if startY == endY {
		line := []rune(buf.Line(startY))
		if startX > len(line) {
			startX = len(line)
		}
		if endX > len(line) {
			endX = len(line)
		}
		return string(line[startX:endX])
	}
	var result []string
	firstLine := []rune(buf.Line(startY))
	if startX > len(firstLine) {
		startX = len(firstLine)
	}
	result = append(result, string(firstLine[startX:]))
	for y := startY + 1; y < endY; y++ {
		result = append(result, buf.Line(y))
	}
	lastLine := []rune(buf.Line(endY))
	if endX > len(lastLine) {
		endX = len(lastLine)
	}
	result = append(result, string(lastLine[:endX]))
	return strings.Join(result, "\n")
}

// Delete deletes the selected text from the buffer and clears the selection.
func (s *Selection) Delete(buf *Buffer) {
	if !s.active {
		return
	}
	startX, startY, endX, endY := s.Normalized()
	if startY == endY {
		line := []rune(buf.Line(startY))
		if endX > len(line) {
			endX = len(line)
		}
		buf.lines[startY] = string(append(line[:startX], line[endX:]...))
	} else {
		topLine := []rune(buf.Line(startY))
		bottomLine := []rune(buf.Line(endY))
		if startX > len(topLine) {
			startX = len(topLine)
		}
		if endX > len(bottomLine) {
			endX = len(bottomLine)
		}
		buf.lines[startY] = string(append(topLine[:startX], bottomLine[endX:]...))
		buf.lines = append(buf.lines[:startY+1], buf.lines[endY+1:]...)
	}
	buf.modified = true
	s.Clear()
}

// Contains returns true if position (x, y) is within the selection.
func (s *Selection) Contains(x, y int) bool {
	if !s.active {
		return false
	}
	startX, startY, endX, endY := s.Normalized()
	if y < startY || y > endY {
		return false
	}
	if y == startY && x < startX {
		return false
	}
	if y == endY && x >= endX {
		return false
	}
	return true
}
