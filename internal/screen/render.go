package screen

import (
	"fmt"
)

// Render draws the entire editor UI.
func (sc *Screen) Render() {
	sc.tcell.Clear()
	ed := sc.ed
	buf := ed.Buffer

	// --- 1. Menu bar ---
	sc.menuBar.Render(sc.tcell, 0, sc.width, sc.palette.MenuBar, sc.palette.MenuSel)

	// --- 2. Editor content ---
	numWidth := 0
	if ed.ShowLineNum {
		n := buf.Lines()
		for n > 0 {
			n /= 10
			numWidth++
		}
		if numWidth < 2 {
			numWidth = 2
		}
	}

	contentH := sc.ContentHeight()
	startY := sc.ContentStartY()

	for y := 0; y < contentH && y+sc.scrollOffset < buf.Lines(); y++ {
		actualY := y + sc.scrollOffset
		line := buf.Line(actualY)

		// Line number
		col := 0
		if ed.ShowLineNum {
			numStr := fmt.Sprintf("%*d ", numWidth, actualY+1)
			for i, ch := range numStr {
				sc.tcell.SetCell(col+i, startY+y, sc.palette.LineNum, ch)
			}
			col += len([]rune(numStr))
		}

		// Separator after line numbers
		if ed.ShowLineNum {
			sc.tcell.SetCell(col, startY+y, sc.palette.LineNum, '│')
			col++
		}

		// Line content
		for i, ch := range line {
			style := sc.palette.Default
			// Selection highlighting (proper coords will be handled in Task 9)
			if ed.Selection.Active() && ed.Selection.Contains(ed.Cursor.X, ed.Cursor.Y) {
				// Placeholder — selection rendering refined later
			}
			if col+i < sc.width {
				sc.tcell.SetCell(col+i, startY+y, style, ch)
			}
		}

		// Fill rest of line
		for x := col + len([]rune(line)); x < sc.width; x++ {
			sc.tcell.SetCell(x, startY+y, sc.palette.Default, ' ')
		}
	}

	// --- 3. Status bar ---
	sc.statusBar.Render(sc.tcell, sc.height-1, sc.width, sc.palette.StatusBar)

	// --- 4. Search bar overlay ---
	sc.searchBar.Render(sc.tcell, sc.width, sc.height)

	// --- 5. Dialog overlay ---
	sc.dialog.Render(sc.tcell, sc.width, sc.height)

	// --- 6. Command palette overlay ---
	sc.cmdPalette.Render(sc.tcell, sc.width, sc.height)

	// --- 6. Cursor ---
	cx := ed.Cursor.X
	cy := ed.Cursor.Y - sc.scrollOffset + startY
	if ed.ShowLineNum {
		// Calculate line number width
		n := buf.Lines()
		wn := 0
		for n > 0 {
			n /= 10
			wn++
		}
		if wn < 2 {
			wn = 2
		}
		cx += wn + 2 // number width + separator '│' + space
	}

	if cy >= startY && cy < startY+contentH {
		sc.tcell.ShowCursor(cx, cy)
	} else {
		sc.tcell.HideCursor()
	}

	sc.tcell.Sync()
}


