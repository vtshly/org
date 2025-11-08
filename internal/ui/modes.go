package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rwejlgaard/org/internal/model"
	"github.com/rwejlgaard/org/internal/parser"
)

func (m uiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle special modes
	switch m.mode {
	case modeEdit:
		return m.updateEditMode(msg)
	case modeConfirmDelete:
		return m.updateConfirmDelete(msg)
	case modeCapture:
		return m.updateCapture(msg)
	case modeAddSubTask:
		return m.updateAddSubTask(msg)
	case modeSetDeadline:
		return m.updateSetDeadline(msg)
	case modeSetPriority:
		return m.updateSetPriority(msg)
	case modeSetEffort:
		return m.updateSetEffort(msg)
	case modeHelp:
		return m.updateHelp(msg)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		m.textarea.SetWidth(msg.Width - 4)
		m.textarea.SetHeight(msg.Height - 10)
		m.textinput.Width = 50 // Set a reasonable width for the text input
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.mode = modeHelp
			m.helpScroll = 0 // Reset scroll when entering help
			return m, nil

		case key.Matches(msg, m.keys.Up):
			if m.reorderMode {
				m.moveItemUp()
			} else {
				if m.cursor > 0 {
					m.cursor--
					// Update scroll to keep cursor visible
					availableHeight := m.height - 6 // Approximate
					if availableHeight < 5 {
						availableHeight = 5
					}
					m.updateScrollOffset(availableHeight)
				}
			}

		case key.Matches(msg, m.keys.Down):
			if m.reorderMode {
				m.moveItemDown()
			} else {
				items := m.getVisibleItems()
				if m.cursor < len(items)-1 {
					m.cursor++
					// Update scroll to keep cursor visible
					availableHeight := m.height - 6 // Approximate
					if availableHeight < 5 {
						availableHeight = 5
					}
					m.updateScrollOffset(availableHeight)
				}
			}

		case key.Matches(msg, m.keys.Left):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				m.cycleStateBackward(items[m.cursor])
				// Auto clock out when changing to DONE
				if items[m.cursor].State == model.StateDONE && items[m.cursor].IsClockedIn() {
					items[m.cursor].ClockOut()
				}
				m.setStatus("State changed")
			}

		case key.Matches(msg, m.keys.Right):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				items[m.cursor].CycleState()
				// Auto clock out when changing to DONE
				if items[m.cursor].State == model.StateDONE && items[m.cursor].IsClockedIn() {
					items[m.cursor].ClockOut()
				}
				m.setStatus("State changed")
			}

		case key.Matches(msg, m.keys.ShiftUp):
			m.moveItemUp()

		case key.Matches(msg, m.keys.ShiftDown):
			m.moveItemDown()

		case key.Matches(msg, m.keys.CycleState):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				items[m.cursor].CycleState()
				// Auto clock out when changing to DONE
				if items[m.cursor].State == model.StateDONE && items[m.cursor].IsClockedIn() {
					items[m.cursor].ClockOut()
				}
				m.setStatus("State changed")
			}

		case key.Matches(msg, m.keys.ToggleFold):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				items[m.cursor].ToggleFold()
				if items[m.cursor].Folded {
					m.setStatus("Folded")
				} else {
					m.setStatus("Unfolded")
				}
			}

		case key.Matches(msg, m.keys.EditNotes):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				m.editingItem = items[m.cursor]
				m.mode = modeEdit
				m.textarea.SetValue(strings.Join(m.editingItem.Notes, "\n"))
				m.textarea.Focus()
				return m, textarea.Blink
			}

		case key.Matches(msg, m.keys.Capture):
			m.mode = modeCapture
			m.textinput.SetValue("")
			m.textinput.Placeholder = "What needs doing?"
			m.textinput.Focus()
			return m, textinput.Blink

		case key.Matches(msg, m.keys.AddSubTask):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				m.editingItem = items[m.cursor]
				m.mode = modeAddSubTask
				m.textinput.SetValue("")
				m.textinput.Placeholder = "Sub-task title"
				m.textinput.Focus()
				return m, textinput.Blink
			}

		case key.Matches(msg, m.keys.Delete):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				m.itemToDelete = items[m.cursor]
				m.mode = modeConfirmDelete
			}

		case key.Matches(msg, m.keys.ToggleView):
			if m.mode == modeList {
				m.mode = modeAgenda
			} else {
				m.mode = modeList
			}
			m.cursor = 0

		case key.Matches(msg, m.keys.Save):
			if err := parser.Save(m.orgFile); err != nil {
				m.setStatus(fmt.Sprintf("Error saving: %v", err))
			} else {
				m.setStatus("Saved!")
			}

		case key.Matches(msg, m.keys.ToggleReorder):
			m.reorderMode = !m.reorderMode
			if m.reorderMode {
				m.setStatus("Reorder mode ON - Use ↑/↓ to move items, 'r' to exit")
			} else {
				m.setStatus("Reorder mode OFF")
			}

		case key.Matches(msg, m.keys.ClockIn):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				if items[m.cursor].ClockIn() {
					m.setStatus("Clocked in!")
				} else {
					m.setStatus("Already clocked in")
				}
			}

		case key.Matches(msg, m.keys.ClockOut):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				if items[m.cursor].ClockOut() {
					m.setStatus("Clocked out!")
				} else {
					m.setStatus("Not clocked in")
				}
			}

		case key.Matches(msg, m.keys.SetDeadline):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				m.editingItem = items[m.cursor]
				m.mode = modeSetDeadline
				m.textinput.SetValue("")
				m.textinput.Placeholder = "YYYY-MM-DD or +N (days from today)"
				m.textinput.Focus()
				return m, textinput.Blink
			}

		case key.Matches(msg, m.keys.SetPriority):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				m.editingItem = items[m.cursor]
				m.mode = modeSetPriority
				return m, nil
			}

		case key.Matches(msg, m.keys.SetEffort):
			items := m.getVisibleItems()
			if len(items) > 0 && m.cursor < len(items) {
				m.editingItem = items[m.cursor]
				m.mode = modeSetEffort
				m.textinput.SetValue("")
				m.textinput.Placeholder = "e.g., 8h, 2d, 1w"
				m.textinput.Focus()
				return m, textinput.Blink
			}
		}
	}

	return m, nil
}

