package ui

import (
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rwejlgaard/org/internal/model"
)

type viewMode int

const (
	modeList viewMode = iota
	modeAgenda
	modeEdit
	modeConfirmDelete
	modeCapture
	modeAddSubTask
	modeSetDeadline
)

type uiModel struct {
	orgFile      *model.OrgFile
	cursor       int
	mode         viewMode
	help         help.Model
	keys         keyMap
	width        int
	height       int
	statusMsg    string
	statusExpiry time.Time
	editingItem  *model.Item
	textarea     textarea.Model
	textinput    textinput.Model
	itemToDelete *model.Item
	reorderMode  bool
}

func initialModel(orgFile *model.OrgFile) uiModel {
	ta := textarea.New()
	ta.Placeholder = "Enter notes here (code blocks supported)..."
	ta.ShowLineNumbers = false

	ti := textinput.New()
	ti.Placeholder = "What needs doing?"
	ti.CharLimit = 200

	h := help.New()
	h.ShowAll = false

	return uiModel{
		orgFile:   orgFile,
		cursor:    0,
		mode:      modeList,
		help:      h,
		keys:      keys,
		textarea:  ta,
		textinput: ti,
	}
}

func (m uiModel) Init() tea.Cmd {
	return nil
}

func (m *uiModel) setStatus(msg string) {
	m.statusMsg = msg
	m.statusExpiry = time.Now().Add(3 * time.Second)
}

func (m uiModel) getVisibleItems() []*model.Item {
	if m.mode == modeAgenda {
		return m.getAgendaItems()
	}
	return m.orgFile.GetAllItems()
}

// RunUI starts the terminal UI
func RunUI(orgFile *model.OrgFile) error {
	p := tea.NewProgram(initialModel(orgFile), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
