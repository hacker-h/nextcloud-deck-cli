package config

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"sort"
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

type ProfileSummary struct {
	Name     string `json:"name"`
	BaseURL  string `json:"base_url,omitempty"`
	Username string `json:"username,omitempty"`
}

type storedConfig struct {
	BaseURL     string                   `json:"base_url,omitempty"`
	Username    string                   `json:"username,omitempty"`
	AppPassword string                   `json:"app_password,omitempty"`
	Profiles    map[string]storedProfile `json:"profiles,omitempty"`
}

type storedProfile struct {
	BaseURL     string `json:"base_url"`
	Username    string `json:"username"`
	AppPassword string `json:"app_password"`
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
	if err := validateBaseURL(cfg.BaseURL); err != nil {
		return Config{}, err
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

func Load() (Config, error) {
	path, err := defaultConfigPath()
	if err != nil {
		return Config{}, err
	}
	return LoadFromSources(path)
}

func LoadProfile(profile string) (Config, error) {
	path, err := defaultConfigPath()
	if err != nil {
		return Config{}, err
	}
	return LoadProfileFromSources(path, profile)
}

func LoadFromSources(path string) (Config, error) {
	return LoadProfileFromSources(path, "")
}

func LoadProfileFromSources(path, profile string) (Config, error) {
	cfg := Config{Timeout: DefaultTimeout}
	if path != "" {
		saved, err := loadSaved(path, profile)
		if err != nil {
			return Config{}, err
		}
		cfg = saved
		if cfg.Timeout <= 0 {
			cfg.Timeout = DefaultTimeout
		}
	}
	overlayEnv(&cfg)
	if raw := strings.TrimSpace(os.Getenv("DECK_TIMEOUT")); raw != "" {
		timeout, err := parseTimeout(raw)
		if err != nil {
			return Config{}, err
		}
		cfg.Timeout = timeout
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = DefaultTimeout
	}
	if err := validateBaseURL(cfg.BaseURL); err != nil {
		return Config{}, err
	}
	if missing := missingConfigValues(cfg); len(missing) > 0 {
		return Config{}, MissingEnvError{Names: missing}
	}
	return cfg, nil
}

func (cfg Config) Save() error {
	path, err := defaultConfigPath()
	if err != nil {
		return err
	}
	return cfg.SaveToPath(path)
}

func (cfg Config) SaveToPath(path string) error {
	if path == "" {
		return fmt.Errorf("config path is required")
	}
	stored, err := loadStoredIfExists(path)
	if err != nil {
		return err
	}
	stored.BaseURL = cfg.BaseURL
	stored.Username = cfg.Username
	stored.AppPassword = cfg.Password
	return saveStored(path, stored)
}

func (cfg Config) SaveProfile(profile string) error {
	path, err := defaultConfigPath()
	if err != nil {
		return err
	}
	return cfg.SaveProfileToPath(path, profile)
}

func (cfg Config) SaveProfileToPath(path, profile string) error {
	profile, err := normalizeProfileName(profile)
	if err != nil {
		return err
	}
	if profile == "" {
		return cfg.SaveToPath(path)
	}
	if path == "" {
		return fmt.Errorf("config path is required")
	}
	stored, err := loadStoredIfExists(path)
	if err != nil {
		return err
	}
	if stored.Profiles == nil {
		stored.Profiles = make(map[string]storedProfile)
	}
	stored.Profiles[profile] = storedProfile{BaseURL: cfg.BaseURL, Username: cfg.Username, AppPassword: cfg.Password}
	return saveStored(path, stored)
}

func ListProfiles() ([]ProfileSummary, error) {
	path, err := defaultConfigPath()
	if err != nil {
		return nil, err
	}
	return ListProfilesFromPath(path)
}

func ListProfilesFromPath(path string) ([]ProfileSummary, error) {
	stored, err := loadStoredIfExists(path)
	if err != nil {
		return nil, err
	}
	profiles := make([]ProfileSummary, 0, len(stored.Profiles)+1)
	if stored.BaseURL != "" || stored.Username != "" || stored.AppPassword != "" {
		profiles = append(profiles, ProfileSummary{Name: "default", BaseURL: normalizeBaseURL(stored.BaseURL), Username: strings.TrimSpace(stored.Username)})
	}
	names := make([]string, 0, len(stored.Profiles))
	for name := range stored.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		profile := stored.Profiles[name]
		profiles = append(profiles, ProfileSummary{Name: name, BaseURL: normalizeBaseURL(profile.BaseURL), Username: strings.TrimSpace(profile.Username)})
	}
	return profiles, nil
}

func saveStored(path string, stored storedConfig) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return fmt.Errorf("set config dir permissions: %w", err)
	}
	temp, err := os.CreateTemp(dir, "config-*.tmp")
	if err != nil {
		return fmt.Errorf("create config temp file: %w", err)
	}
	tempPath := temp.Name()
	defer func() {
		_ = temp.Close()
		_ = os.Remove(tempPath)
	}()
	if err := temp.Chmod(0o600); err != nil {
		return fmt.Errorf("set config file permissions: %w", err)
	}
	if err := json.NewEncoder(temp).Encode(stored); err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close config temp file: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace config file: %w", err)
	}
	return nil
}

func defaultConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "nextcloud-deck-cli", "config.json"), nil
}

func loadSaved(path, profile string) (Config, error) {
	profile, err := normalizeProfileName(profile)
	if err != nil {
		return Config{}, err
	}
	stored, err := loadStoredIfExists(path)
	if err != nil {
		return Config{}, err
	}
	if profile != "" {
		saved, ok := stored.Profiles[profile]
		if !ok {
			return Config{}, fmt.Errorf("profile %q not found", profile)
		}
		return configFromStoredProfile(saved), nil
	}
	return configFromStoredProfile(storedProfile{BaseURL: stored.BaseURL, Username: stored.Username, AppPassword: stored.AppPassword}), nil
}

func loadStoredIfExists(path string) (storedConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return storedConfig{}, nil
		}
		return storedConfig{}, fmt.Errorf("read config file: %w", err)
	}
	var stored storedConfig
	if err := json.Unmarshal(data, &stored); err != nil {
		return storedConfig{}, fmt.Errorf("decode config file: %w", err)
	}
	return stored, nil
}

func configFromStoredProfile(stored storedProfile) Config {
	return Config{
		BaseURL:  normalizeBaseURL(stored.BaseURL),
		Username: strings.TrimSpace(stored.Username),
		Password: strings.TrimSpace(stored.AppPassword),
		Timeout:  DefaultTimeout,
	}
}

func normalizeProfileName(profile string) (string, error) {
	profile = strings.TrimSpace(profile)
	if profile == "" || profile == "default" {
		return "", nil
	}
	if strings.ContainsAny(profile, `/\\`) {
		return "", fmt.Errorf("invalid profile %q: path separators are not allowed", profile)
	}
	for _, r := range profile {
		if r < 0x20 || r == 0x7f {
			return "", fmt.Errorf("invalid profile %q: control characters are not allowed", profile)
		}
	}
	return profile, nil
}

func overlayEnv(cfg *Config) {
	if baseURL := normalizeBaseURL(os.Getenv("NEXTCLOUD_BASE_URL")); baseURL != "" {
		cfg.BaseURL = baseURL
	}
	if username := strings.TrimSpace(os.Getenv("NEXTCLOUD_USERNAME")); username != "" {
		cfg.Username = username
	}
	if password := passwordFromEnv(); password != "" {
		cfg.Password = password
	}
}

func missingConfigValues(cfg Config) []string {
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
	return missing
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

func NormalizeBaseURL(raw string) string {
	return normalizeBaseURL(raw)
}

func ValidateBaseURL(raw string) error {
	return validateBaseURL(raw)
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

func validateBaseURL(raw string) error {
	if raw == "" {
		return nil
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid NEXTCLOUD_BASE_URL %q: %w", raw, err)
	}
	host := parsed.Hostname()
	if host == "" {
		return fmt.Errorf("invalid NEXTCLOUD_BASE_URL %q: host is required", raw)
	}
	switch strings.ToLower(parsed.Scheme) {
	case "https":
		return nil
	case "http":
		if isLocalhost(host) {
			return nil
		}
	}
	return fmt.Errorf("invalid NEXTCLOUD_BASE_URL %q: use https, or http only for localhost/loopback", raw)
}

func isLocalhost(host string) bool {
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}
