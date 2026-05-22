package widgets

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// HelpOverlay shows keyboard shortcut reference.
type HelpOverlay struct {
	Active bool
}

// NewHelpOverlay creates a new help overlay.
func NewHelpOverlay() *HelpOverlay {
	return &HelpOverlay{}
}

// Toggle opens or closes the help overlay.
func (h *HelpOverlay) Toggle() {
	h.Active = !h.Active
}

// HandleKey processes key events.
func (h *HelpOverlay) HandleKey(ev *tcell.EventKey) bool {
	if !h.Active {
		return false
	}
	if ev.Key() == tcell.KeyEscape {
		h.Active = false
		return true
	}
	return false
}

// Render draws the help overlay.
func (hlp *HelpOverlay) Render(s tcell.Screen, w, scrH int) {
	if !hlp.Active {
		return
	}

	shortcuts := []struct{ keys, desc string }{
		{"Ctrl+E", "Toggle View/Edit mode"},
		{"Ctrl+P", "Command palette"},
		{"Ctrl+O", "Open file"},
		{"Ctrl+S", "Save file"},
		{"Ctrl+Q", "Quit"},
		{"Ctrl+F", "Find"},
		{"Ctrl+H", "Replace"},
		{"F3", "Find next"},
		{"Ctrl+Z", "Undo"},
		{"", ""},
		{"Navigation:", ""},
		{"↑↓←→", "Move cursor"},
		{"Shift+↑↓←→", "Select text"},
		{"PgUp / PgDn", "Page up/down"},
		{"Space", "Page down (View mode)"},
		{"Shift+Space", "Page up (View mode)"},
		{"Home / End", "Start/end of line"},
		{"", ""},
		{"Edit mode:", ""},
		{"Ctrl+X", "Cut"},
		{"Ctrl+C", "Copy"},
		{"Ctrl+V", "Paste"},
		{"Ctrl+A", "Select all"},
		{"Del", "Delete forward"},
		{"Backspace", "Delete backward"},
		{"Enter", "New line"},
		{"Tab", "Insert spaces"},
	}

	const boxW = 44
	// Calculate visible height
	visCount := len(shortcuts) + 2
	if visCount > scrH-4 {
		visCount = scrH - 4
	}
	boxH := visCount
	if boxH < 10 {
		boxH = 10
	}

	boxX := (w - boxW) / 2
	boxY := (scrH - boxH) / 2

	bgStyle := tcell.StyleDefault.
		Background(tcell.ColorWhite).
		Foreground(tcell.ColorBlack)
	titleStyle := bgStyle.Reverse(true)
	sectionStyle := bgStyle.Foreground(tcell.ColorDodgerBlue)

	// Draw box
	for py := 0; py < boxH; py++ {
		for px := 0; px < boxW; px++ {
			s.SetCell(boxX+px, boxY+py, bgStyle, ' ')
		}
	}

	// Title
	title := " Keyboard Shortcuts (Esc to close) "
	for i, ch := range title {
		if i < boxW {
			s.SetCell(boxX+i, boxY, titleStyle, ch)
		}
	}

	// Lines
	lineY := boxY + 1
	for _, sc := range shortcuts {
		if lineY >= boxY+boxH-1 {
			break
		}
		if sc.keys == "" && sc.desc == "" {
			// Blank line
			lineY++
			continue
		}
		if sc.desc == "" {
			// Section header
			text := " " + sc.keys
			for i, ch := range text {
				if i < boxW-1 {
					s.SetCell(boxX+i, lineY, sectionStyle, ch)
				}
			}
			lineY++
			continue
		}

		// Key + description
		keyText := "  " + sc.keys
		descText := sc.desc
		keyW := runewidth.StringWidth(keyText)
		for i, ch := range keyText {
			if i < boxW-1 {
				s.SetCell(boxX+i, lineY, bgStyle, ch)
			}
		}
		// Description aligned
		descX := boxX + 22
		for i, ch := range descText {
			x := descX + i
			if x < boxX+boxW-1 {
				s.SetCell(x, lineY, bgStyle, ch)
			}
		}
		// Dotted connector
		for i := keyW; i < descX-boxX && i < boxW; i++ {
			s.SetCell(boxX+i, lineY, bgStyle, '.')
		}
		lineY++
	}

	// Footer
	footerText := fmt.Sprintf(" Esc:close | %d shortcuts ", len(shortcuts))
	fx := boxX + boxW - runewidth.StringWidth(footerText) - 1
	for i, ch := range footerText {
		if fx+i < boxX+boxW {
			s.SetCell(fx+i, boxY+boxH-1, bgStyle, ch)
		}
	}
}
