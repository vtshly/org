package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/alecthomas/chroma/v2/quick"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/rwejlgaard/org/internal/model"
	"github.com/rwejlgaard/org/internal/parser"
)

// dynamicKeyMap is a helper type for rendering keybindings with dynamic layout
type dynamicKeyMap struct {
	rows [][]key.Binding
}

// ShortHelp for dynamicKeyMap
func (d dynamicKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{}
}

// FullHelp for dynamicKeyMap
func (d dynamicKeyMap) FullHelp() [][]key.Binding {
	return d.rows
}

// renderFullHelp renders the help with width-aware layout
func (m uiModel) renderFullHelp() string {
	bindings := m.keys.getAllBindings()

	// Estimate the width needed for each keybinding (key + desc + padding)
	// Average is roughly 20-25 chars per binding
	const estimatedBindingWidth = 22
	const minWidth = 40 // Minimum width before stacking

	var columnsPerRow int
	if m.width < minWidth {
		columnsPerRow = 1 // Stack vertically on very narrow terminals
	} else if m.width < 80 {
		columnsPerRow = 2 // Two columns on narrow terminals
	} else if m.width < 120 {
		columnsPerRow = 3 // Three columns on medium terminals
	} else {
		columnsPerRow = 4 // Four columns on wide terminals
	}

	// Build rows based on columns per row
	var rows [][]key.Binding
	for i := 0; i < len(bindings); i += columnsPerRow {
		end := i + columnsPerRow
		if end > len(bindings) {
			end = len(bindings)
		}
		rows = append(rows, bindings[i:end])
	}

	// Use the help model to render with our dynamic layout
	h := help.New()
	h.Width = m.width
	h.ShowAll = true

	// Create a temporary keyMap for rendering
	dkm := dynamicKeyMap{rows: rows}

	return h.View(dkm)
}

