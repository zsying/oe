package editor

// Mode represents the editor mode.
type Mode int

const (
	// ModeView is read-only navigation mode (default).
	ModeView Mode = iota
	// ModeEdit allows text modification.
	ModeEdit
)

func (m Mode) String() string {
	switch m {
	case ModeView:
		return "VIEW"
	case ModeEdit:
		return "EDIT"
	default:
		return "VIEW"
	}
}