func (m uiModel) updateEditMode(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(msg.Width - 4)
		m.textarea.SetHeight(msg.Height - 10)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			// Save notes and exit edit mode
			if m.editingItem != nil {
				noteText := m.textarea.Value()
				if noteText == "" {
					m.editingItem.Notes = []string{}
				} else {
					m.editingItem.Notes = strings.Split(noteText, "\n")
				}
			}
			m.mode = modeList
			m.textarea.Blur()
			m.setStatus("Notes saved")
			return m, nil
		}
	}

	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m uiModel) updateConfirmDelete(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			// Delete the item
			m.deleteItem(m.itemToDelete)
			m.mode = modeList
			m.itemToDelete = nil
			m.setStatus("Item deleted")
			// Adjust cursor if needed
			items := m.getVisibleItems()
			if m.cursor >= len(items) && len(items) > 0 {
				m.cursor = len(items) - 1
			}
		case "n", "N", "esc":
			m.mode = modeList
			m.itemToDelete = nil
			m.setStatus("Cancelled")
		}
	}
	return m, nil
}

func (m uiModel) updateCapture(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textinput.Width = 50

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			title := strings.TrimSpace(m.textinput.Value())
			if title != "" {
				// Create new TODO at top level
				newItem := &model.Item{
					Level:    1,
					State:    model.StateTODO,
					Title:    title,
					Notes:    []string{},
					Children: []*model.Item{},
				}
				// Insert at beginning
				m.orgFile.Items = append([]*model.Item{newItem}, m.orgFile.Items...)
				m.setStatus("TODO captured!")
			}
			m.mode = modeList
			m.textinput.Blur()
			m.cursor = 0
			return m, nil
		case tea.KeyEsc:
			m.mode = modeList
			m.textinput.Blur()
			m.setStatus("Cancelled")
			return m, nil
		}
	}

	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd
}

func (m uiModel) updateAddSubTask(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textinput.Width = 50

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			title := strings.TrimSpace(m.textinput.Value())
			if title != "" && m.editingItem != nil {
				// Create new sub-task
				newItem := &model.Item{
					Level:    m.editingItem.Level + 1,
					State:    model.StateTODO,
					Title:    title,
					Notes:    []string{},
					Children: []*model.Item{},
				}
				m.editingItem.Children = append(m.editingItem.Children, newItem)
				m.editingItem.Folded = false // Unfold to show new sub-task
				m.setStatus("Sub-task added!")
			}
			m.mode = modeList
			m.textinput.Blur()
			m.editingItem = nil
			return m, nil
		case tea.KeyEsc:
			m.mode = modeList
			m.textinput.Blur()
			m.editingItem = nil
			m.setStatus("Cancelled")
			return m, nil
		}
	}

	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd
}

