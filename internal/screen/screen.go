package screen

import (
	"os"
	"path/filepath"
	"runtime"

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
	scrollOffset int // index of the first visible visual row (soft-wrapped)
	vm           []visualRow
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

	// On Unix, also route clipboard writes through the OSC 52 terminal escape
	// sequence so copied text reaches the system clipboard reliably (works over
	// SSH and inside tmux, independent of xclip/xsel/wl-clipboard).
	if runtime.GOOS != "windows" {
		ed.Clip.SetOSC52Writer(os.Stdout)
	}

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

// ensureCursorVisible adjusts scrollOffset (in visual rows) to keep the
// cursor's visual row within the visible content area.
func (sc *Screen) ensureCursorVisible() {
	contentH := sc.ContentHeight()
	cvr := sc.cursorVisualRow()
	if cvr < 0 {
		return
	}
	if cvr < sc.scrollOffset {
		sc.scrollOffset = cvr
	}
	if cvr >= sc.scrollOffset+contentH {
		sc.scrollOffset = cvr - contentH + 1
	}
	if sc.scrollOffset < 0 {
		sc.scrollOffset = 0
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
		sc.vm = sc.buildWrapMap()
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

	// Navigation keys — Shift extends selection, Ctrl moves by word
	mods := ev.Modifiers()
	shift := mods&tcell.ModShift != 0
	ctrl := mods&tcell.ModCtrl != 0

	switch ev.Key() {
	case tcell.KeyLeft:
		sc.moveWithSelection(shift, func() {
			if ctrl {
				ed.Cursor.MoveWordLeft(buf)
			} else {
				ed.Cursor.MoveLeft(buf)
			}
		})
		return
	case tcell.KeyRight:
		sc.moveWithSelection(shift, func() {
			if ctrl {
				ed.Cursor.MoveWordRight(buf)
			} else {
				ed.Cursor.MoveRight(buf)
			}
		})
		return
	case tcell.KeyUp:
		sc.moveWithSelection(shift, func() { ed.Cursor.MoveUp(buf) })
		return
	case tcell.KeyDown:
		sc.moveWithSelection(shift, func() { ed.Cursor.MoveDown(buf) })
		return
	case tcell.KeyHome:
		if shift {
			sc.ed.Selection.Begin(ed.Cursor.X, ed.Cursor.Y)
		}
		ed.Cursor.MoveToStartOfLine()
		if shift {
			sc.ed.Selection.Extend(ed.Cursor.X, ed.Cursor.Y)
		}
		return
	case tcell.KeyEnd:
		if shift {
			sc.ed.Selection.Begin(ed.Cursor.X, ed.Cursor.Y)
		}
		ed.Cursor.MoveToEndOfLine(buf)
		if shift {
			sc.ed.Selection.Extend(ed.Cursor.X, ed.Cursor.Y)
		}
		return
	case tcell.KeyPgUp:
		if shift {
			sc.ed.Selection.Begin(ed.Cursor.X, ed.Cursor.Y)
		}
		ed.Cursor.MovePageUp(buf, sc.ContentHeight())
		if shift {
			sc.ed.Selection.Extend(ed.Cursor.X, ed.Cursor.Y)
		}
		return
	case tcell.KeyPgDn:
		if shift {
			sc.ed.Selection.Begin(ed.Cursor.X, ed.Cursor.Y)
		}
		ed.Cursor.MovePageDown(buf, sc.ContentHeight())
		if shift {
			sc.ed.Selection.Extend(ed.Cursor.X, ed.Cursor.Y)
		}
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
	case tcell.KeyCtrlY:
		ed.Commands.Find("edit.redo").Handler()
		return
	case tcell.KeyCtrlZ:
		if shift {
			ed.Commands.Find("edit.redo").Handler()
		} else {
			ed.Commands.Find("edit.undo").Handler()
		}
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
			// With a selection, Backspace deletes the whole selection.
			if ed.DeleteSelection() {
				return
			}
			// No-op at the very start of the buffer: don't record a snapshot.
			if ed.Cursor.X == 0 && ed.Cursor.Y == 0 {
				return
			}
			ed.SaveSnapshot(editor.OpGeneric)
			x, y := buf.DeleteBackward(ed.Cursor.X, ed.Cursor.Y)
			ed.Cursor.X, ed.Cursor.Y = x, y
			return
		case tcell.KeyDelete:
			// With a selection, Delete removes the whole selection.
			if ed.DeleteSelection() {
				return
			}
			// No-op at the very end of the buffer: don't record a snapshot.
			if ed.Cursor.X >= buf.LineLen(ed.Cursor.Y) && ed.Cursor.Y >= buf.Lines()-1 {
				return
			}
			ed.SaveSnapshot(editor.OpGeneric)
			buf.DeleteForward(ed.Cursor.X, ed.Cursor.Y)
			return
		case tcell.KeyEnter:
			replaced := ed.DeleteSelection()
			if !replaced {
				ed.SaveSnapshot(editor.OpGeneric)
			}
			x, y := buf.NewLine(ed.Cursor.X, ed.Cursor.Y)
			ed.Cursor.X, ed.Cursor.Y = x, y
			return
		case tcell.KeyTab:
			replaced := ed.DeleteSelection()
			if !replaced {
				ed.SaveSnapshot(editor.OpGeneric)
			}
			for i := 0; i < buf.TabWidth(); i++ {
				buf.Insert(' ', ed.Cursor.X, ed.Cursor.Y)
				ed.Cursor.X++
			}
			return
		case tcell.KeyRune:
			// Typing over a selection replaces it with the typed character.
			replaced := ed.DeleteSelection()
			if !replaced {
				ed.SaveSnapshot(editor.OpInsert)
			}
			buf.Insert(ev.Rune(), ed.Cursor.X, ed.Cursor.Y)
			ed.Cursor.X++
			return
		}
	}
}

// moveWithSelection applies a cursor move while maintaining a Shift-extended
// selection. It begins the selection only on the FIRST Shift press so the anchor
// stays fixed (calling Begin on every step would make the anchor follow the
// cursor and the previously selected range would be lost). When moving without
// Shift, any active selection is cleared.
func (sc *Screen) moveWithSelection(shift bool, move func()) {
	if shift {
		if !sc.ed.Selection.Active() {
			sc.ed.Selection.Begin(sc.ed.Cursor.X, sc.ed.Cursor.Y)
		}
		move()
		sc.ed.Selection.Extend(sc.ed.Cursor.X, sc.ed.Cursor.Y)
	} else {
		sc.ed.Selection.Clear()
		move()
	}
}

// screenToBuffer converts screen coordinates to buffer (rune-index) coordinates.
// Uses runewidth to account for wide characters (CJK, emoji, etc), and maps the
// screen row through the soft-wrap map so wrapped visual rows resolve correctly.
func (sc *Screen) screenToBuffer(sx, sy int) (bx, by int) {
	startY := sc.ContentStartY()
	vr := sy - startY + sc.scrollOffset
	if vr < 0 {
		vr = 0
	}
	if len(sc.vm) == 0 {
		by = 0
		bx = 0
		return
	}
	if vr >= len(sc.vm) {
		vr = len(sc.vm) - 1
	}
	vrow := sc.vm[vr]
	by = vrow.line

	// Subtract line number column width
	col := sx - sc.gutterWidth()
	if col < 0 {
		col = 0
	}

	// Walk runes within this visual row's segment, accumulating visual width.
	line := sc.ed.Buffer.Line(by)
	acc := 0
	ri := 0
	for _, ch := range line {
		if ri < vrow.start {
			ri++
			continue
		}
		if ri >= vrow.end {
			break
		}
		w := runewidth.RuneWidth(ch)
		if col >= acc && col < acc+w {
			bx = ri
			return
		}
		acc += w
		ri++
	}
	// Past end of this segment (click in the trailing gutter of the row).
	bx = vrow.end
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
		if len(sc.vm) > sc.ContentHeight() && sc.scrollOffset < len(sc.vm)-sc.ContentHeight() {
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
