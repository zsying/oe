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

// Render draws the status bar at the given row, using runewidth for CJK support.
func (sb *StatusBar) Render(s tcell.Screen, y, width int, style tcell.Style) {
	ed := sb.Ed
	buf := ed.Buffer

	modeStr := fmt.Sprintf("[%s] Ctrl+E", ed.Mode)
	filename := filepath.Base(buf.Filename())
	if filename == "" {
		filename = "(untitled)"
	}
	posStr := fmt.Sprintf("Ln %d, Col %d", ed.Cursor.Y+1, ed.Cursor.X+1)
	modifiedStr := ""
	if buf.Modified() {
		modifiedStr = " ●"
	}

	hint := " Ctrl+P:面板 "
	leftText := fmt.Sprintf(" %s  %s  %s%s",
		modeStr, filename, posStr, modifiedStr)

	// Use runewidth for accurate display width of CJK characters
	hintWidth := runewidth.StringWidth(hint)

	// Reserve space for hint on the right
	textMax := width - hintWidth
	if textMax < 10 {
		textMax = 10
	}

	// Draw left text (truncated if too long)
	col := 0
	for _, ch := range leftText {
		if col >= textMax {
			break
		}
		w := runewidth.RuneWidth(ch)
		if col+w <= textMax {
			s.SetCell(col, y, style, ch)
			col += w
		} else {
			// Partial character - fill remaining with something
			break
		}
	}

	// Fill between left and hint with spaces
	for col < width-hintWidth {
		s.SetCell(col, y, style, ' ')
		col++
	}

	// Draw hint on the right
	hintCol := 0
	for _, ch := range hint {
		if col >= width {
			break
		}
		s.SetCell(col, y, style, ch)
		col++
		hintCol++
	}

	// Fill remaining
	for col < width {
		s.SetCell(col, y, style, ' ')
		col++
	}
}