func (m uiModel) View() string {
	switch m.mode {
	case modeEdit:
		return m.viewEditMode()
	case modeConfirmDelete:
		return m.viewConfirmDelete()
	case modeCapture:
		return m.viewCapture()
	case modeAddSubTask:
		return m.viewAddSubTask()
	case modeSetDeadline:
		return m.viewSetDeadline()
	case modeSetPriority:
		return m.viewSetPriority()
	case modeSetEffort:
		return m.viewSetEffort()
	}

	// Build footer (status + help)
	var footer strings.Builder

	// Status message
	if time.Now().Before(m.statusExpiry) {
		footer.WriteString(statusStyle.Render(m.statusMsg))
		footer.WriteString("\n")
	}

	// Help
	if m.help.ShowAll {
		footer.WriteString(m.renderFullHelp())
	} else {
		footer.WriteString(m.help.View(m.keys))
	}

	footerHeight := lipgloss.Height(footer.String())

	// Build main content
	var content strings.Builder

	// Title
	title := "Org Mode - List View"
	if m.mode == modeAgenda {
		title = "Org Mode - Agenda View (Next 7 Days)"
	}
	if m.reorderMode {
		reorderIndicator := lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(" [REORDER MODE]")
		content.WriteString(titleStyle.Render(title))
		content.WriteString(reorderIndicator)
	} else {
		content.WriteString(titleStyle.Render(title))
	}
	content.WriteString("\n\n")

	// Calculate available height for items (total - title - footer)
	availableHeight := m.height - 3 - footerHeight // 3 for title + spacing
	if availableHeight < 5 {
		availableHeight = 5 // Minimum height
	}

	// Items
	items := m.getVisibleItems()
	if len(items) == 0 {
		content.WriteString("No items. Press 'c' to capture a new TODO.\n")
	}

	// Build a map of item index to line count (for scrolling)
	itemLineCount := make([]int, len(items))
	for i, item := range items {
		lineCount := 1 // The item itself
		if !item.Folded && len(item.Notes) > 0 && m.mode == modeList {
			filteredNotes := filterLogbookDrawer(item.Notes)
			highlightedNotes := renderNotesWithHighlighting(filteredNotes)
			lineCount += len(highlightedNotes)
		}
		itemLineCount[i] = lineCount
	}

	// Calculate total lines up to cursor
	totalLinesBeforeCursor := 0
	for i := 0; i < m.cursor && i < len(itemLineCount); i++ {
		totalLinesBeforeCursor += itemLineCount[i]
	}

	// Determine the scroll offset (without modifying m - View should be pure)
	scrollOffset := m.scrollOffset
	if totalLinesBeforeCursor < scrollOffset {
		// Cursor is above visible area, scroll up
		scrollOffset = totalLinesBeforeCursor
	} else if totalLinesBeforeCursor >= scrollOffset+availableHeight {
		// Cursor is below visible area, scroll down
		scrollOffset = totalLinesBeforeCursor - availableHeight + 1
	}

	// Render items starting from scroll offset
	itemLines := 0
	for i, item := range items {
		// Calculate which line this item starts at
		itemStartLine := 0
		for j := 0; j < i; j++ {
			itemStartLine += itemLineCount[j]
		}

		// Skip items before scroll offset
		if itemStartLine+itemLineCount[i] <= scrollOffset {
			continue
		}

		// Stop if we've filled the screen
		if itemLines >= availableHeight {
			break
		}

		// Skip partial items at the top if needed
		if itemStartLine < scrollOffset {
			// This item is partially scrolled off the top
			linesToSkip := scrollOffset - itemStartLine
			if linesToSkip < itemLineCount[i] {
				// Render the visible parts
				if linesToSkip == 0 {
					line := m.renderItem(item, i == m.cursor)
					content.WriteString(line)
					content.WriteString("\n")
					itemLines++
					linesToSkip++
				}

				// Render remaining notes
				if !item.Folded && len(item.Notes) > 0 && m.mode == modeList {
					indent := strings.Repeat("  ", item.Level)
					filteredNotes := filterLogbookDrawer(item.Notes)
					highlightedNotes := renderNotesWithHighlighting(filteredNotes)
					for noteIdx := linesToSkip - 1; noteIdx < len(highlightedNotes) && itemLines < availableHeight; noteIdx++ {
						content.WriteString(indent)
						content.WriteString("  " + highlightedNotes[noteIdx])
						content.WriteString("\n")
						itemLines++
					}
				}
			}
			continue
		}

		// Render the full item
		line := m.renderItem(item, i == m.cursor)
		content.WriteString(line)
		content.WriteString("\n")
		itemLines++

		// Show notes if not folded
		if !item.Folded && len(item.Notes) > 0 && m.mode == modeList {
			indent := strings.Repeat("  ", item.Level)
			filteredNotes := filterLogbookDrawer(item.Notes)
			highlightedNotes := renderNotesWithHighlighting(filteredNotes)
			for _, note := range highlightedNotes {
				if itemLines >= availableHeight {
					break
				}
				content.WriteString(indent)
				content.WriteString("  " + note)
				content.WriteString("\n")
				itemLines++
			}
		}
	}

	// Combine content and footer with padding
	contentHeight := lipgloss.Height(content.String())
	paddingNeeded := m.height - contentHeight - footerHeight
	if paddingNeeded < 0 {
		paddingNeeded = 0
	}

	var result strings.Builder
	result.WriteString(content.String())
	if paddingNeeded > 0 {
		result.WriteString(strings.Repeat("\n", paddingNeeded))
	}
	result.WriteString(footer.String())

	return result.String()
}

func (m uiModel) viewConfirmDelete() string {
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(1, 2).
		Width(60)

	var content strings.Builder
	content.WriteString(titleStyle.Render("⚠ Delete Item"))
	content.WriteString("\n\n")

	if m.itemToDelete != nil {
		itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("202")).Bold(true)
		content.WriteString(itemStyle.Render(m.itemToDelete.Title))
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString(statusStyle.Render("This will delete the item and all sub-tasks."))
	content.WriteString("\n\n")
	content.WriteString("Press Y to confirm • N or ESC to cancel")

	dialog := dialogStyle.Render(content.String())

	// Center the dialog horizontally and vertically
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)
}

func (m uiModel) viewCapture() string {
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(1, 2).
		Width(60)

	var content strings.Builder
	content.WriteString(titleStyle.Render("Capture TODO"))
	content.WriteString("\n\n")
	content.WriteString(m.textinput.View())
	content.WriteString("\n\n")
	content.WriteString(statusStyle.Render("Press Enter to save • ESC to cancel"))

	dialog := dialogStyle.Render(content.String())

	// Center the dialog horizontally and vertically
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)
}

