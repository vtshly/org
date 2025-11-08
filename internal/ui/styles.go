package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/rwejlgaard/org/internal/config"
)

// styleMap holds all the styles used in the UI
type styleMap struct {
	todoStyle      lipgloss.Style
	progStyle      lipgloss.Style
	blockStyle     lipgloss.Style
	doneStyle      lipgloss.Style
	cursorStyle    lipgloss.Style
	titleStyle     lipgloss.Style
	scheduledStyle lipgloss.Style
	overdueStyle   lipgloss.Style
	statusStyle    lipgloss.Style
	noteStyle      lipgloss.Style
	foldedStyle    lipgloss.Style
}

// newStyleMapFromConfig creates a styleMap from configuration
func newStyleMapFromConfig(cfg *config.Config) styleMap {
	colors := cfg.Colors

	return styleMap{
		todoStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Todo)),
		progStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Progress)),
		blockStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Blocked)),
		doneStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Done)),
		cursorStyle:    lipgloss.NewStyle().Background(lipgloss.Color(colors.Cursor)),
		titleStyle:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(colors.Title)),
		scheduledStyle: lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Scheduled)),
		overdueStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Overdue)),
		statusStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Status)).Italic(true),
		noteStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Note)).Italic(true),
		foldedStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color(colors.Folded)),
	}
}
