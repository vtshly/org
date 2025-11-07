package ui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up            key.Binding
	Down          key.Binding
	Left          key.Binding
	Right         key.Binding
	ShiftUp       key.Binding
	ShiftDown     key.Binding
	CycleState    key.Binding
	ToggleView    key.Binding
	Quit          key.Binding
	Help          key.Binding
	Capture       key.Binding
	AddSubTask    key.Binding
	Delete        key.Binding
	Save          key.Binding
	ToggleFold    key.Binding
	EditNotes     key.Binding
	ToggleReorder key.Binding
	ClockIn       key.Binding
	ClockOut      key.Binding
	SetDeadline   key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "cycle state backward"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "cycle state forward"),
	),
	ShiftUp: key.NewBinding(
		key.WithKeys("shift+up"),
		key.WithHelp("shift+↑", "move item up"),
	),
	ShiftDown: key.NewBinding(
		key.WithKeys("shift+down"),
		key.WithHelp("shift+↓", "move item down"),
	),
	CycleState: key.NewBinding(
		key.WithKeys("t", " "),
		key.WithHelp("t/space", "cycle todo state"),
	),
	ToggleFold: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "fold/unfold"),
	),
	EditNotes: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "edit notes"),
	),
	ToggleView: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "toggle agenda view"),
	),
	Capture: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "capture TODO"),
	),
	AddSubTask: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "add sub-task"),
	),
	Delete: key.NewBinding(
		key.WithKeys("shift+d"),
		key.WithHelp("shift+d", "delete item"),
	),
	Save: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save"),
	),
	ToggleReorder: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "reorder mode"),
	),
	ClockIn: key.NewBinding(
		key.WithKeys("i"),
		key.WithHelp("i", "clock in"),
	),
	ClockOut: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "clock out"),
	),
	SetDeadline: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "set deadline"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help}
}

func (k keyMap) FullHelp() [][]key.Binding {
	// This will be overridden by custom rendering in viewFullHelp
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.ToggleFold, k.EditNotes, k.ToggleReorder},
		{k.Capture, k.AddSubTask, k.Delete, k.Save},
		{k.ToggleView, k.Help, k.Quit},
	}
}

// getAllBindings returns all keybindings as a flat list
func (k keyMap) getAllBindings() []key.Binding {
	return []key.Binding{
		k.Up, k.Down, k.Left, k.Right,
		k.ToggleFold, k.EditNotes, k.ToggleReorder,
		k.Capture, k.AddSubTask, k.Delete, k.Save,
		k.ClockIn, k.ClockOut, k.SetDeadline,
		k.ToggleView, k.Help, k.Quit,
	}
}
