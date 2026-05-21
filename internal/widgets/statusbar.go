package widgets

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
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

// Render draws the status bar at the given row.
func (sb *StatusBar) Render(s tcell.Screen, y, width int, style tcell.Style) {
	ed := sb.Ed
	buf := ed.Buffer

	modeStr := fmt.Sprintf("[%s]", ed.Mode)
	filename := buf.Filename()
	if filename == "" {
		filename = "(untitled)"
	}
	posStr := fmt.Sprintf("Ln %d, Col %d", ed.Cursor.Y+1, ed.Cursor.X+1)
	modifiedStr := ""
	if buf.Modified() {
		modifiedStr = " ●"
	}
	fileType := buf.FileType()
	if fileType != "" {
		fileType = " " + fileType
	}

	hint := " Ctrl+P:命令面板 "
	text := fmt.Sprintf(" %s  %s%s  %s  UTF-8%s",
		modeStr, filename, fileType, posStr, modifiedStr)

	runes := []rune(text)
	hintRunes := []rune(hint)
	maxW := width

	for i := 0; i < maxW; i++ {
		ch := ' '
		if i < len(runes) {
			ch = runes[i]
		}
		// Draw hint at right side if there's room
		if i >= maxW-len(hintRunes) && i < maxW && len(hintRunes)+len(runes) < maxW-4 {
			hi := i - (maxW - len(hintRunes))
			if hi >= 0 && hi < len(hintRunes) {
				ch = hintRunes[hi]
			}
		}
		s.SetCell(i, y, style, ch)
	}
}
