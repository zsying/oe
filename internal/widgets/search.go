package widgets

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/zsying/oe/internal/editor"
)

// SearchBar implements find-as-you-type search.
type SearchBar struct {
	Ed        *editor.Editor
	Active    bool
	query     []rune
	LastQuery string // preserved for F3 after closing search bar
	mode      searchMode
}

type searchMode int

const (
	searchFind searchMode = iota
	searchReplace
)

// NewSearchBar creates a new search bar.
func NewSearchBar(ed *editor.Editor) *SearchBar {
	return &SearchBar{Ed: ed}
}

// ToggleFind opens the find bar with the last query restored.
func (sb *SearchBar) ToggleFind() {
	sb.mode = searchFind
	if sb.LastQuery != "" {
		sb.query = []rune(sb.LastQuery)
	} else {
		sb.query = nil
	}
	sb.Active = true
}

// ToggleReplace opens the replace bar.
func (sb *SearchBar) ToggleReplace() {
	sb.mode = searchReplace
	sb.query = nil
	sb.Active = true
}

// HandleKey processes key events when search bar is active.
func (sb *SearchBar) HandleKey(ev *tcell.EventKey) bool {
	if !sb.Active {
		return false
	}
	switch ev.Key() {
	case tcell.KeyEscape:
		sb.Active = false
		return true
	case tcell.KeyEnter:
		sb.FindNext()
		return true
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(sb.query) > 0 {
			sb.query = sb.query[:len(sb.query)-1]
		}
		return true
	case tcell.KeyRune:
		sb.query = append(sb.query, ev.Rune())
		sb.FindNext()
		return true
	}
	return false
}

// FindNext searches for the next match, cycling through the buffer.
// Uses rune-aware searching for correct CJK handling.
func (sb *SearchBar) FindNext() bool {
	q := string(sb.query)
	if q == "" {
		return false
	}
	sb.LastQuery = q

	qLower := strings.ToLower(q)
	qRunes := []rune(qLower)
	qLen := len(qRunes)
	buf := sb.Ed.Buffer
	cursor := sb.Ed.Cursor

	// Search forward: start from cursor.Y, skip current match
	for y := cursor.Y; y < buf.Lines(); y++ {
		line := buf.Line(y)
		lineRunes := []rune(strings.ToLower(line))
		startX := 0
		if y == cursor.Y {
			startX = cursor.X + 1 // skip current position to advance
		}
		if startX >= len(lineRunes) {
			continue
		}
		// Search in rune slice
		idx := runeIndex(lineRunes[startX:], qRunes)
		if idx >= 0 {
			cursor.X = startX + idx
			cursor.Y = y
			sb.Ed.Selection.Begin(cursor.X, cursor.Y)
			sb.Ed.Selection.Extend(cursor.X+qLen, cursor.Y)
			return true
		}
	}
	// Wrap around: search from beginning up to cursor.Y
	for y := 0; y <= cursor.Y; y++ {
		line := buf.Line(y)
		lineRunes := []rune(strings.ToLower(line))
		startX := 0
		endX := len(lineRunes)
		if y == cursor.Y {
			endX = cursor.X
		}
		if startX >= endX || endX-startX < qLen {
			continue
		}
		idx := runeIndex(lineRunes[startX:endX], qRunes)
		if idx >= 0 {
			cursor.X = startX + idx
			cursor.Y = y
			sb.Ed.Selection.Begin(cursor.X, cursor.Y)
			sb.Ed.Selection.Extend(cursor.X+qLen, cursor.Y)
			return true
		}
	}
	return false
}

// runeIndex finds the first occurrence of needle in haystack (both rune slices).
// Returns the rune index or -1.
func runeIndex(haystack, needle []rune) int {
	if len(needle) == 0 {
		return 0
	}
	for i := 0; i <= len(haystack)-len(needle); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

// Query returns the current search query string.
func (sb *SearchBar) Query() string { return string(sb.query) }

// FindNextFromLast finds the next match using the saved LastQuery,
// without modifying the Active state (for F3 when search bar is closed).
func (sb *SearchBar) FindNextFromLast() bool {
	if sb.LastQuery == "" {
		return false
	}
	sb.query = []rune(sb.LastQuery)
	return sb.FindNext()
}

// Render draws the search bar at the bottom of the screen.
func (sb *SearchBar) Render(s tcell.Screen, w, h int) {
	if !sb.Active {
		return
	}

	bgStyle := tcell.StyleDefault.
		Background(tcell.ColorYellow).
		Foreground(tcell.ColorBlack)

	y := h - 2

	prefix := " Find: "
	if sb.mode == searchReplace {
		prefix = " Replace: "
	}

	text := prefix + string(sb.query)
	runes := []rune(text)

	for i := 0; i < w; i++ {
		ch := ' '
		if i < len(runes) {
			ch = runes[i]
		}
		s.SetCell(i, y, bgStyle, ch)
	}

	cx := len([]rune(prefix)) + len(sb.query)
	if cx < w {
		s.SetCell(cx, y, bgStyle.Reverse(true), ' ')
	}
}
