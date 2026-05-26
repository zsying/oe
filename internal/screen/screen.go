package screen

import (
	"os"
	"path/filepath"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/zsying/oe/internal/buffer"
	"github.com/zsying/oe/internal/editor"
	"github.com/zsying/oe/internal/widgets"
)

// Screen manages the tcell terminal screen and event loop.
type Screen struct {
	tcell        tcell.Screen
	ed           *editor.Editor
	quit         bool
	width        int
	height       int
	scrollOffset int
	palette      Palette
	statusBar    *widgets.StatusBar
	cmdPalette   *widgets.CommandPalette
	searchBar    *widgets.SearchBar
	fileBrowser  *widgets.FileBrowser
	helpOverlay  *widgets.HelpOverlay
	dialog       *widgets.Dialog
	dragging     bool // whether mouse drag-selection is active
}

// New creates and initializes a new Screen.
func New(ed *editor.Editor) (*Screen, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := s.Init(); err != nil {
		return nil, err
	}
	s.SetStyle(tcell.StyleDefault)
	s.EnableMouse()
	s.EnablePaste()

	w, h := s.Size()

	sc := &Screen{
		tcell:   s,
		ed:      ed,
		width:   w,
		height:  h,
		palette: DefaultPalette(),
	}
	sc.statusBar = widgets.NewStatusBar(ed)
	sc.cmdPalette = widgets.NewCommandPalette(ed)
	sc.searchBar = widgets.NewSearchBar(ed)
	sc.fileBrowser = widgets.NewFileBrowser(ed)
	sc.helpOverlay = widgets.NewHelpOverlay()
	sc.dialog = widgets.NewDialog()

	// Override dialog-requiring command handlers
	ed.Commands.Find("file.open").Handler = func() error {
		doOpen := func(path string) {
			if err := ed.OpenFile(path); err != nil {
				sc.dialog.Show("Error", "Failed to open file: "+err.Error(), nil)
			}
		}
		tryOpen := func() {
			sc.fileBrowser.Open(func(path string) {
				doOpen(path)
			})
		}
		if ed.Buffer.Modified() {
			sc.dialog.ShowThreeWay("Unsaved Changes",
				"Current file has unsaved changes. Save before opening another file?",
				func(action string) {
					switch action {
					case "save":
						ed.Save()
						tryOpen()
					case "discard":
						tryOpen()
					case "cancel":
						// do nothing
					}
				})
		} else {
			tryOpen()
		}
		return nil
	}
	ed.Commands.Find("file.saveAs").Handler = func() error {
		sc.fileBrowser.Open(func(path string) {
			// Check if file already exists
			if _, err := os.Stat(path); err == nil {
				sc.dialog.Show("Overwrite?",
					"File \""+filepath.Base(path)+"\" already exists. Overwrite?",
					func(_ string, ok bool) {
						if ok {
							if err := ed.SaveFileAs(path); err != nil {
								sc.dialog.Show("Error", "Failed to save: "+err.Error(), nil)
							}
						}
					})
			} else {
				if err := ed.SaveFileAs(path); err != nil {
					sc.dialog.Show("Error", "Failed to save: "+err.Error(), nil)
				}
			}
		})
		return nil
	}
	ed.Commands.Find("file.new").Handler = func() error {
		if ed.Buffer.Modified() {
			sc.dialog.Show("Unsaved Changes", "Discard changes and create new file?",
				func(_ string, ok bool) {
					if ok {
						ed.Buffer = buffer.New()
						ed.Cursor = &buffer.Cursor{}
					}
				})
		} else {
			ed.Buffer = buffer.New()
			ed.Cursor = &buffer.Cursor{}
		}
		return nil
	}
	ed.Commands.Find("find.find").Handler = func() error {
		sc.searchBar.ToggleFind()
		return nil
	}
	ed.Commands.Find("find.next").Handler = func() error {
		if sc.searchBar.Query() != "" {
			sc.searchBar.ToggleFind()
			sc.searchBar.FindNext()
			sc.searchBar.ToggleFind()
		}
		return nil
	}
	ed.Commands.Find("find.replace").Handler = func() error {
		sc.searchBar.ToggleReplace()
		return nil
	}
	ed.Commands.Find("help.keyboard").Handler = func() error {
		sc.helpOverlay.Toggle()
		return nil
	}
	ed.Commands.Find("app.quit").Handler = func() error {
		if ed.Buffer.Modified() {
			sc.dialog.ShowThreeWay("Unsaved Changes",
				"File has unsaved changes. Save before quitting?",
				func(action string) {
					switch action {
					case "save":
						ed.Save()
						sc.quit = true
					case "discard":
						sc.quit = true
					case "cancel":
						// do nothing — stay in editor
					}
				})
		} else {
			sc.quit = true
		}
		return nil
	}
	return sc, nil
}

