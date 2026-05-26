package clipboard

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/atotto/clipboard"
)

// Clipboard abstracts system clipboard with internal fallback.
type Clipboard struct {
	internal string
}

// New creates a new clipboard wrapper.
func New() Clipboard {
	return Clipboard{}
}

// Read returns clipboard text — prefers internal buffer first (content copied inside OE),
// then falls back to system clipboard for content copied from outside.
func (c *Clipboard) Read() (string, error) {
	if c.internal != "" {
		return c.internal, nil
	}
	// Try system clipboard library (xsel/xclip on Linux, pbcopy on macOS, win32 on Windows)
	text, err := clipboard.ReadAll()
	if err == nil && text != "" {
		return text, nil
	}
	return "", nil
}

// Write saves text to system clipboard via three backends in order:
//  1. OSC 52 terminal escape — works over SSH / WSL / containers in modern terminals
//  2. atotto clipboard library — xsel/xclip (Linux), pbcopy (macOS), win32 (Windows)
//  3. Internal memory — always works within the editor
func (c *Clipboard) Write(text string) error {
	c.internal = text

	// 1. OSC 52 — terminal emulator writes to system clipboard directly
	if err := writeOSC52(text); err == nil {
		return nil
	}

	// 2. System clipboard library
	if err := clipboard.WriteAll(text); err == nil {
		return nil
	}

	// 3. Internal fallback — always succeeds
	return nil
}

// --- OSC 52 terminal clipboard protocol ---
//
// Many modern terminals (kitty, iTerm2, tmux, Windows Terminal, GNOME Terminal,
// WezTerm, foot) support the Operating System Command 52 escape sequence for
// clipboard access. This works even over SSH or in containers because the
// terminal emulator (which has access to the real system clipboard) handles the
// sequence on the client side.
//
// Write format: ESC ] 52 ; c ; BASE64_DATA BEL
// Read format:  ESC ] 52 ; c ; ? BEL — terminal responds with ESC ] 52 ; c ; BASE64_DATA ST
// Note: Reading via OSC 52 is not implemented here because tcell's event loop
// consumes stdin, making the response impossible to capture synchronously.

const (
	osc52Prefix = "\033]52;c;"
	osc52Suffix = "\007" // BEL — widely supported terminator
)

// writeOSC52 sends a clipboard write request via OSC 52 escape sequence.
// The terminal emulator decodes the base64 payload and places it on the system clipboard.
func writeOSC52(text string) error {
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	_, err := os.Stdout.WriteString(fmt.Sprintf("%s%s%s", osc52Prefix, encoded, osc52Suffix))
	if err != nil {
		return err
	}
	// Flush to ensure the sequence reaches the terminal before tcell redraws
	return os.Stdout.Sync()
}