func (m uiModel) updateSetDeadline(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textinput.Width = 50

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			input := strings.TrimSpace(m.textinput.Value())
			if m.editingItem != nil {
				if input == "" {
					// Empty input clears the deadline
					m.editingItem.Deadline = nil
					// Remove DEADLINE line from notes (only lines starting with DEADLINE:)
					var filteredNotes []string
					for _, note := range m.editingItem.Notes {
						trimmedNote := strings.TrimSpace(note)
						if !strings.HasPrefix(trimmedNote, "DEADLINE:") {
							filteredNotes = append(filteredNotes, note)
						}
					}
					m.editingItem.Notes = filteredNotes
					m.setStatus("Deadline cleared!")
				} else {
					deadline, err := parseDeadlineInput(input)
					if err != nil {
						m.setStatus(fmt.Sprintf("Invalid date: %v", err))
					} else {
						m.editingItem.Deadline = &deadline
						// Also update or add DEADLINE line in notes
						updatedNotes := false
						for i, note := range m.editingItem.Notes {
							trimmedNote := strings.TrimSpace(note)
							if strings.HasPrefix(trimmedNote, "DEADLINE:") {
								m.editingItem.Notes[i] = fmt.Sprintf("DEADLINE: <%s>", parser.FormatOrgDate(deadline))
								updatedNotes = true
								break
							}
						}
						// If DEADLINE wasn't in notes, it will be added by writeItem
						if !updatedNotes {
							// Remove old deadline lines just to be safe
							var filteredNotes []string
							for _, note := range m.editingItem.Notes {
								trimmedNote := strings.TrimSpace(note)
								if !strings.HasPrefix(trimmedNote, "DEADLINE:") {
									filteredNotes = append(filteredNotes, note)
								}
							}
							m.editingItem.Notes = filteredNotes
						}
						m.setStatus("Deadline set!")
					}
				}
			}
			m.mode = modeList
			m.textinput.Blur()
			m.editingItem = nil
			return m, nil
		case tea.KeyEsc:
			m.mode = modeList
			m.textinput.Blur()
			m.editingItem = nil
			m.setStatus("Cancelled")
			return m, nil
		}
	}

	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd
}

// parseDeadlineInput parses deadline input like "2024-01-15" or "+3" (3 days from now)
func parseDeadlineInput(input string) (time.Time, error) {
	// Check if it's a relative date (+N days)
	if strings.HasPrefix(input, "+") {
		daysStr := strings.TrimPrefix(input, "+")
		days := 0
		_, err := fmt.Sscanf(daysStr, "%d", &days)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid relative date format: %s", input)
		}
		return time.Now().AddDate(0, 0, days), nil
	}

	// Try parsing as absolute date
	formats := []string{
		"2006-01-02",
		"2006/01/02",
		"01/02/2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, input); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s (use YYYY-MM-DD or +N)", input)
}

func (m uiModel) updateSetPriority(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "A", "a":
			if m.editingItem != nil {
				m.editingItem.Priority = model.PriorityA
				m.setStatus("Priority set to A")
			}
			m.mode = modeList
			m.editingItem = nil
			return m, nil
		case "B", "b":
			if m.editingItem != nil {
				m.editingItem.Priority = model.PriorityB
				m.setStatus("Priority set to B")
			}
			m.mode = modeList
			m.editingItem = nil
			return m, nil
		case "C", "c":
			if m.editingItem != nil {
				m.editingItem.Priority = model.PriorityC
				m.setStatus("Priority set to C")
			}
			m.mode = modeList
			m.editingItem = nil
			return m, nil
		case " ", "enter":
			// Clear priority
			if m.editingItem != nil {
				m.editingItem.Priority = model.PriorityNone
				m.setStatus("Priority cleared")
			}
			m.mode = modeList
			m.editingItem = nil
			return m, nil
		case "esc":
			m.mode = modeList
			m.editingItem = nil
			m.setStatus("Cancelled")
			return m, nil
		}
	}
	return m, nil
}

func (m uiModel) updateSetEffort(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textinput.Width = 50

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			input := strings.TrimSpace(m.textinput.Value())
			if m.editingItem != nil {
				if input == "" {
					// Empty input clears the effort
					m.editingItem.Effort = ""
					m.setStatus("Effort cleared!")
				} else {
					// Set the effort value
					m.editingItem.Effort = input
					m.setStatus("Effort set!")
				}
			}
			m.mode = modeList
			m.textinput.Blur()
			m.editingItem = nil
			return m, nil
		case tea.KeyEsc:
			m.mode = modeList
			m.textinput.Blur()
			m.editingItem = nil
			m.setStatus("Cancelled")
			return m, nil
		}
	}

	m.textinput, cmd = m.textinput.Update(msg)
	return m, cmd
}

