package model

// UserSettings holds per-user configuration for personalizing the dashboard.
type UserSettings struct {
	ID             string   `json:"id"`
	Latitude       *float64 `json:"latitude"`
	Longitude      *float64 `json:"longitude"`
	CalendarICSURL *string  `json:"calendar_ics_url"`
	Timezone       *string  `json:"timezone"`
}

// NewsCategoryType represents a GNews category available for selection.
type NewsCategoryType struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	SortOrder int    `json:"sort_order"`
}
