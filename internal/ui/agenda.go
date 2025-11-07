package ui

import (
	"time"

	"github.com/rwejlgaard/org/internal/model"
)

// getAgendaItems returns items with scheduling or deadlines within the next 7 days
func (m uiModel) getAgendaItems() []*model.Item {
	var items []*model.Item
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfWeek := startOfDay.AddDate(0, 0, 7)

	// Get all items regardless of folding for agenda view
	var getAllItems func([]*model.Item)
	getAllItems = func(list []*model.Item) {
		for _, item := range list {
			if item.Scheduled != nil && item.Scheduled.Before(endOfWeek) {
				items = append(items, item)
			}
			if item.Deadline != nil && item.Deadline.Before(endOfWeek) {
				items = append(items, item)
			}
			getAllItems(item.Children)
		}
	}
	getAllItems(m.orgFile.Items)

	return items
}
