package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadFromEnvPrefersNormalPassword(t *testing.T) {
	t.Setenv("NEXTCLOUD_BASE_URL", "cloud.example.com/")
	t.Setenv("NEXTCLOUD_USERNAME", "antonia")
	t.Setenv("NEXTCLOUD_PASSWORD", "pw")
	t.Setenv("NEXTCLOUD_APP_PASSWORD", "app-pw")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}
	if cfg.BaseURL != "https://cloud.example.com" {
		t.Fatalf("BaseURL = %q", cfg.BaseURL)
	}
	if cfg.Password != "pw" {
		t.Fatalf("Password = %q", cfg.Password)
	}
	if cfg.Timeout != DefaultTimeout {
		t.Fatalf("Timeout = %s, want %s", cfg.Timeout, DefaultTimeout)
	}
}

func TestLoadFromEnvReadsDeckTimeout(t *testing.T) {
	t.Setenv("NEXTCLOUD_BASE_URL", "cloud.example.com")
	t.Setenv("NEXTCLOUD_USERNAME", "antonia")
	t.Setenv("NEXTCLOUD_PASSWORD", "pw")
	t.Setenv("DECK_TIMEOUT", "5m")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}
	if cfg.Timeout != 5*time.Minute {
		t.Fatalf("Timeout = %s, want 5m", cfg.Timeout)
	}
}

func TestLoadFromEnvRejectsInvalidDeckTimeout(t *testing.T) {
	t.Setenv("NEXTCLOUD_BASE_URL", "cloud.example.com")
	t.Setenv("NEXTCLOUD_USERNAME", "antonia")
	t.Setenv("NEXTCLOUD_PASSWORD", "pw")
	t.Setenv("DECK_TIMEOUT", "0s")

	_, err := LoadFromEnv()
	if err == nil {
		t.Fatal("expected error for invalid timeout")
	}
	if !strings.Contains(err.Error(), "DECK_TIMEOUT") {
		t.Fatalf("error = %v, want DECK_TIMEOUT", err)
	}
}

func TestLoadFromEnvRequiresValues(t *testing.T) {
	t.Setenv("NEXTCLOUD_BASE_URL", "")
	t.Setenv("NEXTCLOUD_USERNAME", "")
	t.Setenv("NEXTCLOUD_PASSWORD", "")
	t.Setenv("NEXTCLOUD_APP_PASSWORD", "")

	if _, err := LoadFromEnv(); err == nil {
		t.Fatal("expected error for missing env")
	}
}

// --- Config edge cases ---

func TestLoadFromEnv_AppPasswordFallback(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestLoadFromEnv_TrimWhitespace(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestNormalizeBaseURL_NoScheme(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestNormalizeBaseURL_TrailingSlash(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestNormalizeBaseURL_Empty(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestNormalizeBaseURL_WithPath(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestNormalizeBaseURL_HttpPreserved(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestPasswordFromEnv_PrefersNormal(t *testing.T) {
	t.Skip("TODO: implement")
}

func TestPasswordFromEnv_FallsBackToApp(t *testing.T) {
	t.Skip("TODO: implement")
}
