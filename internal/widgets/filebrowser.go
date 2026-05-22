package widgets

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/zsying/oe/internal/editor"
)

const maxFileResults = 12

// FileBrowser implements a command-palette-style file open dialog
// with two focus modes: input box and file list.
type FileBrowser struct {
	Ed           *editor.Editor
	Active       bool
	query        []rune
	cwd          string
	entries      []fileEntry
	selIdx       int
	scrollOffset int
	focusOnList  bool // false = focus on input field
	callback     func(path string)
}

type fileEntry struct {
	name     string
	isDir    bool
	fullPath string
	upDir    bool // true if this is the [up-dir] entry
}

// NewFileBrowser creates a new file browser.
func NewFileBrowser(ed *editor.Editor) *FileBrowser {
	cwd, _ := os.Getwd()
	return &FileBrowser{
		Ed:  ed,
		cwd: cwd,
	}
}

// Open shows the file browser. callback is called with the chosen path.
func (fb *FileBrowser) Open(callback func(path string)) {
	fb.callback = callback
	fb.query = nil
	fb.selIdx = 0
	fb.scrollOffset = 0
	fb.focusOnList = false // start with input focus
	fb.refresh()
	fb.Active = true
}

func (fb *FileBrowser) refresh() {
	q := string(fb.query)

	if q == "" {
		fb.listDir(fb.cwd, "")
		fb.prependUpDir()
		return
	}

	// Check if the path contains a directory separator
	if strings.Contains(q, string(filepath.Separator)) || strings.Contains(q, "/") {
		dir := filepath.Dir(q)
		base := filepath.Base(q)
		absDir := fb.cwd
		if filepath.IsAbs(dir) {
			absDir = dir
		} else {
			absDir = filepath.Join(fb.cwd, dir)
		}
		fb.listDir(absDir, base)
	} else {
		fb.listDir(fb.cwd, q)
	}

	fb.prependUpDir()
}

func (fb *FileBrowser) prependUpDir() {
	parent := filepath.Dir(fb.cwd)
	if fb.cwd == parent {
		// At filesystem root — on Windows, show available drives
		if runtime.GOOS == "windows" {
			drives := listAvailableDrives()
			if len(drives) > 0 {
				fb.entries = append(drives, fb.entries...)
			}
		}
		return
	}
	// Normal parent directory
	fb.entries = append([]fileEntry{
		{name: "[up-dir]", isDir: true, fullPath: parent, upDir: true},
	}, fb.entries...)
}

// listAvailableDrives returns a list of available drive roots on Windows.
func listAvailableDrives() []fileEntry {
	if runtime.GOOS != "windows" {
		return nil
	}
	var drives []fileEntry
	for _, letter := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		root := string(letter) + ":\\"
		// Check if drive exists (just try to read it)
		if _, err := os.ReadDir(root); err == nil {
			drives = append(drives, fileEntry{
				name:     string(letter) + ":",
				isDir:    true,
				fullPath: root,
			})
		}
	}
	return drives
}

func (fb *FileBrowser) listDir(dir, filter string) {
	// On Windows, when at a drive root, list available drives
	if runtime.GOOS == "windows" && len(dir) == 3 && dir[1] == ':' && dir[2] == '\\' {
		// Single drive root like C:\ - list normally, no special handling needed
	} else if runtime.GOOS == "windows" && fb.cwd == dir && filepath.Dir(dir) == dir {
		// At virtual root - show drive list. This won't normally trigger since
		// drives are individual roots. But we handle it here.
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		// If can't read (e.g., empty drive), provide empty list
		if runtime.GOOS == "windows" && len(dir) >= 2 && dir[1] == ':' {
			fb.entries = listAvailableDrives()
			// Filter by filter letter
			if filter != "" {
				var filtered []fileEntry
				for _, d := range fb.entries {
					if fuzzyMatchFile(d.name, strings.ToLower(filter)) {
						filtered = append(filtered, d)
					}
				}
				fb.entries = filtered
			}
			return
		}
		fb.entries = nil
		return
	}

	filterLower := strings.ToLower(filter)

	var matched []fileEntry
	for _, entry := range entries {
		name := entry.Name()
		// Skip hidden files unless explicitly filtered
		if strings.HasPrefix(name, ".") && !strings.HasPrefix(filter, ".") {
			continue
		}
		if filter == "" || fuzzyMatchFile(name, filterLower) {
			matched = append(matched, fileEntry{
				name:     name,
				isDir:    entry.IsDir(),
				fullPath: filepath.Join(dir, name),
			})
		}
	}

	// Sort: directories first, then alphabetical
	sort.Slice(matched, func(i, j int) bool {
		if matched[i].isDir != matched[j].isDir {
			return matched[i].isDir
		}
		return strings.ToLower(matched[i].name) < strings.ToLower(matched[j].name)
	})

	fb.entries = matched
}

// fuzzyMatchFile matches by substring (case-insensitive).
func fuzzyMatchFile(text, query string) bool {
	return strings.Contains(strings.ToLower(text), query)
}

