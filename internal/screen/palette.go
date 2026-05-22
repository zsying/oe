package screen

import "github.com/gdamore/tcell/v2"

// Palette defines the color scheme for the editor.
type Palette struct {
	Default   tcell.Style
	ViewDim   tcell.Style // dimmed text in View mode
	MenuBar   tcell.Style
	MenuSel   tcell.Style
	StatusBar tcell.Style
	LineNum   tcell.Style
	Cursor    tcell.Style
	Selection tcell.Style
	Search    tcell.Style
	Modal     tcell.Style
}

// DefaultPalette returns the default color palette.
func DefaultPalette() Palette {
	return Palette{
		Default:   tcell.StyleDefault,
		ViewDim:   tcell.StyleDefault.Dim(true),
		MenuBar:   tcell.StyleDefault.Reverse(true),
		MenuSel:   tcell.StyleDefault.Reverse(true).Foreground(tcell.ColorWhite),
		StatusBar: tcell.StyleDefault.Reverse(true),
		LineNum:   tcell.StyleDefault.Foreground(tcell.ColorGray),
		Cursor:    tcell.StyleDefault.Reverse(true),
		Selection: tcell.StyleDefault.Background(tcell.ColorCornflowerBlue).Foreground(tcell.ColorWhite),
		Search:    tcell.StyleDefault.Background(tcell.ColorYellow).Foreground(tcell.ColorBlack),
		Modal:     tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack),
	}
}
