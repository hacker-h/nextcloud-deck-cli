package config

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestLoadFromEnvPrefersNormalPassword(t *testing.T) {
	t.Setenv("NEXTCLOUD_BASE_URL", "cloud.example.com/")
	t.Setenv("NEXTCLOUD_USERNAME", " antonia ")
	t.Setenv("NEXTCLOUD_PASSWORD", " pw ")
	t.Setenv("NEXTCLOUD_APP_PASSWORD", "app")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}
	if cfg.BaseURL != "https://cloud.example.com" || cfg.Username != "antonia" || cfg.Password != "pw" {
		t.Fatalf("cfg = %#v", cfg)
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

	_, err := LoadFromEnv()
	if err == nil {
		t.Fatal("expected error for missing env")
	}
	var missing MissingEnvError
	if !reflect.TypeOf(err).AssignableTo(reflect.TypeOf(missing)) {
		t.Fatalf("error type = %T, want MissingEnvError", err)
	}
	for _, want := range []string{"NEXTCLOUD_BASE_URL", "NEXTCLOUD_USERNAME", "NEXTCLOUD_PASSWORD or NEXTCLOUD_APP_PASSWORD"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error = %q, missing %q", err.Error(), want)
		}
	}
}

func TestLoadFromEnv_AppPasswordFallback(t *testing.T) {
	t.Setenv("NEXTCLOUD_BASE_URL", "https://cloud.example.com")
	t.Setenv("NEXTCLOUD_USERNAME", "antonia")
	t.Setenv("NEXTCLOUD_PASSWORD", "")
	t.Setenv("NEXTCLOUD_APP_PASSWORD", " app-pw ")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}
	if cfg.Password != "app-pw" {
		t.Fatalf("Password = %q", cfg.Password)
	}
}

func TestLoadFromEnv_TrimWhitespace(t *testing.T) {
	t.Setenv("NEXTCLOUD_BASE_URL", " https://cloud.example.com/root/ ")
	t.Setenv("NEXTCLOUD_USERNAME", " antonia ")
	t.Setenv("NEXTCLOUD_PASSWORD", " pw ")

	cfg, err := LoadFromEnv()
	if err != nil {
		t.Fatalf("LoadFromEnv() error = %v", err)
	}
	if cfg.BaseURL != "https://cloud.example.com/root" || cfg.Username != "antonia" || cfg.Password != "pw" {
		t.Fatalf("cfg = %#v", cfg)
	}
}

func TestNormalizeBaseURL_NoScheme(t *testing.T) {
	if got := normalizeBaseURL("cloud.example.com"); got != "https://cloud.example.com" {
		t.Fatalf("normalizeBaseURL() = %q", got)
	}
}

func TestNormalizeBaseURL_TrailingSlash(t *testing.T) {
	if got := normalizeBaseURL("https://cloud.example.com/root/"); got != "https://cloud.example.com/root" {
		t.Fatalf("normalizeBaseURL() = %q", got)
	}
}

func TestNormalizeBaseURL_Empty(t *testing.T) {
	if got := normalizeBaseURL(" \t "); got != "" {
		t.Fatalf("normalizeBaseURL() = %q", got)
	}
}

func TestNormalizeBaseURL_WithPath(t *testing.T) {
	if got := normalizeBaseURL("cloud.example.com/nextcloud/"); got != "https://cloud.example.com/nextcloud" {
		t.Fatalf("normalizeBaseURL() = %q", got)
	}
}

func TestNormalizeBaseURL_HttpPreserved(t *testing.T) {
	if got := normalizeBaseURL("http://cloud.example.com/"); got != "http://cloud.example.com" {
		t.Fatalf("normalizeBaseURL() = %q", got)
	}
}

func TestPasswordFromEnv_PrefersNormal(t *testing.T) {
	t.Setenv("NEXTCLOUD_PASSWORD", " normal ")
	t.Setenv("NEXTCLOUD_APP_PASSWORD", "app")
	if got := passwordFromEnv(); got != "normal" {
		t.Fatalf("passwordFromEnv() = %q", got)
	}
}

func TestPasswordFromEnv_FallsBackToApp(t *testing.T) {
	t.Setenv("NEXTCLOUD_PASSWORD", " ")
	t.Setenv("NEXTCLOUD_APP_PASSWORD", " app ")
	if got := passwordFromEnv(); got != "app" {
		t.Fatalf("passwordFromEnv() = %q", got)
	}
}