// resolvePath expands ~ to home, resolves relative paths, and cleans the result.
func resolvePath(cwd, input string) string {
	if input == "" {
		return cwd
	}
	// Expand ~ to home directory
	if strings.HasPrefix(input, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			if input == "~" {
				return home
			}
			input = filepath.Join(home, input[1:])
		}
	}
	// Handle absolute paths
	if filepath.IsAbs(input) {
		return filepath.Clean(input)
	}
	// Handle Windows drive letters (e.g., "d:/path" or "d:" or "c:")
	if len(input) >= 2 && input[1] == ':' {
		if len(input) == 2 {
			input = input + string(filepath.Separator) // "d:" → "d:\"
		}
		return filepath.Clean(input)
	}
	// Relative path: join with cwd
	return filepath.Clean(filepath.Join(cwd, input))
}

// HandleKey processes key events with dual-focus logic.
func (fb *FileBrowser) HandleKey(ev *tcell.EventKey) bool {
	if !fb.Active {
		return false
	}

	// Esc always cancels
	if ev.Key() == tcell.KeyEscape {
		fb.Active = false
		return true
	}

	if fb.focusOnList {
		return fb.handleKeyList(ev)
	}
	return fb.handleKeyInput(ev)
}

// handleKeyInput: focus is on the text input field.
func (fb *FileBrowser) handleKeyInput(ev *tcell.EventKey) bool {
	switch ev.Key() {
	case tcell.KeyEnter:
		q := strings.TrimSpace(string(fb.query))
		if q == "" {
			return true
		}
		path := resolvePath(fb.cwd, q)
		// Check if path is a directory — navigate into it
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			fb.cwd = path
			fb.query = nil
			fb.selIdx = 0
			fb.scrollOffset = 0
			fb.refresh()
			return true
		}
		fb.Active = false
		if fb.callback != nil {
			fb.callback(path)
		}
		return true

	case tcell.KeyUp, tcell.KeyDown:
		// Switch focus to file list (select first real entry, skipping [up-dir])
		fb.focusOnList = true
		if len(fb.entries) > 0 {
			fb.selIdx = 0
			fb.scrollOffset = 0
		}
		return true

	case tcell.KeyTab:
		// Switch focus to file list
		fb.focusOnList = true
		if len(fb.entries) > 0 {
			fb.selIdx = 0
			fb.scrollOffset = 0
		}
		return true

	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(fb.query) > 0 {
			fb.query = fb.query[:len(fb.query)-1]
			fb.refresh()
		}
		return true

	case tcell.KeyRune:
		fb.query = append(fb.query, ev.Rune())
		fb.refresh()
		return true
	}
	return false
}

// handleKeyList: focus is on the file list.
func (fb *FileBrowser) handleKeyList(ev *tcell.EventKey) bool {
	n := len(fb.entries)
	if n == 0 {
		return true
	}

	switch ev.Key() {
	case tcell.KeyEnter:
		entry := fb.entries[fb.selIdx]
		if entry.upDir || entry.isDir {
			// Navigate into directory
			fb.cwd = entry.fullPath
			fb.query = nil
			fb.selIdx = 0
			fb.scrollOffset = 0
			fb.focusOnList = false // back to input after navigating
			fb.refresh()
		} else {
			// Select file
			fb.Active = false
			if fb.callback != nil {
				fb.callback(entry.fullPath)
			}
		}
		return true

	case tcell.KeyUp:
		if fb.selIdx > 0 {
			fb.selIdx--
		} else {
			// At top — switch focus back to input
			fb.focusOnList = false
			return true
		}
		if fb.selIdx < fb.scrollOffset {
			fb.scrollOffset = fb.selIdx
		}
		return true

	case tcell.KeyDown:
		if fb.selIdx < n-1 {
			fb.selIdx++
		} else {
			// At bottom — switch focus back to input
			fb.focusOnList = false
			return true
		}
		if fb.selIdx >= fb.scrollOffset+maxFileResults {
			fb.scrollOffset = fb.selIdx - maxFileResults + 1
		}
		return true

	case tcell.KeyTab:
		// Switch focus back to input
		fb.focusOnList = false
		return true

	case tcell.KeyRune:
		// Typing switches focus to input and appends the character
		fb.focusOnList = false
		fb.query = append(fb.query, ev.Rune())
		fb.refresh()
		return true

	case tcell.KeyBackspace, tcell.KeyBackspace2:
		// Backspace in list mode: navigate to parent directory
		parent := filepath.Dir(fb.cwd)
		if parent != fb.cwd {
			fb.cwd = parent
			fb.query = nil
			fb.selIdx = 0
			fb.scrollOffset = 0
			fb.focusOnList = false
			fb.refresh()
		}
		return true
	}
	return false
}

