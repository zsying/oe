package clipboard

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
)

// TestWriteOSC52 verifies the OSC 52 escape sequence is emitted correctly.
func TestWriteOSC52(t *testing.T) {
	var buf bytes.Buffer
	writeOSC52(&buf, "hello")

	got := buf.String()
	if !strings.HasPrefix(got, "\x1b]52;c;") {
		t.Fatalf("missing OSC 52 prefix, got %q", got)
	}
	if !strings.HasSuffix(got, "\x07") {
		t.Fatalf("missing BEL terminator, got %q", got)
	}
	// The middle must be the base64 of "hello".
	middle := got[len("\x1b]52;c;") : len(got)-1]
	if middle != base64.StdEncoding.EncodeToString([]byte("hello")) {
		t.Fatalf("unexpected payload %q", middle)
	}
}

// TestWriteRoutesToOSC52 verifies that, when an OSC 52 writer is set, Write
// emits the escape sequence (simulating the Linux clipboard fix) and still
// keeps the internal fallback.
func TestWriteRoutesToOSC52(t *testing.T) {
	var buf bytes.Buffer
	c := New()
	c.SetOSC52Writer(&buf)

	if err := c.Write("abc"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), base64.StdEncoding.EncodeToString([]byte("abc"))) {
		t.Fatalf("OSC 52 writer did not receive the payload, got %q", buf.String())
	}
	// Internal fallback must also hold the text.
	internal, _ := c.Read()
	if internal != "abc" {
		t.Fatalf("internal fallback should hold 'abc', got %q", internal)
	}
}