func (m uiModel) viewAddSubTask() string {
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(1, 2).
		Width(60)

	var content strings.Builder
	content.WriteString(titleStyle.Render("Add Sub-Task"))
	content.WriteString("\n")
	if m.editingItem != nil {
		content.WriteString(statusStyle.Render(fmt.Sprintf("Under: %s", m.editingItem.Title)))
	}
	content.WriteString("\n\n")
	content.WriteString(m.textinput.View())
	content.WriteString("\n\n")
	content.WriteString(statusStyle.Render("Press Enter to save • ESC to cancel"))

	dialog := dialogStyle.Render(content.String())

	// Center the dialog horizontally and vertically
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)
}

func (m uiModel) viewSetDeadline() string {
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("141")).
		Padding(1, 2).
		Width(60)

	var content strings.Builder
	content.WriteString(titleStyle.Render("Set Deadline"))
	content.WriteString("\n")
	if m.editingItem != nil {
		content.WriteString(statusStyle.Render(fmt.Sprintf("For: %s", m.editingItem.Title)))
	}
	content.WriteString("\n\n")
	content.WriteString(m.textinput.View())
	content.WriteString("\n\n")
	content.WriteString(statusStyle.Render("Examples: 2025-12-31, +7 (7 days from now)"))
	content.WriteString("\n")
	content.WriteString(statusStyle.Render("Leave empty to clear deadline"))
	content.WriteString("\n")
	content.WriteString(statusStyle.Render("Press Enter to save • ESC to cancel"))

	dialog := dialogStyle.Render(content.String())

	// Center the dialog horizontally and vertically
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)
}

func (m uiModel) viewSetPriority() string {
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("214")).
		Padding(1, 2).
		Width(60)

	var content strings.Builder
	content.WriteString(titleStyle.Render("Set Priority"))
	content.WriteString("\n")
	if m.editingItem != nil {
		content.WriteString(statusStyle.Render(fmt.Sprintf("For: %s", m.editingItem.Title)))
		content.WriteString("\n")
		if m.editingItem.Priority != model.PriorityNone {
			content.WriteString(statusStyle.Render(fmt.Sprintf("Current: [#%s]", m.editingItem.Priority)))
		}
	}
	content.WriteString("\n\n")

	// Show priority options with styling
	priorityAStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	priorityBStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
	priorityCStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)

	content.WriteString(priorityAStyle.Render("[A] High Priority") + "\n")
	content.WriteString(priorityBStyle.Render("[B] Medium Priority") + "\n")
	content.WriteString(priorityCStyle.Render("[C] Low Priority") + "\n")
	content.WriteString("\n")
	content.WriteString(statusStyle.Render("Press Space/Enter to clear priority"))
	content.WriteString("\n")
	content.WriteString(statusStyle.Render("Press ESC to cancel"))

	dialog := dialogStyle.Render(content.String())

	// Center the dialog horizontally and vertically
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)
}

func (m uiModel) viewSetEffort() string {
	dialogStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("141")).
		Padding(1, 2).
		Width(60)

	var content strings.Builder
	content.WriteString(titleStyle.Render("Set Effort"))
	content.WriteString("\n")
	if m.editingItem != nil {
		content.WriteString(statusStyle.Render(fmt.Sprintf("For: %s", m.editingItem.Title)))
		content.WriteString("\n")
		if m.editingItem.Effort != "" {
			content.WriteString(statusStyle.Render(fmt.Sprintf("Current: %s", m.editingItem.Effort)))
		}
	}
	content.WriteString("\n\n")
	content.WriteString(m.textinput.View())
	content.WriteString("\n\n")
	content.WriteString(statusStyle.Render("Examples: 8h, 2d, 1w, 4h30m"))
	content.WriteString("\n")
	content.WriteString(statusStyle.Render("Leave empty to clear effort"))
	content.WriteString("\n")
	content.WriteString(statusStyle.Render("Press Enter to save • ESC to cancel"))

	dialog := dialogStyle.Render(content.String())

	// Center the dialog horizontally and vertically
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, dialog)
}

