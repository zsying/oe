package buffer

import (
	"os"
	"strings"
)

// Buffer represents a text buffer using a line array.
type Buffer struct {
	lines    []string
	filename string
	fileType string
	modified bool
	tabWidth int
}

// New creates a new empty buffer.
func New() *Buffer {
	return &Buffer{
		lines:    []string{""},
		tabWidth: 4,
	}
}

// Lines returns the number of lines.
func (b *Buffer) Lines() int { return len(b.lines) }

// Line returns the content of line n (0-indexed).
func (b *Buffer) Line(n int) string {
	if n < 0 || n >= len(b.lines) {
		return ""
	}
	return b.lines[n]
}

// Filename returns the file path.
func (b *Buffer) Filename() string { return b.filename }

// SetFilename sets the file path.
func (b *Buffer) SetFilename(s string) { b.filename = s }

// Modified returns whether the buffer has unsaved changes.
func (b *Buffer) Modified() bool { return b.modified }

// SetModified sets the modified flag explicitly.
func (b *Buffer) SetModified(m bool) { b.modified = m }

// TabWidth returns the tab width.
func (b *Buffer) TabWidth() int { return b.tabWidth }

// FileType returns the file extension.
func (b *Buffer) FileType() string { return b.fileType }

// SetFileType sets the file type.
func (b *Buffer) SetFileType(s string) { b.fileType = s }

// Insert inserts a rune at (x, y) in the buffer.
func (b *Buffer) Insert(ch rune, x, y int) {
	if y < 0 || y >= len(b.lines) {
		return
	}
	line := []rune(b.lines[y])
	if x < 0 {
		x = 0
	}
	if x > len(line) {
		x = len(line)
	}
	newLine := string(append(line[:x], append([]rune{ch}, line[x:]...)...))
	b.lines[y] = newLine
	b.modified = true
}

// DeleteBackward deletes the character before (x, y) and returns the new cursor position.
func (b *Buffer) DeleteBackward(x, y int) (newX, newY int) {
	if y < 0 || y >= len(b.lines) {
		return x, y
	}
	if x > 0 {
		line := []rune(b.lines[y])
		b.lines[y] = string(append(line[:x-1], line[x:]...))
		b.modified = true
		return x - 1, y
	}
	if y > 0 {
		prevLen := len([]rune(b.lines[y-1]))
		b.lines[y-1] += b.lines[y]
		b.lines = append(b.lines[:y], b.lines[y+1:]...)
		b.modified = true
		return prevLen, y - 1
	}
	return 0, 0
}

// DeleteForward deletes the character at or after (x, y).
func (b *Buffer) DeleteForward(x, y int) {
	if y < 0 || y >= len(b.lines) {
		return
	}
	line := []rune(b.lines[y])
	if x < len(line) {
		b.lines[y] = string(append(line[:x], line[x+1:]...))
		b.modified = true
		return
	}
	if y < len(b.lines)-1 {
		b.lines[y] += b.lines[y+1]
		b.lines = append(b.lines[:y+1], b.lines[y+2:]...)
		b.modified = true
	}
}

// NewLine splits the line at (x, y) and returns the new cursor position.
func (b *Buffer) NewLine(x, y int) (newX, newY int) {
	if y < 0 || y >= len(b.lines) {
		return 0, y
	}
	line := []rune(b.lines[y])
	if x < 0 {
		x = 0
	}
	if x > len(line) {
		x = len(line)
	}
	left := string(line[:x])
	right := string(line[x:])
	b.lines = append(b.lines[:y+1], append([]string{right}, b.lines[y+1:]...)...)
	b.lines[y] = left
	b.modified = true
	return 0, y + 1
}

// InsertText inserts multi-line text at (x, y) and returns new cursor position.
func (b *Buffer) InsertText(text string, x, y int) (newX, newY int) {
	if y < 0 || y >= len(b.lines) {
		return x, y
	}
	lines := strings.Split(text, "\n")
	firstLine := []rune(b.lines[y])
	if x < 0 {
		x = 0
	}
	if x > len(firstLine) {
		x = len(firstLine)
	}
	b.lines[y] = string(append(firstLine[:x], append([]rune(lines[0]), firstLine[x:]...)...))
	newX = x + len([]rune(lines[0]))
	newY = y
	for i := 1; i < len(lines); i++ {
		b.lines = append(b.lines[:newY+1], append([]string{lines[i]}, b.lines[newY+1:]...)...)
		newY++
		newX = len([]rune(lines[i]))
	}
	b.modified = true
	return
}

// Load reads a file into the buffer.
func (b *Buffer) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := strings.ReplaceAll(string(data), "\r\n", "\n")
	b.lines = strings.Split(content, "\n")
	// Remove trailing empty line from split
	if len(b.lines) > 0 && b.lines[len(b.lines)-1] == "" {
		b.lines = b.lines[:len(b.lines)-1]
	}
	if len(b.lines) == 0 {
		b.lines = []string{""}
	}
	b.filename = path
	b.modified = false
	return nil
}

// Save writes the buffer to its file.
func (b *Buffer) Save() error {
	data := []byte(strings.Join(b.lines, "\n"))
	if err := os.WriteFile(b.filename, data, 0644); err != nil {
		return err
	}
	b.modified = false
	return nil
}

// LineLen returns the length of line y in runes.
func (b *Buffer) LineLen(y int) int {
	if y < 0 || y >= len(b.lines) {
		return 0
	}
	return len([]rune(b.lines[y]))
}

// TotalChars returns the total number of characters (including newlines).
func (b *Buffer) TotalChars() (count int) {
	for _, l := range b.lines {
		count += len([]rune(l)) + 1
	}
	return
}

// LinesSlice returns a copy of the lines slice (for snapshots).
func (b *Buffer) LinesSlice() []string {
	s := make([]string, len(b.lines))
	copy(s, b.lines)
	return s
}

// SetLines restores lines from a snapshot.
func (b *Buffer) SetLines(lines []string) {
	b.lines = make([]string, len(lines))
	copy(b.lines, lines)
}
