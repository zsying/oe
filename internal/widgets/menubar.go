package widgets

import (
	"github.com/gdamore/tcell/v2"
	"github.com/user/editor/internal/editor"
)

// MenuEntry represents a single item in a dropdown menu.
type MenuEntry struct {
	Title     string
	CommandID string // empty = separator
	Shortcut  string
}

// MenuItem represents a top-level menu with its dropdown.
type MenuItem struct {
	Title   string
	Submenu []MenuEntry
}

// MenuBar renders the top menu bar with dropdown support.
type MenuBar struct {
	Ed      *editor.Editor
	Items   []MenuItem
	openIdx int // -1 = closed
}

// DefaultMenus returns the standard menu layout.
func DefaultMenus() []MenuItem {
	return []MenuItem{
		{
			Title: "File",
			Submenu: []MenuEntry{
				{Title: "New", CommandID: "file.new"},
				{Title: "Open…", CommandID: "file.open", Shortcut: "Ctrl+O"},
				{Title: "", CommandID: ""},
				{Title: "Save", CommandID: "file.save", Shortcut: "Ctrl+S"},
				{Title: "Save As…", CommandID: "file.saveAs"},
				{Title: "", CommandID: ""},
				{Title: "Exit", CommandID: "app.quit", Shortcut: "Ctrl+Q"},
			},
		},
		{
			Title: "Edit",
			Submenu: []MenuEntry{
				{Title: "Undo", CommandID: "edit.undo", Shortcut: "Ctrl+Z"},
				{Title: "", CommandID: ""},
				{Title: "Cut", CommandID: "edit.cut", Shortcut: "Ctrl+X"},
				{Title: "Copy", CommandID: "edit.copy", Shortcut: "Ctrl+C"},
				{Title: "Paste", CommandID: "edit.paste", Shortcut: "Ctrl+V"},
				{Title: "Delete", CommandID: "edit.delete", Shortcut: "Del"},
				{Title: "", CommandID: ""},
				{Title: "Select All", CommandID: "edit.selectAll", Shortcut: "Ctrl+A"},
			},
		},
		{
			Title: "Selection",
			Submenu: []MenuEntry{
				{Title: "Select All", CommandID: "edit.selectAll", Shortcut: "Ctrl+A"},
			},
		},
		{
			Title: "Find",
			Submenu: []MenuEntry{
				{Title: "Find…", CommandID: "find.find", Shortcut: "Ctrl+F"},
				{Title: "Find Next", CommandID: "find.next", Shortcut: "F3"},
				{Title: "", CommandID: ""},
				{Title: "Replace…", CommandID: "find.replace", Shortcut: "Ctrl+H"},
			},
		},
		{
			Title: "Help",
			Submenu: []MenuEntry{
				{Title: "About", CommandID: ""},
			},
		},
	}
}

// NewMenuBar creates a new menu bar.
func NewMenuBar(ed *editor.Editor) *MenuBar {
	return &MenuBar{
		Ed:      ed,
		Items:   DefaultMenus(),
		openIdx: -1,
	}
}

// IsOpen returns whether a menu dropdown is open.
func (mb *MenuBar) IsOpen() bool { return mb.openIdx >= 0 }

// Close closes any open menu.
func (mb *MenuBar) Close() { mb.openIdx = -1 }

// Render draws the menu bar at the given row.
func (mb *MenuBar) Render(s tcell.Screen, y, width int, style, selStyle tcell.Style) {
	col := 1
	// Draw menu titles
	for i, item := range mb.Items {
		title := " " + item.Title + " "
		st := style
		if mb.openIdx == i {
			st = selStyle
		}
		for j, ch := range title {
			s.SetCell(col+j, y, st, ch)
		}
		col += len([]rune(title)) + 1
	}
	// Fill remaining bar
	for col < width {
		s.SetCell(col, y, style, ' ')
		col++
	}

	// Draw open submenu
	if mb.openIdx >= 0 && mb.openIdx < len(mb.Items) {
		mb.renderSubmenu(s, mb.openIdx)
	}
}

func (mb *MenuBar) renderSubmenu(s tcell.Screen, idx int) {
	item := mb.Items[idx]

	// Calculate X position of this menu
	col := 1
	for i := 0; i < idx; i++ {
		col += len([]rune(mb.Items[i].Title)) + 2
	}

	popBg := tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack)

	for i, entry := range item.Submenu {
		y := 1 + i

		if entry.Title == "" {
			// Separator — draw a line
			for x := col; x < col+30; x++ {
				s.SetCell(x, y, popBg, '─')
			}
			continue
		}

		text := " " + entry.Title
		for j, ch := range text {
			s.SetCell(col+j, y, popBg, ch)
		}

		// Shortcut aligned right
		if entry.Shortcut != "" {
			sc := " " + entry.Shortcut
			scX := col + 28 - len([]rune(sc))
			for j, ch := range sc {
				s.SetCell(scX+j, y, popBg, ch)
			}
		}
	}
}

// HandleMouseClick processes clicks on the menu bar.
// Returns (commandID, handled).
func (mb *MenuBar) HandleMouseClick(x, y int) (string, bool) {
	if y != 0 {
		return "", false
	}
	col := 1
	for i, item := range mb.Items {
		title := " " + item.Title + " "
		titleLen := len([]rune(title))
		if x >= col && x < col+titleLen {
			if mb.openIdx == i {
				mb.openIdx = -1 // toggle off
			} else {
				mb.openIdx = i
			}
			return "", true
		}
		col += titleLen + 1
	}
	// Clicked outside any menu title
	mb.openIdx = -1
	return "", true
}

// HandleSubmenuClick processes clicks on an open submenu item.
func (mb *MenuBar) HandleSubmenuClick(x, y int) (string, bool) {
	if mb.openIdx < 0 {
		return "", false
	}
	// Calculate menu X position
	col := 1
	for i := 0; i < mb.openIdx; i++ {
		col += len([]rune(mb.Items[i].Title)) + 2
	}

	item := mb.Items[mb.openIdx]
	subY := y - 1 // submenu starts at line 1
	if subY < 0 || subY >= len(item.Submenu) {
		mb.openIdx = -1
		return "", true
	}
	entry := item.Submenu[subY]
	if entry.CommandID == "" {
		// Separator — ignore
		return "", true
	}
	mb.openIdx = -1
	return entry.CommandID, true
}