func (m uiModel) viewEditMode() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Editing Notes"))
	b.WriteString("\n")
	if m.editingItem != nil {
		b.WriteString(fmt.Sprintf("Item: %s\n", m.editingItem.Title))
	}
	b.WriteString(statusStyle.Render("Press ESC to save and exit"))
	b.WriteString("\n\n")

	b.WriteString(m.textarea.View())

	return b.String()
}

// filterLogbookDrawer removes LOGBOOK and PROPERTIES drawer content and scheduling metadata from notes
func filterLogbookDrawer(notes []string) []string {
	var filtered []string
	inLogbook := false
	inProperties := false

	for _, note := range notes {
		trimmed := strings.TrimSpace(note)

		// Check for start of LOGBOOK drawer
		if trimmed == ":LOGBOOK:" {
			inLogbook = true
			continue
		}

		// Check for start of PROPERTIES drawer
		if trimmed == ":PROPERTIES:" {
			inProperties = true
			continue
		}

		// Check for end of drawer
		if trimmed == ":END:" {
			if inLogbook {
				inLogbook = false
				continue
			}
			if inProperties {
				inProperties = false
				continue
			}
		}

		// Skip lines inside LOGBOOK or PROPERTIES drawer
		if inLogbook || inProperties {
			continue
		}

		// Skip SCHEDULED and DEADLINE lines
		if strings.HasPrefix(trimmed, "SCHEDULED:") || strings.HasPrefix(trimmed, "DEADLINE:") {
			continue
		}

		filtered = append(filtered, note)
	}

	return filtered
}

// renderNotesWithHighlighting renders notes with syntax highlighting for code blocks
func renderNotesWithHighlighting(notes []string) []string {
	if len(notes) == 0 {
		return notes
	}

	var result []string
	var inCodeBlock bool
	var codeLanguage string
	var codeLines []string
	var codeBlockDelimiter string // Track whether we're in #+BEGIN_SRC or ``` block

	for _, note := range notes {
		trimmed := strings.TrimSpace(note)

		// Check for org-mode style code block start
		if strings.HasPrefix(trimmed, "#+BEGIN_SRC") {
			inCodeBlock = true
			codeBlockDelimiter = "org"
			// Extract language
			parts := strings.Fields(trimmed)
			if len(parts) > 1 {
				codeLanguage = strings.ToLower(parts[1])
			} else {
				codeLanguage = "text"
			}
			result = append(result, note) // Keep the delimiter visible
			codeLines = []string{}
			continue
		}

		// Check for markdown style code block start
		if strings.HasPrefix(trimmed, "```") {
			if !inCodeBlock {
				// Starting a code block
				inCodeBlock = true
				codeBlockDelimiter = "markdown"
				// Extract language
				lang := strings.TrimPrefix(trimmed, "```")
				if lang != "" {
					codeLanguage = strings.ToLower(lang)
				} else {
					codeLanguage = "text"
				}
				result = append(result, note) // Keep the delimiter visible
				codeLines = []string{}
				continue
			} else if codeBlockDelimiter == "markdown" {
				// Ending a markdown code block
				inCodeBlock = false
				// Highlight and add the code
				if len(codeLines) > 0 {
					highlighted := highlightCode(strings.Join(codeLines, "\n"), codeLanguage)
					highlightedLines := strings.Split(highlighted, "\n")
					result = append(result, highlightedLines...)
				}
				result = append(result, note) // Keep the delimiter visible
				codeLines = []string{}
				codeLanguage = ""
				codeBlockDelimiter = ""
				continue
			}
		}

		// Check for org-mode style code block end
		if strings.HasPrefix(trimmed, "#+END_SRC") {
			inCodeBlock = false
			// Highlight and add the code
			if len(codeLines) > 0 {
				highlighted := highlightCode(strings.Join(codeLines, "\n"), codeLanguage)
				highlightedLines := strings.Split(highlighted, "\n")
				result = append(result, highlightedLines...)
			}
			result = append(result, note) // Keep the delimiter visible
			codeLines = []string{}
			codeLanguage = ""
			codeBlockDelimiter = ""
			continue
		}

		// If in code block, accumulate lines
		if inCodeBlock {
			codeLines = append(codeLines, note)
		} else {
			result = append(result, note)
		}
	}

	// Handle case where code block wasn't closed
	if inCodeBlock && len(codeLines) > 0 {
		highlighted := highlightCode(strings.Join(codeLines, "\n"), codeLanguage)
		highlightedLines := strings.Split(highlighted, "\n")
		result = append(result, highlightedLines...)
	}

	return result
}

