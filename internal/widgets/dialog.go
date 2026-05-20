package widgets

import (
	"github.com/gdamore/tcell/v2"
)

// DialogType determines the dialog mode.
type DialogType int

const (
	// DialogConfirm shows OK/Cancel prompt.
	DialogConfirm DialogType = iota
	// DialogInput shows a text input field.
	DialogInput
	// DialogMessage shows an informational message.
	DialogMessage
)

// Dialog is a modal overlay for confirmations and input.
type Dialog struct {
	Type     DialogType
	Title    string
	Message  string
	Input    []rune
	Active   bool
	callback func(result string, ok bool)
}

// NewDialog creates a new dialog.
func NewDialog() *Dialog {
	return &Dialog{}
}

// Show displays a confirm/message dialog.
func (d *Dialog) Show(title, msg string, callback func(string, bool)) {
	d.Type = DialogConfirm
	d.Title = title
	d.Message = msg
	d.Input = nil
	d.callback = callback
	d.Active = true
}

// ShowInput displays an input dialog with a default value.
func (d *Dialog) ShowInput(title, msg, defaultVal string, callback func(string, bool)) {
	d.Type = DialogInput
	d.Title = title
	d.Message = msg
	d.Input = []rune(defaultVal)
	d.callback = callback
	d.Active = true
}

// HandleKey processes a key event when the dialog is active.
// Returns true if the key was consumed.
func (d *Dialog) HandleKey(ev *tcell.EventKey) bool {
	if !d.Active {
		return false
	}
	switch ev.Key() {
	case tcell.KeyEscape:
		d.Active = false
		if d.callback != nil {
			d.callback("", false)
		}
		return true
	case tcell.KeyEnter:
		d.Active = false
		if d.callback != nil {
			if d.Type == DialogInput {
				d.callback(string(d.Input), true)
			} else {
				d.callback("", true)
			}
		}
		return true
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(d.Input) > 0 {
			d.Input = d.Input[:len(d.Input)-1]
		}
		return true
	case tcell.KeyRune:
		d.Input = append(d.Input, ev.Rune())
		return true
	}
	return false
}

// Render draws the dialog overlay.
func (d *Dialog) Render(s tcell.Screen, w, h int) {
	if !d.Active {
		return
	}

	const (
		dialogW  = 50
		dialogH  = 8
	)

	dialogX := (w - dialogW) / 2
	dialogY := h / 3

	bgStyle := tcell.StyleDefault.
		Background(tcell.ColorWhite).
		Foreground(tcell.ColorBlack)

	// Draw background box
	for py := 0; py < dialogH; py++ {
		for px := 0; px < dialogW; px++ {
			s.SetCell(dialogX+px, dialogY+py, bgStyle, ' ')
		}
	}

	// Title bar
	titleStyle := bgStyle.Reverse(true)
	titleRunes := []rune(" " + d.Title + " ")
	for i := 0; i < len(titleRunes) && i < dialogW; i++ {
		s.SetCell(dialogX+i, dialogY, titleStyle, titleRunes[i])
	}

	// Message
	msgRunes := []rune(d.Message)
	for i := 0; i < len(msgRunes) && i < dialogW-4; i++ {
		s.SetCell(dialogX+2+i, dialogY+2, bgStyle, msgRunes[i])
	}

	if d.Type == DialogInput {
		// Input field
		inputStr := " " + string(d.Input)
		for i, ch := range inputStr {
			s.SetCell(dialogX+2+i, dialogY+4, bgStyle, ch)
		}
		// Input cursor
		cx := dialogX + 2 + len(d.Input)
		if cx < dialogX+dialogW-1 {
			s.SetCell(cx, dialogY+4, bgStyle.Reverse(true), ' ')
		}
	} else {
		// OK button
		okText := " [ OK ] "
		okX := dialogX + dialogW - len(okText) - 2
		for i, ch := range okText {
			s.SetCell(okX+i, dialogY+dialogH-2, bgStyle, ch)
		}
	}
}
