package buffer

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

// MoveLeft moves the cursor left by one character.
func (c *Cursor) MoveLeft(buf *Buffer) {
	if c.X > 0 {
		c.X--
	}
}

// MoveRight moves the cursor right by one character.
func (c *Cursor) MoveRight(buf *Buffer) {
	if c.X < buf.LineLen(c.Y) {
		c.X++
	}
}

// MoveToStartOfLine moves the cursor to the beginning of the current line.
func (c *Cursor) MoveToStartOfLine() {
	c.X = 0
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