func (m *uiModel) cycleStateBackward(item *model.Item) {
	switch item.State {
	case model.StateNone:
		item.State = model.StateDONE
	case model.StateTODO:
		item.State = model.StateNone
	case model.StatePROG:
		item.State = model.StateTODO
	case model.StateBLOCK:
		item.State = model.StatePROG
	case model.StateDONE:
		item.State = model.StateBLOCK
	}
}

func (m *uiModel) deleteItem(item *model.Item) {
	var removeFromList func([]*model.Item, *model.Item) []*model.Item
	removeFromList = func(items []*model.Item, target *model.Item) []*model.Item {
		result := []*model.Item{}
		for _, it := range items {
			if it == target {
				continue
			}
			it.Children = removeFromList(it.Children, target)
			result = append(result, it)
		}
		return result
	}
	m.orgFile.Items = removeFromList(m.orgFile.Items, item)
}

func (m *uiModel) moveItemUp() {
	items := m.getVisibleItems()
	if len(items) == 0 || m.cursor == 0 {
		return
	}

	currentItem := items[m.cursor]

	// Find the previous sibling (not child, not parent's sibling)
	prevSibling := m.findPreviousSibling(currentItem)
	if prevSibling == nil {
		m.setStatus("Cannot move - already at top of list")
		return
	}

	m.swapItems(currentItem, prevSibling)
	m.setStatus("Item moved up")

	// Update cursor to follow the item
	items = m.getVisibleItems()
	for i, item := range items {
		if item == currentItem {
			m.cursor = i
			break
		}
	}
}

func (m *uiModel) moveItemDown() {
	items := m.getVisibleItems()
	if len(items) == 0 || m.cursor >= len(items)-1 {
		return
	}

	currentItem := items[m.cursor]

	// Find the next sibling (not child, not parent's sibling)
	nextSibling := m.findNextSibling(currentItem)
	if nextSibling == nil {
		m.setStatus("Cannot move - already at bottom of list")
		return
	}

	m.swapItems(currentItem, nextSibling)
	m.setStatus("Item moved down")

	// Update cursor to follow the item
	items = m.getVisibleItems()
	for i, item := range items {
		if item == currentItem {
			m.cursor = i
			break
		}
	}
}

func (m *uiModel) findPreviousSibling(item *model.Item) *model.Item {
	// Find the parent list containing this item and return the previous sibling
	var findInList func([]*model.Item) *model.Item
	findInList = func(items []*model.Item) *model.Item {
		for i, it := range items {
			if it == item && i > 0 {
				// Found the item and there's a previous sibling
				return items[i-1]
			}
			// Recursively check children
			if result := findInList(it.Children); result != nil {
				return result
			}
		}
		return nil
	}
	return findInList(m.orgFile.Items)
}

func (m *uiModel) findNextSibling(item *model.Item) *model.Item {
	// Find the parent list containing this item and return the next sibling
	var findInList func([]*model.Item) *model.Item
	findInList = func(items []*model.Item) *model.Item {
		for i, it := range items {
			if it == item && i < len(items)-1 {
				// Found the item and there's a next sibling
				return items[i+1]
			}
			// Recursively check children
			if result := findInList(it.Children); result != nil {
				return result
			}
		}
		return nil
	}
	return findInList(m.orgFile.Items)
}

func (m *uiModel) swapItems(item1, item2 *model.Item) {
	// Find parent list containing both items
	var swapInList func([]*model.Item) bool
	swapInList = func(items []*model.Item) bool {
		for i := 0; i < len(items)-1; i++ {
			if items[i] == item1 && items[i+1] == item2 {
				items[i], items[i+1] = items[i+1], items[i]
				return true
			}
			if items[i] == item2 && items[i+1] == item1 {
				items[i], items[i+1] = items[i+1], items[i]
				return true
			}
			if swapInList(items[i].Children) {
				return true
			}
		}
		if len(items) > 0 && swapInList(items[len(items)-1].Children) {
			return true
		}
		return false
	}
	swapInList(m.orgFile.Items)
}

func (m uiModel) updateHelp(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "?", "esc", "q":
			m.mode = modeList
			m.helpScroll = 0 // Reset scroll when exiting
			return m, nil
		case "up", "k":
			if m.helpScroll > 0 {
				m.helpScroll--
			}
			return m, nil
		case "down", "j":
			m.helpScroll++
			// The view will handle clamping to max scroll
			return m, nil
		case "pageup":
			m.helpScroll -= 10
			if m.helpScroll < 0 {
				m.helpScroll = 0
			}
			return m, nil
		case "pagedown":
			m.helpScroll += 10
			return m, nil
		case "home", "g":
			m.helpScroll = 0
			return m, nil
		}
	}
	return m, nil
}
