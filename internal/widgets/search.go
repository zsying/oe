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
		sb.FindNext() // cycle to next match
		return true
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(sb.query) > 0 {
			sb.query = sb.query[:len(sb.query)-1]
		}
		return true
	case tcell.KeyRune:
		sb.query = append(sb.query, ev.Rune())
		sb.FindFirst() // find first match from cursor
		return true
	}
	return false
}

// runeIndex finds the first occurrence of needle in haystack (both rune slices).
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

// search finds the query in lineRunes starting from startX (rune index).
// Returns the rune index of the match, or -1.
func searchInLine(lineRunes []rune, queryRunes []rune, startX int) int {
	if startX >= len(lineRunes) || len(queryRunes) == 0 {
		return -1
	}
	idx := runeIndex(lineRunes[startX:], queryRunes)
	if idx < 0 {
		return -1
	}
	return startX + idx
}

// FindFirst searches from cursor.X (find first match at or after cursor).
func (sb *SearchBar) FindFirst() bool {
	return sb.findFrom(cursorAt)
}

// FindNext searches from cursor.X+1 (skip current, find next match).
func (sb *SearchBar) FindNext() bool {
	return sb.findFrom(cursorNext)
}

type findMode int

const (
	cursorAt   findMode = iota // search from current cursor position
	cursorNext                 // search from cursor + 1 (skip current match)
)

func (sb *SearchBar) findFrom(mode findMode) bool {
	q := string(sb.query)
	if q == "" {
		return false
	}
	sb.LastQuery = q

	qRunes := []rune(strings.ToLower(q))
	qLen := len(qRunes)
	buf := sb.Ed.Buffer
	cursor := sb.Ed.Cursor

	// Search forward from cursor.Y
	for y := cursor.Y; y < buf.Lines(); y++ {
		lineRunes := []rune(strings.ToLower(buf.Line(y)))
		startX := 0
		if y == cursor.Y {
			if mode == cursorNext {
				startX = cursor.X + 1
			} else {
				startX = cursor.X
			}
		}
		idx := searchInLine(lineRunes, qRunes, startX)
		if idx >= 0 {
			cursor.X = idx
			cursor.Y = y
			sb.Ed.Selection.Begin(cursor.X, cursor.Y)
			sb.Ed.Selection.Extend(cursor.X+qLen, cursor.Y)
			return true
		}
	}
	// Wrap around: search from beginning up to cursor.Y
	for y := 0; y <= cursor.Y; y++ {
		lineRunes := []rune(strings.ToLower(buf.Line(y)))
		startX := 0
		endX := len(lineRunes)
		if y == cursor.Y {
			if mode == cursorNext {
				endX = cursor.X
			} else {
				endX = cursor.X + 1
			}
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
