package widgets

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
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
	// DialogThreeWay shows three buttons (e.g. Save / Don't Save / Cancel).
	DialogThreeWay
)

// Dialog is a modal overlay for confirmations and input.
type Dialog struct {
	Type     DialogType
	Title    string
	Message  string
	Input    []rune
	Active   bool
	callback interface{} // func(string, bool) or func(string)
}

// NewDialog creates a new dialog.
func NewDialog() *Dialog {
	return &Dialog{}
}

// Show displays a confirm/message dialog.
// callback receives (result string, ok bool). ok=false means cancelled.
func (d *Dialog) Show(title, msg string, callback func(string, bool)) {
	d.Type = DialogMessage
	d.Title = title
	d.Message = msg
	d.Input = nil
	d.callback = callback
	d.Active = true
}

// ShowInput displays an input dialog with a default value.
// callback receives (result string, ok bool). ok=false means cancelled.
func (d *Dialog) ShowInput(title, msg, defaultVal string, callback func(string, bool)) {
	d.Type = DialogInput
	d.Title = title
	d.Message = msg
	d.Input = []rune(defaultVal)
	d.callback = callback
	d.Active = true
}

// ShowThreeWay displays a dialog with three buttons: Save, Don't Save, Cancel.
// callback receives the action string: "save", "discard", or "cancel".
func (d *Dialog) ShowThreeWay(title, msg string, callback func(action string)) {
	d.Type = DialogThreeWay
	d.Title = title
	d.Message = msg
	d.Input = nil
	d.callback = callback
	d.Active = true
}

// HandleKey processes a key event when the dialog is active.
// Returns true if the key was consumed.
func (d *Dialog) HandleKey(ev *tcell.EventKey) bool {
	if !d.Active {
		return false
	}

	if d.Type == DialogThreeWay {
		return d.handleKeyThreeWay(ev)
	}

	switch ev.Key() {
	case tcell.KeyEscape:
		d.Active = false
		if cb, ok := d.callback.(func(string, bool)); ok {
			cb("", false)
		}
		return true
	case tcell.KeyEnter:
		d.Active = false
		if cb, ok := d.callback.(func(string, bool)); ok {
			if d.Type == DialogInput {
				cb(string(d.Input), true)
			} else {
				cb("", true)
			}
		}
		return true
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(d.Input) > 0 {
			d.Input = d.Input[:len(d.Input)-1]
		}
		return true
	case tcell.KeyRune:
		if d.Type == DialogInput {
			d.Input = append(d.Input, ev.Rune())
		}
		return true
	}
	return false
}

func (d *Dialog) handleKeyThreeWay(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyEscape:
		// Esc = Cancel
		d.Active = false
		if cb, ok := d.callback.(func(string)); ok {
			cb("cancel")
		}
		return true
	case tcell.KeyRune:
		switch ev.Rune() {
		case 's', 'S':
			// S = Save
			d.Active = false
			if cb, ok := d.callback.(func(string)); ok {
				cb("save")
			}
			return true
		case 'd', 'D':
			// D = Don't Save / Discard
			d.Active = false
			if cb, ok := d.callback.(func(string)); ok {
				cb("discard")
			}
			return true
		case 'c', 'C':
			// C = Cancel
			d.Active = false
			if cb, ok := d.callback.(func(string)); ok {
				cb("cancel")
			}
			return true
		}
		fallthrough
	default:
		// Enter = Save (default action)
		if ev.Key() == tcell.KeyEnter {
			d.Active = false
			if cb, ok := d.callback.(func(string)); ok {
				cb("save")
			}
			return true
		}
	}
	return false
}

// Render draws the dialog overlay.
func (d *Dialog) Render(s tcell.Screen, w, h int) {
	if !d.Active {
		return
	}

	const (
		dialogW = 56
		dialogH = 10
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

	// Message — wrap long messages
	msgRunes := []rune(d.Message)
	msgY := dialogY + 2
	msgMaxW := dialogW - 4
	col := 0
	for i := 0; i < len(msgRunes) && msgY < dialogY+dialogH-3; i++ {
		if msgRunes[i] == '\n' || col >= msgMaxW {
			msgY++
			col = 0
		}
		if msgRunes[i] != '\n' {
			s.SetCell(dialogX+2+col, msgY, bgStyle, msgRunes[i])
			col++
		}
	}

	if d.Type == DialogInput {
		// Input field
		inputStr := " " + string(d.Input)
		for i, ch := range inputStr {
			s.SetCell(dialogX+2+i, dialogY+5, bgStyle, ch)
		}
		// Input cursor
		cx := dialogX + 2 + runewidth.StringWidth(string(d.Input))
		if cx < dialogX+dialogW-2 {
			s.SetCell(cx, dialogY+5, bgStyle.Reverse(true), ' ')
		}
	} else if d.Type == DialogThreeWay {
		// Three buttons: [Save]  [Don't Save]  [Cancel]
		btnY := dialogY + dialogH - 2
		btns := []struct {
			label string
			key   string
		}{
			{"Save", "(S)"},
			{"Don't Save", "(D)"},
			{"Cancel", "(C)"},
		}

		totalW := 0
		for _, b := range btns {
			w := runewidth.StringWidth(b.label) + runewidth.StringWidth(b.key) + 4
			totalW += w
		}
		totalW += 4 // padding between buttons

		curX := dialogX + (dialogW-totalW)/2
		if curX < dialogX+1 {
			curX = dialogX + 1
		}

		for _, b := range btns {
			btnStyle := bgStyle
			label := " " + b.label + " " + b.key + " "
			for j, ch := range label {
				s.SetCell(curX+j, btnY, btnStyle, ch)
			}
			curX += runewidth.StringWidth(label) + 2
		}
	} else if d.Type == DialogMessage {
		// OK button
		okText := " [ OK ] "
		okX := dialogX + dialogW - len(okText) - 2
		for i, ch := range okText {
			s.SetCell(okX+i, dialogY+dialogH-2, bgStyle, ch)
		}
	}
}