// Quit signals the event loop to exit.
func (sc *Screen) Quit() { sc.quit = true }

// Close restores the terminal.
func (sc *Screen) Close() { sc.tcell.Fini() }

// ensureCursorVisible adjusts scrollOffset to keep the cursor in view.
func (sc *Screen) ensureCursorVisible() {
	contentH := sc.ContentHeight()
	cursorY := sc.ed.Cursor.Y

	if cursorY < sc.scrollOffset {
		sc.scrollOffset = cursorY
	}
	if cursorY >= sc.scrollOffset+contentH {
		sc.scrollOffset = cursorY - contentH + 1
	}
}

// Run starts the event loop.
func (sc *Screen) Run() error {
	defer sc.Close()

	// Check minimum terminal size
	if sc.width < 40 || sc.height < 10 {
		sc.Close()
		return nil // terminal too small, just exit
	}

	for !sc.quit {
		sc.ensureCursorVisible()
		sc.Render()
		ev := sc.tcell.PollEvent()
		switch e := ev.(type) {
		case *tcell.EventKey:
			sc.handleKey(e)
		case *tcell.EventMouse:
			sc.handleMouse(e)
		case *tcell.EventResize:
			sc.width, sc.height = e.Size()
		}
	}
	return nil
}

// ContentHeight returns the number of lines available for editor content.
func (sc *Screen) ContentHeight() int {
	return sc.height - 2 // 1 for menu bar, 1 for status bar
}

// ContentStartY returns the y-coordinate where editor content begins.
func (sc *Screen) ContentStartY() int {
	return 1 // after menu bar
}

