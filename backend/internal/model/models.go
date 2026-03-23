package model

// DashboardResponse is the aggregated response for GET /api/dashboard
type DashboardResponse struct {
	Weather    *WeatherData    `json:"weather"`    // null = unavailable
	Calendar   []CalendarEvent `json:"calendar"`
	Tasks      []Task          `json:"tasks"`
	TasksTotal int             `json:"tasksTotal"`
	Stocks     []StockQuote    `json:"stocks"`     // nil slice = null
	Meta       *MetaData       `json:"meta"`       // null = unavailable
}

// WeatherData holds current weather conditions
type WeatherData struct {
	Temp      float64          `json:"temp"`
	High      float64          `json:"high"`
	Low       float64          `json:"low"`
	Condition string           `json:"condition"`
	Icon      string           `json:"icon"`
	Humidity  int              `json:"humidity"`
	WindSpeed float64          `json:"windSpeed"`
	UVIndex   float64          `json:"uvIndex"`
	AQI       int              `json:"aqi"`
	AQILabel  string           `json:"aqiLabel"`
	Hourly    []HourlyForecast `json:"hourly"`
}

// HourlyForecast is a single hour in the weather forecast
type HourlyForecast struct {
	Time string  `json:"time"`
	Temp float64 `json:"temp"`
	Icon string  `json:"icon"`
}

// CalendarEvent is a single calendar event
type CalendarEvent struct {
	Time     string `json:"time"`
	Title    string `json:"title"`
	Color    string `json:"color"`
	Duration string `json:"duration"`
}

// Task is a to-do item
type Task struct {
	ID       string `json:"id"`
	Text     string `json:"text"`
	Done     bool   `json:"done"`
	Priority string `json:"priority"` // "high", "medium", "low"
}

// NewsItem is a single news article
type NewsItem struct {
	Title  string `json:"title"`
	Source string `json:"source"`
	Time   string `json:"time"`
	URL    string `json:"url"`
}

// NewsCategory is a named group of news articles for one GNews topic.
type NewsCategory struct {
	Name  string     `json:"name"`
	Items []NewsItem `json:"items"`
}

// StockQuote is a single stock/crypto quote
type StockQuote struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
	Change float64 `json:"change"`
	Pct    float64 `json:"pct"`
}

// SymbolSearchResult is a single result from a Finnhub symbol search
type SymbolSearchResult struct {
	Symbol      string `json:"symbol"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

// MetaData holds sunrise/sunset and quote info
type MetaData struct {
	Sunrise  string `json:"sunrise"`
	Sunset   string `json:"sunset"`
	Daylight string `json:"daylight"`
	Quote    Quote  `json:"quote"`
}

// Quote is a daily inspirational quote
type Quote struct {
	Text   string `json:"text"`
	Author string `json:"author"`
}

// TaskCreate holds fields for creating a new task.
type TaskCreate struct {
	ID         string
	UserID     string
	Text       string
	Done       bool
	PriorityID string
}

// TaskUpdate holds optional fields for updating a task.
type TaskUpdate struct {
	Done       *bool
	Text       *string
	PriorityID *string
}
