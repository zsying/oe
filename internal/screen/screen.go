package screen

import (
	"github.com/gdamore/tcell/v2"
	"github.com/user/editor/internal/buffer"
	"github.com/user/editor/internal/editor"
	"github.com/user/editor/internal/widgets"
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
	menuBar      *widgets.MenuBar
	cmdPalette   *widgets.CommandPalette
	searchBar    *widgets.SearchBar
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
	sc.menuBar = widgets.NewMenuBar(ed)
	sc.cmdPalette = widgets.NewCommandPalette(ed)
	sc.searchBar = widgets.NewSearchBar(ed)
	sc.dialog = widgets.NewDialog()

	// Override dialog-requiring command handlers
	ed.Commands.Find("file.open").Handler = func() error {
		sc.dialog.ShowInput("Open File", "Enter file path:", "", func(path string, ok bool) {
			if ok && path != "" {
				ed.OpenFile(path)
			}
		})
		return nil
	}
	ed.Commands.Find("file.saveAs").Handler = func() error {
		sc.dialog.ShowInput("Save As", "Enter file path:", ed.Buffer.Filename(), func(path string, ok bool) {
			if ok && path != "" {
				ed.SaveFileAs(path)
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
	ed.Commands.Find("app.quit").Handler = func() error {
		if ed.Buffer.Modified() {
			sc.dialog.Show("Unsaved Changes", "Save changes before quitting?",
				func(_ string, ok bool) {
					if ok {
						ed.Save()
					}
					sc.quit = true
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
	// Route to dialog first if active
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

	// Navigation keys (both modes)
	switch ev.Key() {
	case tcell.KeyLeft:
		ed.Cursor.MoveLeft(buf)
		return
	case tcell.KeyRight:
		ed.Cursor.MoveRight(buf)
		return
	case tcell.KeyUp:
		ed.Cursor.MoveUp(buf)
		return
	case tcell.KeyDown:
		ed.Cursor.MoveDown(buf)
		return
	case tcell.KeyHome:
		ed.Cursor.MoveToStartOfLine()
		return
	case tcell.KeyEnd:
		ed.Cursor.MoveToEndOfLine(buf)
		return
	case tcell.KeyPgUp:
		ed.Cursor.MovePageUp(buf, sc.ContentHeight())
		return
	case tcell.KeyPgDn:
		ed.Cursor.MovePageDown(buf, sc.ContentHeight())
		return
	}

	// Close menu on Escape
	if ev.Key() == tcell.KeyEscape {
		sc.menuBar.Close()
	}

	// Global shortcuts
	switch ev.Key() {
	case tcell.KeyCtrlQ:
		sc.quit = true
		return
	case tcell.KeyCtrlE:
		ed.ToggleMode()
		return
	case tcell.KeyCtrlF:
		sc.searchBar.ToggleFind()
		return
	case tcell.KeyF3:
		if sc.searchBar.Query() != "" {
			sc.searchBar.ToggleFind()
			sc.searchBar.FindNext()
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

// screenToBuffer converts screen coordinates to buffer coordinates.
func (sc *Screen) screenToBuffer(sx, sy int) (bx, by int) {
	startY := sc.ContentStartY()
	by = sy - startY + sc.scrollOffset
	if by < 0 {
		by = 0
	}
	if by >= sc.ed.Buffer.Lines() {
		by = sc.ed.Buffer.Lines() - 1
	}

	bx = sx
	numWidth := 0
	if sc.ed.ShowLineNum {
		n := sc.ed.Buffer.Lines()
		for n > 0 {
			n /= 10
			numWidth++
		}
		if numWidth < 2 {
			numWidth = 2
		}
		bx -= numWidth + 2
	}
	if bx < 0 {
		bx = 0
	}
	if bx > sc.ed.Buffer.LineLen(by) {
		bx = sc.ed.Buffer.LineLen(by)
	}
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
		// Check menu bar click first
		if cmdID, handled := sc.menuBar.HandleMouseClick(x, y); handled {
			if cmdID != "" {
				if cmd := sc.ed.Commands.Find(cmdID); cmd != nil {
					cmd.Handler()
				}
			}
			return
		}
		// Check submenu click
		if sc.menuBar.IsOpen() {
			if cmdID, handled := sc.menuBar.HandleSubmenuClick(x, y); handled {
				if cmdID != "" {
					if cmd := sc.ed.Commands.Find(cmdID); cmd != nil {
						cmd.Handler()
					}
				}
				return
			}
		}

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