// handleKey processes keyboard events.
func (sc *Screen) handleKey(ev *tcell.EventKey) {
	// Route to file browser first if active
	if sc.fileBrowser.Active {
		sc.fileBrowser.HandleKey(ev)
		return
	}
	// Route to help overlay next if active
	if sc.helpOverlay.Active {
		sc.helpOverlay.HandleKey(ev)
		return
	}
	// Route to dialog next if active
	if sc.dialog.Active {
		sc.dialog.HandleKey(ev)
		return
	}
	// Route to search bar next if active
	if sc.searchBar.Active {
		sc.searchBar.HandleKey(ev)
		return
	}
	// Route to command palette next if active
	if sc.cmdPalette.Active {
		sc.cmdPalette.HandleKey(ev)
		return
	}

	ed := sc.ed
	buf := ed.Buffer

	// Esc clears selection when no modal is active
	if ev.Key() == tcell.KeyEscape {
		sc.ed.Selection.Clear()
	}

	// Navigation keys — Shift extends selection (use bitwise AND, some terminals report different modifier masks)
	shift := ev.Modifiers()&tcell.ModShift != 0

	switch ev.Key() {
	case tcell.KeyLeft:
		if shift && !sc.ed.Selection.Active() { sc.ed.Selection.Begin(ed.Cursor.X, ed.Cursor.Y) }
		if !shift { sc.ed.Selection.Clear() }
		ed.Cursor.MoveLeft(buf)
		if shift { sc.ed.Selection.Extend(ed.Cursor.X, ed.Cursor.Y) }
		return
	case tcell.KeyRight:
		if shift && !sc.ed.Selection.Active() { sc.ed.Selection.Begin(ed.Cursor.X, ed.Cursor.Y) }
		if !shift { sc.ed.Selection.Clear() }
		ed.Cursor.MoveRight(buf)
		if shift { sc.ed.Selection.Extend(ed.Cursor.X, ed.Cursor.Y) }
		return
	case tcell.KeyUp:
		if shift && !sc.ed.Selection.Active() { sc.ed.Selection.Begin(ed.Cursor.X, ed.Cursor.Y) }
		if !shift { sc.ed.Selection.Clear() }
		ed.Cursor.MoveUp(buf)
		if shift { sc.ed.Selection.Extend(ed.Cursor.X, ed.Cursor.Y) }
		return
	case tcell.KeyDown:
		if shift && !sc.ed.Selection.Active() { sc.ed.Selection.Begin(ed.Cursor.X, ed.Cursor.Y) }
		if !shift { sc.ed.Selection.Clear() }
		ed.Cursor.MoveDown(buf)
		if shift { sc.ed.Selection.Extend(ed.Cursor.X, ed.Cursor.Y) }
		return
	case tcell.KeyHome:
		if shift && !sc.ed.Selection.Active() { sc.ed.Selection.Begin(ed.Cursor.X, ed.Cursor.Y) }
		if !shift { sc.ed.Selection.Clear() }
		ed.Cursor.MoveToStartOfLine()
		if shift { sc.ed.Selection.Extend(ed.Cursor.X, ed.Cursor.Y) }
		return
	case tcell.KeyEnd:
		if shift && !sc.ed.Selection.Active() { sc.ed.Selection.Begin(ed.Cursor.X, ed.Cursor.Y) }
		if !shift { sc.ed.Selection.Clear() }
		ed.Cursor.MoveToEndOfLine(buf)
		if shift { sc.ed.Selection.Extend(ed.Cursor.X, ed.Cursor.Y) }
		return
	case tcell.KeyPgUp:
		if shift && !sc.ed.Selection.Active() { sc.ed.Selection.Begin(ed.Cursor.X, ed.Cursor.Y) }
		if !shift { sc.ed.Selection.Clear() }
		ed.Cursor.MovePageUp(buf, sc.ContentHeight())
		if shift { sc.ed.Selection.Extend(ed.Cursor.X, ed.Cursor.Y) }
		return
	case tcell.KeyPgDn:
		if shift && !sc.ed.Selection.Active() { sc.ed.Selection.Begin(ed.Cursor.X, ed.Cursor.Y) }
		if !shift { sc.ed.Selection.Clear() }
		ed.Cursor.MovePageDown(buf, sc.ContentHeight())
		if shift { sc.ed.Selection.Extend(ed.Cursor.X, ed.Cursor.Y) }
		return
	}

	// View mode: Space = PgDn, Shift+Space = PgUp
	if ed.Mode == editor.ModeView && (ev.Key() == tcell.KeyRune && ev.Rune() == ' ') {
		if shift {
			ed.Cursor.MovePageUp(buf, sc.ContentHeight())
		} else {
			ed.Cursor.MovePageDown(buf, sc.ContentHeight())
		}
		return
	}

	// Global shortcuts
	switch ev.Key() {
	case tcell.KeyCtrlQ:
		ed.Commands.Find("app.quit").Handler()
		return
	case tcell.KeyCtrlE:
		ed.ToggleMode()
		return
	case tcell.KeyCtrlO:
		ed.Commands.Find("file.open").Handler()
		return
	case tcell.KeyCtrlZ:
		ed.Commands.Find("edit.undo").Handler()
		return
	case tcell.KeyCtrlF:
		sc.searchBar.ToggleFind()
		return
	case tcell.KeyF3:
		if sc.searchBar.Active {
			sc.searchBar.FindNext()
		} else if sc.searchBar.FindNextFromLast() {
			// found match, search bar stays closed
		} else {
			sc.searchBar.ToggleFind()
		}
		return
	case tcell.KeyCtrlH:
		sc.searchBar.ToggleReplace()
		return
	case tcell.KeyCtrlP:
		sc.cmdPalette.Toggle()
		return
	case tcell.KeyCtrlC:
		ed.Commands.Find("edit.copy").Handler()
		return
	case tcell.KeyCtrlX:
		ed.Commands.Find("edit.cut").Handler()
		return
	case tcell.KeyCtrlV:
		ed.Commands.Find("edit.paste").Handler()
		return
	case tcell.KeyCtrlA:
		ed.Commands.Find("edit.selectAll").Handler()
		return
	case tcell.KeyCtrlS:
		ed.Save()
		return
	}

	// Edit mode keys
	if ed.Mode == editor.ModeEdit {
		switch ev.Key() {
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			ed.SaveSnapshot()
			x, y := buf.DeleteBackward(ed.Cursor.X, ed.Cursor.Y)
			ed.Cursor.X, ed.Cursor.Y = x, y
			return
		case tcell.KeyDelete:
			ed.SaveSnapshot()
			buf.DeleteForward(ed.Cursor.X, ed.Cursor.Y)
			return
		case tcell.KeyEnter:
			ed.SaveSnapshot()
			x, y := buf.NewLine(ed.Cursor.X, ed.Cursor.Y)
			ed.Cursor.X, ed.Cursor.Y = x, y
			return
		case tcell.KeyTab:
			for i := 0; i < buf.TabWidth(); i++ {
				buf.Insert(' ', ed.Cursor.X, ed.Cursor.Y)
				ed.Cursor.X++
			}
			return
		case tcell.KeyRune:
			ed.SaveSnapshot()
			buf.Insert(ev.Rune(), ed.Cursor.X, ed.Cursor.Y)
			ed.Cursor.X++
			return
		}
	}
}

