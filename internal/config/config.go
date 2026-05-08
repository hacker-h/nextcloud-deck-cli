package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	BaseURL  string
	Username string
	Password string
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
