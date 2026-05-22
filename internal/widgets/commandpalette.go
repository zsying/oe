package widgets

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/zsying/oe/internal/editor"
)

const maxVisibleResults = 10

// CommandPalette implements Ctrl+P command search overlay.
type CommandPalette struct {
	Ed           *editor.Editor
	Active       bool
	query        []rune
	results      []editor.Command
	selIdx       int
	scrollOffset int
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
		cp.scrollOffset = 0
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
		n := len(cp.results)
		if n == 0 {
			return true
		}
		if cp.selIdx > 0 {
			cp.selIdx--
		} else {
			// Wrap to bottom
			cp.selIdx = n - 1
			cp.scrollOffset = n - maxVisibleResults
			if cp.scrollOffset < 0 {
				cp.scrollOffset = 0
			}
		}
		// Scroll if above visible area
		if cp.selIdx < cp.scrollOffset {
			cp.scrollOffset = cp.selIdx
		}
		return true
	case tcell.KeyDown:
		n := len(cp.results)
		if n == 0 {
			return true
		}
		if cp.selIdx < n-1 {
			cp.selIdx++
		} else {
			// Wrap to top
			cp.selIdx = 0
			cp.scrollOffset = 0
		}
		// Scroll if below visible area
		if cp.selIdx >= cp.scrollOffset+maxVisibleResults {
			cp.scrollOffset = cp.selIdx - maxVisibleResults + 1
		}
		return true
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(cp.query) > 0 {
			cp.query = cp.query[:len(cp.query)-1]
			cp.selIdx = 0
			cp.scrollOffset = 0
			cp.results = cp.filter()
		}
		return true
	case tcell.KeyRune:
		cp.query = append(cp.query, ev.Rune())
		cp.selIdx = 0
		cp.scrollOffset = 0
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
		// Match by title (fuzzy)
		if fuzzyMatch(cmd.Title, qLower) {
			result = append(result, cmd)
			continue
		}
		// Match by ID parts: e.g. "search" matches "find.find" → "find"
		for _, part := range strings.Split(cmd.ID, ".") {
			if fuzzyMatch(part, qLower) {
				result = append(result, cmd)
				break
			}
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

	listLen := len(cp.results)
	visCount := maxVisibleResults
	if visCount > listLen {
		visCount = listLen
	}
	totalH := visCount + 3 // input line + 1 blank + results + 1 blank bottom
	if totalH < 4 {
		totalH = 4
	}
	if totalH > h-2 {
		totalH = h - 2
	}

	const paletteW = 48
	paletteX := (w - paletteW) / 2
	paletteY := h / 4

	bgStyle := tcell.StyleDefault.
		Background(tcell.ColorWhite).
		Foreground(tcell.ColorBlack)
	selStyle := bgStyle.
		Background(tcell.ColorCornflowerBlue).
		Foreground(tcell.ColorWhite)

	// Draw background rectangle (with extra cell on left for wide-character bleed)
	for py := 0; py < totalH; py++ {
		if paletteX > 0 {
			s.SetCell(paletteX-1, paletteY+py, bgStyle, ' ')
		}
		for px := 0; px < paletteW; px++ {
			s.SetCell(paletteX+px, paletteY+py, bgStyle, ' ')
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

	// Draw visible results with scrolling
	startIdx := cp.scrollOffset
	endIdx := startIdx + visCount
	if endIdx > listLen {
		endIdx = listLen
	}

	for i := startIdx; i < endIdx; i++ {
		cmd := cp.results[i]
		st := bgStyle
		if i == cp.selIdx {
			st = selStyle
		}
		row := paletteY + 2 + (i - startIdx)

		// Fill entire row with background (selected or normal)
		for x := 0; x < paletteW; x++ {
			s.SetCell(paletteX+x, row, st, ' ')
		}

		// Draw command name
		text := "  " + cmd.Title
		for j, ch := range text {
			s.SetCell(paletteX+j, row, st, ch)
		}
		// Shortcut on right
		if cmd.Shortcut != "" {
			sc := " " + cmd.Shortcut
			scX := paletteX + paletteW - len([]rune(sc)) - 2
			if scX > paletteX {
				for j, ch := range sc {
					s.SetCell(scX+j, row, st, ch)
				}
			}
		}
	}

	// Scroll indicator
	if listLen > visCount {
		scrollText := ""
		if cp.scrollOffset > 0 && cp.scrollOffset+visCount < listLen {
			scrollText = " ▲▼ "
		} else if cp.scrollOffset > 0 {
			scrollText = " ▲ "
		} else {
			scrollText = " ▼ "
		}
		for j, ch := range scrollText {
			s.SetCell(paletteX+paletteW-len([]rune(scrollText))+j, paletteY+totalH-1, bgStyle, ch)
		}
	}
}