// highlightCode applies syntax highlighting to code
func highlightCode(code, language string) string {
	if code == "" {
		return code
	}

	var buf strings.Builder
	err := quick.Highlight(&buf, code, language, "terminal256", "monokai")
	if err != nil {
		// If highlighting fails, return the original code
		return code
	}

	return strings.TrimRight(buf.String(), "\n")
}

func (m uiModel) renderItem(item *model.Item, isCursor bool) string {
	var b strings.Builder

	// Indentation for level
	indent := strings.Repeat("  ", item.Level-1)
	b.WriteString(indent)

	// Fold indicator
	if len(item.Children) > 0 || len(item.Notes) > 0 {
		if item.Folded {
			b.WriteString(foldedStyle.Render("▶ "))
		} else {
			b.WriteString(foldedStyle.Render("▼ "))
		}
	} else {
		b.WriteString("  ")
	}

	// State
	stateStr := ""
	switch item.State {
	case model.StateTODO:
		stateStr = todoStyle.Render("[TODO] ")
	case model.StatePROG:
		stateStr = progStyle.Render("[PROG] ")
	case model.StateBLOCK:
		stateStr = blockStyle.Render("[BLOCK]")
	case model.StateDONE:
		stateStr = doneStyle.Render("[DONE] ")
	default:
		stateStr = "       " // Empty space for alignment
	}
	b.WriteString(stateStr)
	b.WriteString(" ")

	// Priority
	if item.Priority != model.PriorityNone {
		var priorityStyle lipgloss.Style
		switch item.Priority {
		case model.PriorityA:
			priorityStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
		case model.PriorityB:
			priorityStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
		case model.PriorityC:
			priorityStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Bold(true)
		}
		b.WriteString(priorityStyle.Render(fmt.Sprintf("[#%s] ", item.Priority)))
	}

	// Title
	b.WriteString(item.Title)

	// Effort
	if item.Effort != "" {
		effortStr := fmt.Sprintf(" (Effort: %s)", item.Effort)
		effortStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("141")) // Purple
		b.WriteString(effortStyle.Render(effortStr))
	}

	// Clock status
	if item.IsClockedIn() {
		duration := item.GetCurrentClockDuration()
		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60
		clockStr := fmt.Sprintf(" [CLOCKED IN: %dh %dm]", hours, minutes)
		clockStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true) // Bright green
		b.WriteString(clockStyle.Render(clockStr))
	}

	// Total clocked time (show if there are any clock entries)
	if len(item.ClockEntries) > 0 {
		totalDuration := item.GetTotalClockDuration()
		totalHours := int(totalDuration.Hours())
		totalMinutes := int(totalDuration.Minutes()) % 60

		// Format the time display based on magnitude
		var timeStr string
		if totalHours > 0 {
			timeStr = fmt.Sprintf("%dh %dm", totalHours, totalMinutes)
		} else {
			timeStr = fmt.Sprintf("%dm", totalMinutes)
		}

		totalTimeStr := fmt.Sprintf(" (Time: %s)", timeStr)
		totalTimeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("141")) // Purple, similar to scheduled
		b.WriteString(totalTimeStyle.Render(totalTimeStr))
	}

	// Scheduling info
	now := time.Now()
	if item.Scheduled != nil {
		schedStr := fmt.Sprintf(" (Scheduled: %s)", parser.FormatOrgDate(*item.Scheduled))
		if item.Scheduled.Before(now) {
			b.WriteString(overdueStyle.Render(schedStr))
		} else {
			b.WriteString(scheduledStyle.Render(schedStr))
		}
	}
	if item.Deadline != nil {
		deadlineStr := fmt.Sprintf(" (Deadline: %s)", parser.FormatOrgDate(*item.Deadline))
		if item.Deadline.Before(now) {
			b.WriteString(overdueStyle.Render(deadlineStr))
		} else {
			b.WriteString(scheduledStyle.Render(deadlineStr))
		}
	}

	line := b.String()
	if isCursor {
		return cursorStyle.Render(line)
	}
	return line
}
