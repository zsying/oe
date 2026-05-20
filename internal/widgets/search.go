package widgets

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/user/editor/internal/editor"
)

// SearchBar implements find-as-you-type search.
type SearchBar struct {
	Ed      *editor.Editor
	Active  bool
	query   []rune
	replace []rune
	mode    searchMode // find or replace
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

// ToggleFind opens the find bar.
func (sb *SearchBar) ToggleFind() {
	sb.mode = searchFind
	sb.query = nil
	sb.Active = true
}

// ToggleReplace opens the replace bar.
func (sb *SearchBar) ToggleReplace() {
	sb.mode = searchReplace
	sb.query = nil
	sb.replace = nil
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
		if sb.mode == searchFind {
			if len(sb.query) > 0 {
				sb.query = sb.query[:len(sb.query)-1]
			}
		} else {
			if len(sb.query) > 0 {
				sb.query = sb.query[:len(sb.query)-1]
			}
		}
		return true
	case tcell.KeyTab:
		if sb.mode == searchReplace {
			// Toggle between query and replace field
		}
		return true
	case tcell.KeyRune:
		if sb.mode == searchFind {
			sb.query = append(sb.query, ev.Rune())
			sb.FindNext()
		} else {
			sb.query = append(sb.query, ev.Rune())
			sb.FindNext()
		}
		return true
	}
	return false
}

// FindNext searches for the next match after the cursor.
func (sb *SearchBar) FindNext() bool {
	q := string(sb.query)
	if q == "" {
		return false
	}
	qLower := strings.ToLower(q)
	buf := sb.Ed.Buffer
	cursor := sb.Ed.Cursor

	// Search from current cursor position forward
	for y := cursor.Y; y < buf.Lines(); y++ {
		line := buf.Line(y)
		lineLower := strings.ToLower(line)
		startX := 0
		if y == cursor.Y {
			startX = cursor.X
		}
		idx := strings.Index(lineLower[startX:], qLower)
		if idx >= 0 {
			cursor.X = startX + idx
			cursor.Y = y
			// Select the match
			sb.Ed.Selection.Begin(cursor.X, cursor.Y)
			sb.Ed.Selection.Extend(cursor.X+len([]rune(q)), cursor.Y)
			return true
		}
	}
	// Wrap around: search from beginning
	for y := 0; y <= cursor.Y; y++ {
		line := buf.Line(y)
		lineLower := strings.ToLower(line)
		startX := 0
		endX := len([]rune(line))
		if y == cursor.Y {
			endX = cursor.X
		}
		idx := strings.Index(lineLower[startX:endX], qLower)
		if idx >= 0 {
			cursor.X = startX + idx
			cursor.Y = y
			sb.Ed.Selection.Begin(cursor.X, cursor.Y)
			sb.Ed.Selection.Extend(cursor.X+len([]rune(q)), cursor.Y)
			return true
		}
	}
	return false
}

// Query returns the current search query string.
func (sb *SearchBar) Query() string { return string(sb.query) }

// Render draws the search bar at the bottom of the screen.
func (sb *SearchBar) Render(s tcell.Screen, w, h int) {
	if !sb.Active {
		return
	}

	bgStyle := tcell.StyleDefault.
		Background(tcell.ColorYellow).
		Foreground(tcell.ColorBlack)

	y := h - 2 // One row above status bar

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

	// Cursor in search bar
	cx := len([]rune(prefix)) + len(sb.query)
	if cx < w {
		s.SetCell(cx, y, bgStyle.Reverse(true), ' ')
	}
}
