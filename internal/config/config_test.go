package config

import (
	"encoding/json"
	"os"
	"path/filepath"
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
	if !reflect.TypeOf(err).AssignableTo(reflect.TypeFor[MissingEnvError]()) {
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

func TestLoadFromEnvRejectsExternalHTTP(t *testing.T) {
	t.Setenv("NEXTCLOUD_BASE_URL", "http://cloud.example.com")
	t.Setenv("NEXTCLOUD_USERNAME", "antonia")
	t.Setenv("NEXTCLOUD_PASSWORD", "pw")

	_, err := LoadFromEnv()
	if err == nil {
		t.Fatal("expected error for unsafe base URL")
	}
	if !strings.Contains(err.Error(), "http://cloud.example.com") {
		t.Fatalf("error = %v", err)
	}
}

func TestLoadFromEnvRejectsMissingHost(t *testing.T) {
	t.Setenv("NEXTCLOUD_BASE_URL", "https:///root")
	t.Setenv("NEXTCLOUD_USERNAME", "antonia")
	t.Setenv("NEXTCLOUD_PASSWORD", "pw")

	_, err := LoadFromEnv()
	if err == nil {
		t.Fatal("expected error for missing URL host")
	}
	if !strings.Contains(err.Error(), "host is required") {
		t.Fatalf("error = %v", err)
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

func TestSaveToPathAndLoadFromSources(t *testing.T) {
	clearConfigEnv(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	want := Config{BaseURL: "https://cloud.example.com/root", Username: "antonia", Password: "pw"}
	if err := want.SaveToPath(path); err != nil {
		t.Fatalf("SaveToPath() error = %v", err)
	}
	if info, err := os.Stat(dir); err != nil {
		t.Fatalf("Stat(dir) error = %v", err)
	} else if got := info.Mode().Perm(); got != 0o700 {
		t.Fatalf("dir mode = %o, want 700", got)
	}
	if info, err := os.Stat(path); err != nil {
		t.Fatalf("Stat(path) error = %v", err)
	} else if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("file mode = %o, want 600", got)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if got := string(data); got != "{\"base_url\":\"https://cloud.example.com/root\",\"username\":\"antonia\",\"app_password\":\"pw\"}\n" {
		t.Fatalf("saved config = %q", got)
	}
	got, err := LoadFromSources(path)
	if err != nil {
		t.Fatalf("LoadFromSources() error = %v", err)
	}
	if got.BaseURL != want.BaseURL || got.Username != want.Username || got.Password != want.Password {
		t.Fatalf("loaded config = %#v", got)
	}
	if got.Timeout != DefaultTimeout {
		t.Fatalf("Timeout = %s, want %s", got.Timeout, DefaultTimeout)
	}
}

func TestLoadFromSourcesEnvOverridesSavedConfig(t *testing.T) {
	clearConfigEnv(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := (Config{BaseURL: "https://saved.example.com", Username: "saved", Password: "saved-pw"}).SaveToPath(path); err != nil {
		t.Fatalf("SaveToPath() error = %v", err)
	}
	t.Setenv("NEXTCLOUD_BASE_URL", "cloud.example.com/root/")
	t.Setenv("NEXTCLOUD_USERNAME", " env-user ")
	t.Setenv("NEXTCLOUD_PASSWORD", "")
	t.Setenv("NEXTCLOUD_APP_PASSWORD", " env-pw ")
	t.Setenv("DECK_TIMEOUT", "5m")

	got, err := LoadFromSources(path)
	if err != nil {
		t.Fatalf("LoadFromSources() error = %v", err)
	}
	if got.BaseURL != "https://cloud.example.com/root" || got.Username != "env-user" || got.Password != "env-pw" {
		t.Fatalf("loaded config = %#v", got)
	}
	if got.Timeout != 5*time.Minute {
		t.Fatalf("Timeout = %s, want 5m", got.Timeout)
	}
}

func TestSaveProfileToPathAndLoadProfileFromSources(t *testing.T) {
	clearConfigEnv(t)
	path := filepath.Join(t.TempDir(), "config.json")
	defaultConfig := Config{BaseURL: "https://default.example.com", Username: "default", Password: "default-pw"}
	workConfig := Config{BaseURL: "https://work.example.com/root", Username: "work", Password: "work-pw"}
	if err := defaultConfig.SaveToPath(path); err != nil {
		t.Fatalf("SaveToPath() error = %v", err)
	}
	if err := workConfig.SaveProfileToPath(path, "work"); err != nil {
		t.Fatalf("SaveProfileToPath() error = %v", err)
	}

	gotDefault, err := LoadFromSources(path)
	if err != nil {
		t.Fatalf("LoadFromSources() error = %v", err)
	}
	if gotDefault.BaseURL != defaultConfig.BaseURL || gotDefault.Username != defaultConfig.Username || gotDefault.Password != defaultConfig.Password {
		t.Fatalf("default config = %#v", gotDefault)
	}
	gotAlias, err := LoadProfileFromSources(path, "default")
	if err != nil {
		t.Fatalf("LoadProfileFromSources(default) error = %v", err)
	}
	if gotAlias.BaseURL != defaultConfig.BaseURL || gotAlias.Username != defaultConfig.Username || gotAlias.Password != defaultConfig.Password {
		t.Fatalf("default alias config = %#v", gotAlias)
	}
	gotProfile, err := LoadProfileFromSources(path, "work")
	if err != nil {
		t.Fatalf("LoadProfileFromSources(work) error = %v", err)
	}
	if gotProfile.BaseURL != workConfig.BaseURL || gotProfile.Username != workConfig.Username || gotProfile.Password != workConfig.Password {
		t.Fatalf("profile config = %#v", gotProfile)
	}
}

func TestLoadProfileFromSourcesEnvOverridesSelectedProfile(t *testing.T) {
	clearConfigEnv(t)
	path := filepath.Join(t.TempDir(), "config.json")
	if err := (Config{BaseURL: "https://work.example.com", Username: "work", Password: "work-pw"}).SaveProfileToPath(path, "work"); err != nil {
		t.Fatalf("SaveProfileToPath() error = %v", err)
	}
	t.Setenv("NEXTCLOUD_BASE_URL", "env.example.com/root/")
	t.Setenv("NEXTCLOUD_USERNAME", " env-user ")
	t.Setenv("NEXTCLOUD_APP_PASSWORD", " env-pw ")

	got, err := LoadProfileFromSources(path, "work")
	if err != nil {
		t.Fatalf("LoadProfileFromSources() error = %v", err)
	}
	if got.BaseURL != "https://env.example.com/root" || got.Username != "env-user" || got.Password != "env-pw" {
		t.Fatalf("config = %#v", got)
	}
}

func TestSaveProfileToPathReplacesDuplicateProfile(t *testing.T) {
	clearConfigEnv(t)
	path := filepath.Join(t.TempDir(), "config.json")
	if err := (Config{BaseURL: "https://old.example.com", Username: "old", Password: "old-pw"}).SaveProfileToPath(path, "work"); err != nil {
		t.Fatalf("SaveProfileToPath(old) error = %v", err)
	}
	if err := (Config{BaseURL: "https://new.example.com", Username: "new", Password: "new-pw"}).SaveProfileToPath(path, "work"); err != nil {
		t.Fatalf("SaveProfileToPath(new) error = %v", err)
	}

	got, err := LoadProfileFromSources(path, "work")
	if err != nil {
		t.Fatalf("LoadProfileFromSources() error = %v", err)
	}
	if got.BaseURL != "https://new.example.com" || got.Username != "new" || got.Password != "new-pw" {
		t.Fatalf("config = %#v", got)
	}
}

func TestLoadProfileFromSourcesMissingProfile(t *testing.T) {
	clearConfigEnv(t)
	path := filepath.Join(t.TempDir(), "config.json")
	if err := (Config{BaseURL: "https://default.example.com", Username: "default", Password: "pw"}).SaveToPath(path); err != nil {
		t.Fatalf("SaveToPath() error = %v", err)
	}

	_, err := LoadProfileFromSources(path, "missing")
	if err == nil || !strings.Contains(err.Error(), `profile "missing" not found`) {
		t.Fatalf("error = %v", err)
	}
}

func TestSaveProfileToPathRejectsInvalidProfileNames(t *testing.T) {
	for _, profile := range []string{"work/personal", `work\personal`, "work\x00personal"} {
		err := (Config{BaseURL: "https://work.example.com", Username: "work", Password: "work-pw"}).SaveProfileToPath(filepath.Join(t.TempDir(), "config.json"), profile)
		if err == nil {
			t.Fatalf("SaveProfileToPath(%q) succeeded", profile)
		}
	}
}

func TestListProfilesFromPathOmitsPasswords(t *testing.T) {
	clearConfigEnv(t)
	path := filepath.Join(t.TempDir(), "config.json")
	if err := (Config{BaseURL: "https://default.example.com", Username: "default", Password: "default-secret"}).SaveToPath(path); err != nil {
		t.Fatalf("SaveToPath() error = %v", err)
	}
	if err := (Config{BaseURL: "https://work.example.com", Username: "work", Password: "work-secret"}).SaveProfileToPath(path, "work"); err != nil {
		t.Fatalf("SaveProfileToPath(work) error = %v", err)
	}

	profiles, err := ListProfilesFromPath(path)
	if err != nil {
		t.Fatalf("ListProfilesFromPath() error = %v", err)
	}
	want := []ProfileSummary{{Name: "default", BaseURL: "https://default.example.com", Username: "default"}, {Name: "work", BaseURL: "https://work.example.com", Username: "work"}}
	if !reflect.DeepEqual(profiles, want) {
		t.Fatalf("profiles = %#v, want %#v", profiles, want)
	}
	data, err := json.Marshal(profiles)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if strings.Contains(string(data), "secret") {
		t.Fatalf("profiles leaked password: %s", string(data))
	}
}

func TestLoadFromSourcesRejectsExternalHTTP(t *testing.T) {
	clearConfigEnv(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := (Config{BaseURL: "http://cloud.example.com", Username: "saved", Password: "saved-pw"}).SaveToPath(path); err != nil {
		t.Fatalf("SaveToPath() error = %v", err)
	}

	_, err := LoadFromSources(path)
	if err == nil {
		t.Fatal("expected error for unsafe base URL")
	}
	if !strings.Contains(err.Error(), "http://cloud.example.com") {
		t.Fatalf("error = %v", err)
	}
}

func TestLoadFromSourcesAllowsLocalhostHTTP(t *testing.T) {
	clearConfigEnv(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	want := Config{BaseURL: "http://localhost:8080/root", Username: "saved", Password: "saved-pw"}
	if err := want.SaveToPath(path); err != nil {
		t.Fatalf("SaveToPath() error = %v", err)
	}

	got, err := LoadFromSources(path)
	if err != nil {
		t.Fatalf("LoadFromSources() error = %v", err)
	}
	if got.BaseURL != want.BaseURL {
		t.Fatalf("BaseURL = %q, want %q", got.BaseURL, want.BaseURL)
	}
}

func TestLoadFromSourcesAllowsLoopbackHTTP(t *testing.T) {
	clearConfigEnv(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	want := Config{BaseURL: "http://127.0.0.1:8080/root", Username: "saved", Password: "saved-pw"}
	if err := want.SaveToPath(path); err != nil {
		t.Fatalf("SaveToPath() error = %v", err)
	}

	got, err := LoadFromSources(path)
	if err != nil {
		t.Fatalf("LoadFromSources() error = %v", err)
	}
	if got.BaseURL != want.BaseURL {
		t.Fatalf("BaseURL = %q, want %q", got.BaseURL, want.BaseURL)
	}
}

func TestLoadFromSourcesAllowsHTTPS(t *testing.T) {
	clearConfigEnv(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	want := Config{BaseURL: "https://cloud.example.com/root", Username: "saved", Password: "saved-pw"}
	if err := want.SaveToPath(path); err != nil {
		t.Fatalf("SaveToPath() error = %v", err)
	}

	got, err := LoadFromSources(path)
	if err != nil {
		t.Fatalf("LoadFromSources() error = %v", err)
	}
	if got.BaseURL != want.BaseURL {
		t.Fatalf("BaseURL = %q, want %q", got.BaseURL, want.BaseURL)
	}
}

func TestLoadFromSourcesMissingConfig(t *testing.T) {
	clearConfigEnv(t)
	_, err := LoadFromSources(filepath.Join(t.TempDir(), "missing.json"))
	if err == nil {
		t.Fatal("expected error for missing config")
	}
	if !reflect.TypeOf(err).AssignableTo(reflect.TypeFor[MissingEnvError]()) {
		t.Fatalf("error type = %T, want MissingEnvError", err)
	}
}

func clearConfigEnv(t *testing.T) {
	t.Helper()
	t.Setenv("NEXTCLOUD_BASE_URL", "")
	t.Setenv("NEXTCLOUD_USERNAME", "")
	t.Setenv("NEXTCLOUD_PASSWORD", "")
	t.Setenv("NEXTCLOUD_APP_PASSWORD", "")
	t.Setenv("DECK_TIMEOUT", "")
	t.Setenv("DECK_PROFILE", "")
}
