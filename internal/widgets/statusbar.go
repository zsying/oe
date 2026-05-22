package widgets

import (
	"fmt"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/zsying/oe/internal/editor"
)

// StatusBar renders the bottom status line.
type StatusBar struct {
	Ed *editor.Editor
}

// NewStatusBar creates a new status bar.
func NewStatusBar(ed *editor.Editor) *StatusBar {
	return &StatusBar{Ed: ed}
}

// Render draws the status bar at the given row — everything left-aligned.
func (sb *StatusBar) Render(s tcell.Screen, y, width int, style tcell.Style) {
	ed := sb.Ed
	buf := ed.Buffer

	modeStr := fmt.Sprintf("[%s]", ed.Mode)
	filename := filepath.Base(buf.Filename())
	if filename == "" {
		filename = "(untitled)"
	}
	posStr := fmt.Sprintf("Ln %d, Col %d", ed.Cursor.Y+1, ed.Cursor.X+1)
	modifiedStr := ""
	if buf.Modified() {
		modifiedStr = " ●"
	}

	text := fmt.Sprintf(" %s  Ctrl+E编辑  %s  %s%s  Ctrl+P面板",
		modeStr, filename, posStr, modifiedStr)

	col := 0
	for _, ch := range text {
		if col >= width {
			break
		}
		w := runewidth.RuneWidth(ch)
		if col+w <= width {
			s.SetCell(col, y, style, ch)
			col += w
		} else {
			break
		}
	}
	// Fill rest with spaces
	for col < width {
		s.SetCell(col, y, style, ' ')
		col++
	}
}
