package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/daily-dashboard/backend/internal/model"
	"github.com/daily-dashboard/backend/internal/repository"
)

// ErrSettingsNotFound is returned when user settings do not exist.
var ErrSettingsNotFound = errors.New("user settings not found")

// ErrSettingsValidation is returned when settings input fails validation.
var ErrSettingsValidation = errors.New("settings validation failed")

// ErrCategoryNotFound is returned when an invalid news category ID is provided.
var ErrCategoryNotFound = errors.New("news category not found")

// UserSettingsService manages user settings via a repository.
type UserSettingsService struct {
	repo repository.UserSettingsRepository
	enc  *EncryptionService // nil means no encryption (dev mode)
}

// NewUserSettingsService creates a new UserSettingsService backed by the given repository.
// enc may be nil to disable encryption of sensitive fields.
func NewUserSettingsService(repo repository.UserSettingsRepository, enc *EncryptionService) *UserSettingsService {
	return &UserSettingsService{repo: repo, enc: enc}
}

// Get returns the settings for the given user, or nil if no settings have been configured yet.
func (s *UserSettingsService) Get(ctx context.Context, userID string) (*model.UserSettings, error) {
	row, err := s.repo.Get(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrSettingsNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user settings: %w", err)
	}
	m := settingsRowToModel(row)
	if err := s.decryptSensitiveFields(&m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Upsert creates or updates settings for the given user, returning the final state.
func (s *UserSettingsService) Upsert(ctx context.Context, userID string, u repository.UserSettingsUpsert) (model.UserSettings, error) {
	if u.Timezone != nil {
		trimmed := strings.TrimSpace(*u.Timezone)
		u.Timezone = &trimmed
	}
	if u.CalendarICSURL != nil {
		trimmed := strings.TrimSpace(*u.CalendarICSURL)
		u.CalendarICSURL = &trimmed
		if strings.HasPrefix(trimmed, "enc:") {
			return model.UserSettings{}, fmt.Errorf("%w: calendar URL must not start with \"enc:\"", ErrSettingsValidation)
		}
		if trimmed != "" {
			if err := validateCalendarURL(trimmed); err != nil {
				return model.UserSettings{}, err
			}
		}
		if s.enc != nil && trimmed != "" {
			encrypted, encErr := s.enc.Encrypt(trimmed)
			if encErr != nil {
				return model.UserSettings{}, fmt.Errorf("encrypt calendar URL: %w", encErr)
			}
			u.CalendarICSURL = &encrypted
		}
	}

	row, err := s.repo.Upsert(ctx, userID, u)
	if err != nil {
		return model.UserSettings{}, fmt.Errorf("upsert user settings: %w", err)
	}
	m := settingsRowToModel(row)
	if err := s.decryptSensitiveFields(&m); err != nil {
		return model.UserSettings{}, err
	}
	return m, nil
}

// ListAllCategories returns all available news category types.
func (s *UserSettingsService) ListAllCategories(ctx context.Context) ([]model.NewsCategoryType, error) {
	rows, err := s.repo.ListAllCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all news categories: %w", err)
	}
	result := make([]model.NewsCategoryType, 0, len(rows))
	for _, r := range rows {
		result = append(result, categoryRowToModel(r))
	}
	return result, nil
}

// ListSelectedCategories returns the news categories the user has selected.
func (s *UserSettingsService) ListSelectedCategories(ctx context.Context, userID string) ([]model.NewsCategoryType, error) {
	rows, err := s.repo.ListSelectedCategories(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list selected news categories: %w", err)
	}
	result := make([]model.NewsCategoryType, 0, len(rows))
	for _, r := range rows {
		result = append(result, categoryRowToModel(r))
	}
	return result, nil
}

// SetSelectedCategories replaces the user's selected news categories.
// Returns ErrCategoryNotFound if any provided category ID does not exist.
func (s *UserSettingsService) SetSelectedCategories(ctx context.Context, userID string, categoryIDs []string) error {
	// Validate that all provided category IDs exist.
	all, err := s.repo.ListAllCategories(ctx)
	if err != nil {
		return fmt.Errorf("validate categories: %w", err)
	}
	valid := make(map[string]struct{}, len(all))
	for _, c := range all {
		valid[c.ID] = struct{}{}
	}
	for _, id := range categoryIDs {
		if _, ok := valid[id]; !ok {
			return fmt.Errorf("%w: %q", ErrCategoryNotFound, id)
		}
	}

	if err := s.repo.SetSelectedCategories(ctx, userID, categoryIDs); err != nil {
		return fmt.Errorf("set selected categories: %w", err)
	}
	return nil
}

// validateCalendarURL rejects URLs that could cause SSRF:
// - Only http/https schemes allowed
// - Hostname must be present
// - Literal private/loopback/link-local IPs rejected
// - Known cloud metadata hostnames blocked
func validateCalendarURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("%w: invalid calendar URL", ErrSettingsValidation)
	}
	switch u.Scheme {
	case "https", "http":
		// ok
	default:
		return fmt.Errorf("%w: calendar URL must use https:// or http:// scheme", ErrSettingsValidation)
	}

	hostname := u.Hostname()
	if hostname == "" {
		return fmt.Errorf("%w: calendar URL must include a hostname", ErrSettingsValidation)
	}

	// Reject literal private/loopback/link-local IPs.
	if ip := net.ParseIP(hostname); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
			ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
			return fmt.Errorf("%w: calendar URL must not point to a private or reserved address", ErrSettingsValidation)
		}
	}

	// Block known cloud metadata hostnames.
	lower := strings.ToLower(hostname)
	for _, blocked := range []string{"metadata.google.internal", "metadata.internal", "instance-data"} {
		if lower == blocked {
			return fmt.Errorf("%w: calendar URL must not point to a cloud metadata endpoint", ErrSettingsValidation)
		}
	}

	return nil
}

func (s *UserSettingsService) decryptSensitiveFields(m *model.UserSettings) error {
	if s.enc != nil && m.CalendarICSURL != nil && *m.CalendarICSURL != "" {
		decrypted, err := s.enc.Decrypt(*m.CalendarICSURL)
		if err != nil {
			return fmt.Errorf("decrypt calendar URL: %w", err)
		}
		m.CalendarICSURL = &decrypted
	}
	return nil
}

func settingsRowToModel(r repository.UserSettingsRow) model.UserSettings {
	return model.UserSettings{
		ID:             r.ID,
		Latitude:       r.Latitude,
		Longitude:      r.Longitude,
		CalendarICSURL: r.CalendarICSURL,
		Timezone:       r.Timezone,
	}
}

func categoryRowToModel(r repository.NewsCategoryTypeRow) model.NewsCategoryType {
	return model.NewsCategoryType{
		ID:        r.ID,
		Label:     r.Label,
		SortOrder: r.SortOrder,
	}
}
