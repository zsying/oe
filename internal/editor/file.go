package editor

import (
	"os"
	"path/filepath"
	"strings"
)

// OpenFile loads a file into the editor.
func (e *Editor) OpenFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet — will be created on save
			e.Buffer.SetFilename(path)
			return nil
		}
		return err
	}
	if info.IsDir() {
		return os.ErrInvalid
	}
	if err := e.Buffer.Load(path); err != nil {
		return err
	}
	e.Buffer.SetFileType(strings.TrimPrefix(filepath.Ext(path), "."))
	return nil
}

// SaveFile writes the buffer to its current file path.
func (e *Editor) SaveFile() error {
	if e.Buffer.Filename() == "" {
		// No filename set — caller should prompt for Save As
		return nil
	}
	return e.Buffer.Save()
}

// SaveFileAs writes the buffer to a new path and updates the filename.
func (e *Editor) SaveFileAs(path string) error {
	e.Buffer.SetFilename(path)
	return e.Buffer.Save()
}