// screenToBuffer converts screen coordinates to buffer (rune-index) coordinates.
// Uses runewidth to account for wide characters (CJK, emoji, etc).
func (sc *Screen) screenToBuffer(sx, sy int) (bx, by int) {
	startY := sc.ContentStartY()
	by = sy - startY + sc.scrollOffset
	if by < 0 {
		by = 0
	}
	if by >= sc.ed.Buffer.Lines() {
		by = sc.ed.Buffer.Lines() - 1
	}

	// Subtract line number column width
	col := sx
	if sc.ed.ShowLineNum {
		n := sc.ed.Buffer.Lines()
		wn := 0
		for n > 0 {
			n /= 10
			wn++
		}
		if wn < 2 {
			wn = 2
		}
		col -= wn + 2
	}
	if col < 0 {
		col = 0
	}

	// Walk runes in the line accumulating visual width
	// NOTE: for range over string yields byte index, so use a separate rune counter
	line := sc.ed.Buffer.Line(by)
	acc := 0
	runeIdx := 0
	for _, ch := range line {
		w := runewidth.RuneWidth(ch)
		if col >= acc && col < acc+w {
			bx = runeIdx
			return
		}
		acc += w
		runeIdx++
	}
	// Past end of line
	bx = runeIdx
	return
}

// handleMouse processes mouse events.
func (sc *Screen) handleMouse(ev *tcell.EventMouse) {
	x, y := ev.Position()
	btn := ev.Buttons()

	switch {
	case btn&tcell.WheelUp != 0:
		if sc.scrollOffset > 0 {
			sc.scrollOffset--
		}
	case btn&tcell.WheelDown != 0:
		if sc.scrollOffset < sc.ed.Buffer.Lines()-sc.ContentHeight() {
			sc.scrollOffset++
		}
	case btn&tcell.Button1 != 0:
		// Mouse in edit area
		startY := sc.ContentStartY()
		if y < startY {
			return
		}

		bx, by := sc.screenToBuffer(x, y)

		if sc.dragging {
			// Extend selection
			sc.ed.Selection.Extend(bx, by)
		} else {
			// Start new selection or click
			sc.ed.Selection.Begin(bx, by)
			sc.dragging = true
		}
		sc.ed.Cursor.X = bx
		sc.ed.Cursor.Y = by

	case btn == tcell.ButtonNone:
		// Mouse release — end drag selection
		if sc.dragging {
			sc.dragging = false
		}
	}
}
