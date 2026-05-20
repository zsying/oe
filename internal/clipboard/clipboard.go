package clipboard

import (
	"github.com/atotto/clipboard"
)

// Clipboard abstracts system clipboard with internal fallback.
type Clipboard struct {
	internal string
	avail    bool
}

// New creates a new clipboard wrapper and detects system clipboard availability.
func New() Clipboard {
	c := Clipboard{}
	err := clipboard.WriteAll("")
	c.avail = err == nil
	return c
}

// Available returns whether system clipboard is accessible.
func (c *Clipboard) Available() bool { return c.avail }

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

// Write saves text to system clipboard and internal fallback.
func (c *Clipboard) Write(text string) error {
	c.internal = text
	if c.avail {
		return clipboard.WriteAll(text)
	}
	return nil
}
