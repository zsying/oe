package widgets

import (
	"fmt"
	"path/filepath"

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

	modeStr := fmt.Sprintf("[%s] Ctrl+E切换", ed.Mode)
	filename := filepath.Base(buf.Filename())
	if filename == "" {
		filename = "(untitled)"
	}
	posStr := fmt.Sprintf("Ln %d, Col %d", ed.Cursor.Y+1, ed.Cursor.X+1)
	modifiedStr := ""
	if buf.Modified() {
		modifiedStr = " ●"
	}

	hint := " Ctrl+P:命令面板 "
	text := fmt.Sprintf(" %s  %s  %s%s",
		modeStr, filename, posStr, modifiedStr)

	leftRunes := []rune(text)
	hintRunes := []rune(hint)

	// Reserve space for hint on the right
	hintW := len(hintRunes)
	textMax := width - hintW
	if textMax < 10 {
		textMax = 10
	}

	for i := 0; i < width; i++ {
		ch := ' '
		if i < textMax && i < len(leftRunes) {
			ch = leftRunes[i]
		}
		// Draw hint on the right
		if i >= width-hintW {
			hi := i - (width - hintW)
			if hi >= 0 && hi < len(hintRunes) {
				ch = hintRunes[hi]
			}
		}
		s.SetCell(i, y, style, ch)
	}
}
