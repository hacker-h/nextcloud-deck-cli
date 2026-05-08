package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"
)

const DefaultTimeout = 90 * time.Second

type Config struct {
	BaseURL  string
	Username string
	Password string
	Timeout  time.Duration
}

type MissingEnvError struct {
	Names []string
}

func (e MissingEnvError) Error() string {
	return fmt.Sprintf("missing env: %s", strings.Join(e.Names, ", "))
}

func LoadFromEnv() (Config, error) {
	cfg := Config{
		BaseURL:  normalizeBaseURL(os.Getenv("NEXTCLOUD_BASE_URL")),
		Username: strings.TrimSpace(os.Getenv("NEXTCLOUD_USERNAME")),
		Password: passwordFromEnv(),
		Timeout:  DefaultTimeout,
	}
	if raw := strings.TrimSpace(os.Getenv("DECK_TIMEOUT")); raw != "" {
		timeout, err := parseTimeout(raw)
		if err != nil {
			return Config{}, err
		}
		cfg.Timeout = timeout
	}

	var missing []string
	if cfg.BaseURL == "" {
		missing = append(missing, "NEXTCLOUD_BASE_URL")
	}
	if cfg.Username == "" {
		missing = append(missing, "NEXTCLOUD_USERNAME")
	}
	if cfg.Password == "" {
		missing = append(missing, "NEXTCLOUD_PASSWORD or NEXTCLOUD_APP_PASSWORD")
	}
	if len(missing) > 0 {
		return Config{}, MissingEnvError{Names: missing}
	}

	return cfg, nil
}

func parseTimeout(raw string) (time.Duration, error) {
	timeout, err := time.ParseDuration(strings.TrimSpace(raw))
	if err != nil {
		return 0, fmt.Errorf("invalid DECK_TIMEOUT %q: %w", raw, err)
	}
	if timeout <= 0 {
		return 0, fmt.Errorf("invalid DECK_TIMEOUT %q: must be greater than 0", raw)
	}
	return timeout, nil
}

func passwordFromEnv() string {
	if value := strings.TrimSpace(os.Getenv("NEXTCLOUD_PASSWORD")); value != "" {
		return value
	}
	return strings.TrimSpace(os.Getenv("NEXTCLOUD_APP_PASSWORD"))
}

func normalizeBaseURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	if !strings.Contains(trimmed, "://") {
		trimmed = "https://" + trimmed
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return strings.TrimRight(trimmed, "/")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	return parsed.String()
}
