package screen

import (
	"fmt"

	"github.com/mattn/go-runewidth"
	"github.com/zsying/oe/internal/editor"
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
	totalRows := len(sc.vm)

	for y := 0; y < contentH; y++ {
		row := y + sc.scrollOffset
		if row >= totalRows {
			break
		}
		v := sc.vm[row]
		line := buf.Line(v.line)

		// Line number gutter (shown only on the first wrapped segment of a line)
		col := 0
		if ed.ShowLineNum {
			if v.start == 0 {
				numStr := fmt.Sprintf("%*d ", numWidth, v.line+1)
				for i, ch := range numStr {
					sc.tcell.SetCell(col+i, startY+y, sc.palette.LineNum, ch)
				}
				col += runewidth.StringWidth(numStr)
			} else {
				// Continuation row: blank gutter aligned with the line-number area
				for i := 0; i < numWidth+1; i++ {
					sc.tcell.SetCell(col+i, startY+y, sc.palette.LineNum, ' ')
				}
				col += numWidth + 1
			}
			sc.tcell.SetCell(col, startY+y, sc.palette.LineNum, '│')
			col++
		}

		// Line content within this visual row — selection aware, runewidth aware
		baseStyle := sc.palette.Default
		if ed.Mode == editor.ModeView {
			baseStyle = sc.palette.ViewDim
		}
		runeIdx := 0
		for _, ch := range line {
			if runeIdx < v.start {
				runeIdx++
				continue
			}
			if runeIdx >= v.end {
				break
			}
			style := baseStyle
			if ed.Selection.Active() && ed.Selection.Contains(runeIdx, v.line) {
				style = sc.palette.Selection
			}
			w := runewidth.RuneWidth(ch)
			if col+w <= sc.width {
				sc.tcell.SetCell(col, startY+y, style, ch)
			}
			col += w
			runeIdx++
		}

		// Fill rest of the row with background
		for col < sc.width {
			sc.tcell.SetCell(col, startY+y, sc.palette.Default, ' ')
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

	// --- 6. Cursor — position within the soft-wrapped visual row ---
	cvr := sc.cursorVisualRow()
	cy := cvr - sc.scrollOffset + startY

	// Visual column = width of runes from the visual row's start up to the cursor.
	var vr visualRow
	if cvr >= 0 && cvr < len(sc.vm) {
		vr = sc.vm[cvr]
	}
	line := buf.Line(vr.line)
	cx := 0
	runeIdx := 0
	for _, ch := range line {
		if runeIdx >= vr.start && runeIdx < ed.Cursor.X {
			cx += runewidth.RuneWidth(ch)
		}
		runeIdx++
	}
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

	// Hide cursor when any modal overlay is active
	modalActive := sc.fileBrowser.Active || sc.dialog.Active ||
		sc.cmdPalette.Active || sc.searchBar.Active || sc.helpOverlay.Active
	if modalActive {
		sc.tcell.HideCursor()
	} else if cy >= startY && cy < startY+contentH {
		sc.tcell.ShowCursor(cx, cy)
	} else {
		sc.tcell.HideCursor()
	}

	sc.tcell.Sync()
}
