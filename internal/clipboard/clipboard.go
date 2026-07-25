package clipboard

import (
	"encoding/base64"
	"io"

	"github.com/atotto/clipboard"
)

// Clipboard abstracts system clipboard with internal fallback.
type Clipboard struct {
	internal string
	avail    bool
	// osc52, when set, enables clipboard writes via the OSC 52 terminal
	// escape sequence. This is the most reliable way for a terminal editor to
	// set the system clipboard on Linux (and macOS), including over SSH and
	// inside tmux, without depending on xclip/xsel/wl-clipboard being present
	// or a clipboard manager running.
	osc52 io.Writer
}

// New creates a new clipboard wrapper and detects system clipboard availability.
func New() Clipboard {
	c := Clipboard{}
	err := clipboard.WriteAll("")
	c.avail = err == nil
	return c
}

// SetOSC52Writer enables clipboard writes through the OSC 52 escape sequence.
// Pass the terminal writer (typically os.Stdout). Passing nil disables it.
func (c *Clipboard) SetOSC52Writer(w io.Writer) {
	c.osc52 = w
}

// Available returns whether any clipboard backend is usable.
func (c *Clipboard) Available() bool { return c.avail || c.osc52 != nil }

// Read returns clipboard text from system clipboard or internal fallback.
func (c *Clipboard) Read() (string, error) {
	if c.avail {
		text, err := clipboard.ReadAll()
		if err == nil && text != "" {
			return text, nil
		}
	}
	return c.internal, nil
}

// Write saves text to the OSC 52 terminal clipboard (when enabled), the system
// clipboard (when available), and the internal fallback.
func (c *Clipboard) Write(text string) error {
	c.internal = text
	if c.osc52 != nil {
		writeOSC52(c.osc52, text)
	}
	if c.avail {
		return clipboard.WriteAll(text)
	}
	return nil
}

// writeOSC52 emits an OSC 52 sequence that sets the terminal's clipboard to the
// base64-encoded text: ESC ] 52 ; c ; <base64> BEL
func writeOSC52(w io.Writer, text string) {
	b64 := base64.StdEncoding.EncodeToString([]byte(text))
	seq := "\x1b]52;c;" + b64 + "\x07"
	_, _ = io.WriteString(w, seq)
}
