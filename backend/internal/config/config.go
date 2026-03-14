package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port            string
	GNewsAPIKey     string
	FinnhubAPIKey   string
	APINinjasAPIKey string
	ICSCalendarURL  string
	Latitude        float64
	Longitude       float64
	Timezone        *time.Location
	SQLitePath      string

	GoogleClientID     string
	GoogleClientSecret string
	GoogleCallbackURL  string // default: http://localhost:8080/api/auth/callback
	SessionSecret      string // HMAC key for session cookies; should be 32+ random bytes
	FrontendURL        string // where to redirect after successful login; default: http://localhost:5173
}

func Load() *Config {
	loadDotEnv()
	lat := parseFloat(os.Getenv("LATITUDE"), 37.7749)
	lon := parseFloat(os.Getenv("LONGITUDE"), -122.4194)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	tz := time.Local
	if tzName := os.Getenv("TIMEZONE"); tzName != "" {
		if loc, err := time.LoadLocation(tzName); err == nil {
			tz = loc
		}
	}

	callbackURL := os.Getenv("GOOGLE_CALLBACK_URL")
	if callbackURL == "" {
		callbackURL = "http://localhost:8080/api/auth/callback"
	}
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	return &Config{
		Port:            port,
		GNewsAPIKey:     os.Getenv("GNEWS_API_KEY"),
		FinnhubAPIKey:   os.Getenv("FINNHUB_API_KEY"),
		APINinjasAPIKey: os.Getenv("API_NINJAS_API_KEY"),
		ICSCalendarURL:  os.Getenv("CALENDAR_ICS_URL"),
		Latitude:        lat,
		Longitude:       lon,
		Timezone:        tz,
		SQLitePath:      sqlitePath(os.Getenv("SQLITE_PATH")),

		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleCallbackURL:  callbackURL,
		SessionSecret:      os.Getenv("SESSION_SECRET"),
		FrontendURL:        frontendURL,
	}
}

// loadDotEnv reads a .env file and sets any keys not already present in the
// environment. It checks the current directory first, then the parent (so the
// server can be run from either the repo root or the backend/ subdirectory).
func loadDotEnv() {
	for _, path := range []string{".env", "../.env"} {
		if parseDotEnv(path) {
			return
		}
	}
}

func parseDotEnv(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key != "" && os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
	return true
}

func sqlitePath(s string) string {
	if s == "" {
		return "dashboard.db"
	}
	return s
}

func parseFloat(s string, def float64) float64 {
	if s == "" {
		return def
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return def
	}
	return v
}