// Render draws the file browser overlay.
func (fb *FileBrowser) Render(s tcell.Screen, w, h int) {
	if !fb.Active {
		return
	}

	listLen := len(fb.entries)
	visCount := maxFileResults
	if visCount > listLen {
		visCount = listLen
	}
	totalH := visCount + 4
	if totalH < 5 {
		totalH = 5
	}
	if totalH > h-2 {
		totalH = h - 2
	}

	const fw = 55
	fx := (w - fw) / 2
	fy := h / 4

	bgStyle := tcell.StyleDefault.
		Background(tcell.ColorWhite).
		Foreground(tcell.ColorBlack)
	dirStyle := bgStyle.Foreground(tcell.ColorDodgerBlue)
	selDirStyle := bgStyle.
		Background(tcell.ColorCornflowerBlue).
		Foreground(tcell.ColorYellow)
	selStyle := bgStyle.
		Background(tcell.ColorCornflowerBlue).
		Foreground(tcell.ColorWhite)
	pathStyle := bgStyle.Foreground(tcell.ColorGray)

	// --- Background box ---
	for py := 0; py < totalH; py++ {
		for px := 0; px < fw; px++ {
			s.SetCell(fx+px, fy+py, bgStyle, ' ')
		}
	}

	// --- Current path (dim, line 0) ---
	cwdText := " " + fb.cwd
	if runewidth.StringWidth(cwdText) > fw-4 {
		// Truncate from left
		runes := []rune(cwdText)
		for runewidth.StringWidth(string(runes)) > fw-6 {
			runes = runes[1:]
		}
		cwdText = " …" + string(runes)
	}
	for i, ch := range cwdText {
		s.SetCell(fx+i, fy, pathStyle, ch)
	}

	// --- Input line (line 1) ---
	inputY := fy + 1
	prefix := " Path> "
	prefixRunes := []rune(prefix)

	// Draw prefix
	for i, ch := range prefixRunes {
		s.SetCell(fx+i, inputY, bgStyle, ch)
	}

	// Draw query text
	px := fx + len(prefixRunes)
	queryRunes := []rune(string(fb.query))
	for _, ch := range queryRunes {
		w := runewidth.RuneWidth(ch)
		if px+w < fx+fw-2 {
			s.SetCell(px, inputY, bgStyle, ch)
			px += w
		}
	}

	// Input cursor — blink block
	cursorStyle := bgStyle.Reverse(true)
	if fb.focusOnList {
		// No cursor when list is focused
	} else {
		cursorOffset := runewidth.StringWidth(string(queryRunes))
		cx := fx + len(prefixRunes) + cursorOffset
		if cx < fx+fw-1 {
			s.SetCell(cx, inputY, cursorStyle, ' ')
		}
	}

	// Clear rest of input line
	for i := px; i < fx+fw; i++ {
		s.SetCell(i, inputY, bgStyle, ' ')
	}

	// --- Separator (line 2) ---
	sepStyle := bgStyle.Foreground(tcell.ColorGray)
	for x := 0; x < fw; x++ {
		s.SetCell(fx+x, fy+2, sepStyle, '─')
	}

	// Focus indicator: label on separator
	focusLabel := " [List] "
	if !fb.focusOnList {
		focusLabel = " [Input] "
	}
	for i, ch := range focusLabel {
		s.SetCell(fx+fw-len([]rune(focusLabel))+i, fy+2, sepStyle, ch)
	}

	// --- File list (from line 3) ---
	startIdx := fb.scrollOffset
	endIdx := startIdx + visCount
	if endIdx > listLen {
		endIdx = listLen
	}

	for i := startIdx; i < endIdx; i++ {
		entry := fb.entries[i]
		displayY := fy + 3 + (i - startIdx)
		if displayY >= fy+totalH-1 {
			break
		}

		isSel := i == fb.selIdx && fb.focusOnList

		// Determine style
		var st tcell.Style
		switch {
		case isSel && entry.isDir:
			st = selDirStyle
		case isSel:
			st = selStyle
		case entry.upDir:
			st = dirStyle
		case entry.isDir:
			st = dirStyle
		default:
			st = bgStyle
		}

		// Fill entire row
		for x := 0; x < fw; x++ {
			s.SetCell(fx+x, displayY, st, ' ')
		}

		// Draw icon
		icon := "  "
		if entry.upDir {
			icon = "↑ " // parent directory
		} else if entry.isDir {
			icon = "> " // subdirectory
		}

		// Draw entry text
		text := icon + entry.name
		for j, ch := range text {
			s.SetCell(fx+j, displayY, st, ch)
		}
	}

	// --- Scroll indicator ---
	if listLen > visCount {
		scrollText := ""
		if fb.scrollOffset > 0 && fb.scrollOffset+visCount < listLen {
			scrollText = " ▲▼ "
		} else if fb.scrollOffset > 0 {
			scrollText = " ▲ "
		} else {
			scrollText = " ▼ "
		}
		for j, ch := range scrollText {
			s.SetCell(fx+fw-len([]rune(scrollText))+j, fy+totalH-1, bgStyle, ch)
		}
	}
}
