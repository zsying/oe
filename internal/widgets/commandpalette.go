package widgets

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/user/editor/internal/editor"
)

// CommandPalette implements Ctrl+P command search overlay.
type CommandPalette struct {
	Ed      *editor.Editor
	Active  bool
	query   []rune
	results []editor.Command
	selIdx  int
}

// NewCommandPalette creates a new command palette.
func NewCommandPalette(ed *editor.Editor) *CommandPalette {
	return &CommandPalette{Ed: ed}
}

// Toggle opens or closes the palette.
func (cp *CommandPalette) Toggle() {
	cp.Active = !cp.Active
	if cp.Active {
		cp.query = nil
		cp.selIdx = 0
		cp.results = cp.filter()
	}
}

// HandleKey processes key events when the palette is active.
// Returns true if the key was consumed.
func (cp *CommandPalette) HandleKey(ev *tcell.EventKey) bool {
	if !cp.Active {
		return false
	}
	switch ev.Key() {
	case tcell.KeyEscape:
		cp.Active = false
		return true
	case tcell.KeyEnter:
		if cp.selIdx >= 0 && cp.selIdx < len(cp.results) {
			cp.results[cp.selIdx].Handler()
			cp.Active = false
		}
		return true
	case tcell.KeyUp:
		if cp.selIdx > 0 {
			cp.selIdx--
		}
		return true
	case tcell.KeyDown:
		if cp.selIdx < len(cp.results)-1 {
			cp.selIdx++
		}
		return true
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(cp.query) > 0 {
			cp.query = cp.query[:len(cp.query)-1]
			cp.selIdx = 0
			cp.results = cp.filter()
		}
		return true
	case tcell.KeyRune:
		cp.query = append(cp.query, ev.Rune())
		cp.selIdx = 0
		cp.results = cp.filter()
		return true
	}
	return false
}

func (cp *CommandPalette) filter() []editor.Command {
	all := cp.Ed.Commands.All()
	q := strings.TrimSpace(string(cp.query))
	if q == "" {
		return all
	}
	qLower := strings.ToLower(q)
	var result []editor.Command
	for _, cmd := range all {
		if fuzzyMatch(cmd.Title, qLower) {
			result = append(result, cmd)
		}
	}
	return result
}

// fuzzyMatch returns true if all characters of query appear in order in text.
func fuzzyMatch(text, query string) bool {
	textLower := strings.ToLower(text)
	qi := 0
	for _, ch := range textLower {
		if qi < len(query) && byte(ch) == query[qi] {
			qi++
		}
	}
	return qi == len(query)
}

// Render draws the command palette overlay.
func (cp *CommandPalette) Render(s tcell.Screen, w, h int) {
	if !cp.Active {
		return
	}

	const paletteW = 45
	listH := len(cp.results) + 3
	if listH > 14 {
		listH = 14
	}
	if listH < 4 {
		listH = 4
	}
	paletteX := (w - paletteW) / 2
	paletteY := h / 4

	bgStyle := tcell.StyleDefault.
		Background(tcell.ColorWhite).
		Foreground(tcell.ColorBlack)
	selStyle := bgStyle.
		Background(tcell.ColorCornflowerBlue).
		Foreground(tcell.ColorWhite)

	// Draw background rectangle
	for py := 0; py < listH; py++ {
		for px := 0; px < paletteW; px++ {
			ch := ' '
			s.SetCell(paletteX+px, paletteY+py, bgStyle, ch)
		}
	}

	// Draw prompt
	prompt := "> " + string(cp.query)
	for i, ch := range prompt {
		s.SetCell(paletteX+i, paletteY, bgStyle, ch)
	}

	// Draw cursor after prompt
	cursorX := paletteX + 2 + len(cp.query)
	if cursorX < paletteX+paletteW-1 {
		s.SetCell(cursorX, paletteY, bgStyle.Reverse(true), ' ')
	}

	// Draw results
	for i, cmd := range cp.results {
		if i >= listH-2 {
			break
		}
		st := bgStyle
		if i == cp.selIdx {
			st = selStyle
		}
		text := "  " + cmd.Title
		for j, ch := range text {
			s.SetCell(paletteX+j, paletteY+2+i, st, ch)
		}
		// Shortcut on right
		if cmd.Shortcut != "" {
			sc := " " + cmd.Shortcut
			scX := paletteX + paletteW - len([]rune(sc)) - 2
			if scX > paletteX {
				for j, ch := range sc {
					s.SetCell(scX+j, paletteY+2+i, st, ch)
				}
			}
		}
	}
}
