package buffer

import "unicode"

// Cursor represents a position in the buffer.
type Cursor struct {
	X, Y int
}

// MoveUp moves the cursor up one line.
func (c *Cursor) MoveUp(buf *Buffer) {
	if c.Y > 0 {
		c.Y--
	}
	c.ClampX(buf)
}

// MoveDown moves the cursor down one line.
func (c *Cursor) MoveDown(buf *Buffer) {
	if c.Y < buf.Lines()-1 {
		c.Y++
	}
	c.ClampX(buf)
}

// MoveLeft moves the cursor left by one character. At the start of a line it
// wraps to the end of the previous line; at the very top it stays put.
func (c *Cursor) MoveLeft(buf *Buffer) {
	if c.X > 0 {
		c.X--
		return
	}
	if c.Y > 0 {
		c.Y--
		c.X = buf.LineLen(c.Y)
	}
}

// MoveRight moves the cursor right by one character. At the end of a line it
// wraps to the start of the next line; at the very bottom it stays put.
func (c *Cursor) MoveRight(buf *Buffer) {
	if c.X < buf.LineLen(c.Y) {
		c.X++
		return
	}
	if c.Y < buf.Lines()-1 {
		c.Y++
		c.X = 0
	}
}

// MoveToStartOfLine moves the cursor to the beginning of the current line.
func (c *Cursor) MoveToStartOfLine() {
	c.X = 0
}

// isWordRune reports whether r is part of a "word" (letters, digits, underscore).
func isWordRune(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// MoveWordLeft moves the cursor to the start of the previous word, crossing
// line boundaries. If the cursor is already in the middle of a word it jumps
// to that word's start; if at a word's start it jumps to the previous word.
// At the very top of the buffer it stays put.
func (c *Cursor) MoveWordLeft(buf *Buffer) {
	for {
		if c.X == 0 {
			if c.Y == 0 {
				return
			}
			c.Y--
			c.X = buf.LineLen(c.Y)
			if buf.LineLen(c.Y) != 0 {
				return
			}
			// Empty previous line: keep going up.
			continue
		}
		runes := []rune(buf.Line(c.Y))
		// Skip non-word characters to the left.
		for c.X > 0 && !isWordRune(runes[c.X-1]) {
			c.X--
		}
		// Skip word characters to the left.
		for c.X > 0 && isWordRune(runes[c.X-1]) {
			c.X--
		}
		return
	}
}

// MoveWordRight moves the cursor to the start of the next word, crossing line
// boundaries. At the very bottom of the buffer it stays put.
func (c *Cursor) MoveWordRight(buf *Buffer) {
	for {
		lineLen := buf.LineLen(c.Y)
		if c.X >= lineLen {
			if c.Y >= buf.Lines()-1 {
				return
			}
			c.Y++
			c.X = 0
			if buf.LineLen(c.Y) != 0 {
				return
			}
			// Empty next line: keep going down.
			continue
		}
		runes := []rune(buf.Line(c.Y))
		// Skip word characters to the right (current word).
		for c.X < lineLen && isWordRune(runes[c.X]) {
			c.X++
		}
		// Skip non-word characters to the right (separator).
		for c.X < lineLen && !isWordRune(runes[c.X]) {
			c.X++
		}
		return
	}
}

// MoveToEndOfLine moves the cursor to the end of the current line.
func (c *Cursor) MoveToEndOfLine(buf *Buffer) {
	c.X = buf.LineLen(c.Y)
}

// MovePageUp moves the cursor up by viewHeight lines.
func (c *Cursor) MovePageUp(buf *Buffer, viewHeight int) {
	c.Y -= viewHeight
	if c.Y < 0 {
		c.Y = 0
	}
	c.ClampX(buf)
}

// MovePageDown moves the cursor down by viewHeight lines.
func (c *Cursor) MovePageDown(buf *Buffer, viewHeight int) {
	c.Y += viewHeight
	if c.Y >= buf.Lines() {
		c.Y = buf.Lines() - 1
	}
	c.ClampX(buf)
}

// ClampX ensures the cursor X position is valid for the current line.
func (c *Cursor) ClampX(buf *Buffer) {
	if c.X > buf.LineLen(c.Y) {
		c.X = buf.LineLen(c.Y)
	}
}
