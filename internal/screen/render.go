package screen

import (
	"fmt"

	"github.com/mattn/go-runewidth"
)

// Render draws the entire editor UI.
func (sc *Screen) Render() {
	sc.tcell.Clear()
	ed := sc.ed
	buf := ed.Buffer

	// --- 1. Editor content ---
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
			col += runewidth.StringWidth(numStr)
		}

		// Separator after line numbers
		if ed.ShowLineNum {
			sc.tcell.SetCell(col, startY+y, sc.palette.LineNum, '│')
			col++
		}

		// Line content — use runewidth to account for wide chars
		for _, ch := range line {
			style := sc.palette.Default
			w := runewidth.RuneWidth(ch)
			if col+w <= sc.width {
				sc.tcell.SetCell(col, startY+y, style, ch)
			}
			col += w
		}

		// Fill rest of line
		for col < sc.width {
			if col < sc.width {
				sc.tcell.SetCell(col, startY+y, sc.palette.Default, ' ')
			}
			col++
		}
	}

	// --- 2. Status bar ---
	sc.statusBar.Render(sc.tcell, sc.height-1, sc.width, sc.palette.StatusBar)

	// --- 3. Search bar overlay ---
	sc.searchBar.Render(sc.tcell, sc.width, sc.height)

	// --- 4. File browser overlay ---
	sc.fileBrowser.Render(sc.tcell, sc.width, sc.height)

	// --- 5. Dialog overlay ---
	sc.dialog.Render(sc.tcell, sc.width, sc.height)

	// --- 6. Help overlay ---
	sc.helpOverlay.Render(sc.tcell, sc.width, sc.height)

	// --- 7. Command palette overlay ---
	sc.cmdPalette.Render(sc.tcell, sc.width, sc.height)

	// --- 6. Cursor — calculate visual column using runewidth ---
	line := buf.Line(ed.Cursor.Y)
	cx := 0
	runeIdx := 0
	for _, ch := range line {
		if runeIdx >= ed.Cursor.X {
			break
		}
		cx += runewidth.RuneWidth(ch)
		runeIdx++
	}
	cy := ed.Cursor.Y - sc.scrollOffset + startY
	if ed.ShowLineNum {
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
