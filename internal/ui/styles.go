package ui

import "github.com/charmbracelet/lipgloss"

// Styles for UI rendering
var (
	todoStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("202")) // Orange
	progStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("220")) // Yellow
	blockStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // Red
	doneStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("34"))  // Green
	cursorStyle    = lipgloss.NewStyle().Background(lipgloss.Color("240"))
	titleStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	scheduledStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("141")) // Purple
	overdueStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // Red
	statusStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
	noteStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("246")).Italic(true)
	foldedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
)
