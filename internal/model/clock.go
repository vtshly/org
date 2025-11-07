package model

import "time"

// ClockEntry represents a single clock entry
type ClockEntry struct {
	Start time.Time
	End   *time.Time // nil if currently clocked in
}
