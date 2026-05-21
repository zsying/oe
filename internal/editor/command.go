package editor

// Command represents an executable editor command.
type Command struct {
	ID       string
	Title    string
	Shortcut string
	Handler  func() error
}

// CommandRegistry holds all available commands.
type CommandRegistry struct {
	commands []Command
}

// NewCommandRegistry creates the default command set.
func NewCommandRegistry(ed *Editor) *CommandRegistry {
	r := &CommandRegistry{}
	r.commands = []Command{
		{ID: "file.open", Title: "Open File…", Shortcut: "Ctrl+O", Handler: ed.cmdOpen},
		{ID: "file.save", Title: "Save", Shortcut: "Ctrl+S", Handler: ed.cmdSave},
		{ID: "file.saveAs", Title: "Save As…", Shortcut: "", Handler: ed.cmdSaveAs},
		{ID: "file.new", Title: "New File", Shortcut: "", Handler: ed.cmdNew},
		{ID: "app.quit", Title: "Quit", Shortcut: "Ctrl+Q", Handler: ed.cmdQuit},
		{ID: "edit.undo", Title: "Undo", Shortcut: "Ctrl+Z", Handler: ed.cmdUndo},
		{ID: "edit.cut", Title: "Cut", Shortcut: "Ctrl+X", Handler: ed.cmdCut},
		{ID: "edit.copy", Title: "Copy", Shortcut: "Ctrl+C", Handler: ed.cmdCopy},
		{ID: "edit.paste", Title: "Paste", Shortcut: "Ctrl+V", Handler: ed.cmdPaste},
		{ID: "edit.delete", Title: "Delete", Shortcut: "Del", Handler: ed.cmdDelete},
		{ID: "edit.selectAll", Title: "Select All", Shortcut: "Ctrl+A", Handler: ed.cmdSelectAll},
		{ID: "view.toggleLineNumber", Title: "Toggle Line Numbers", Shortcut: "", Handler: ed.cmdToggleLineNum},
		{ID: "mode.toggle", Title: "Toggle View/Edit Mode", Shortcut: "Ctrl+E", Handler: ed.cmdToggleMode},
		{ID: "find.find", Title: "Find…", Shortcut: "Ctrl+F", Handler: ed.cmdFind},
		{ID: "find.next", Title: "Find Next", Shortcut: "F3", Handler: ed.cmdFindNext},
		{ID: "find.replace", Title: "Replace…", Shortcut: "Ctrl+H", Handler: ed.cmdReplace},
	}
	return r
}

// All returns all registered commands.
func (r *CommandRegistry) All() []Command { return r.commands }

// Find looks up a command by ID and returns a pointer to it.
func (r *CommandRegistry) Find(id string) *Command {
	for i := range r.commands {
		if r.commands[i].ID == id {
			return &r.commands[i]
		}
	}
	return nil
}
