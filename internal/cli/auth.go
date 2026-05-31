package cli

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/hacker-h/nextcloud-deck-api/internal/config"
)

func isAuthSetupCommand(args []string) bool {
	return len(args) >= 2 && args[0] == "auth" && args[1] == "setup"
}

func isAuthProfilesCommand(args []string) bool {
	return len(args) >= 2 && args[0] == "auth" && args[1] == "profiles"
}

func runAuth(_ *runtime, args []string) error {
	if len(args) == 0 {
		return validationError("auth requires setup or profiles")
	}
	return validationf("unknown auth command %q", args[0])
}

func runAuthSetup(args []string, stdout io.Writer, profile string) error {
	if len(args) > 0 {
		return validationf("unknown auth setup argument %q", args[0])
	}
	reader := bufio.NewReader(commandStdin)
	baseURL, err := promptRequired(reader, stdout, "Nextcloud base URL: ")
	if err != nil {
		return err
	}
	normalizedBaseURL := config.NormalizeBaseURL(baseURL)
	if normalizedBaseURL == "" {
		return validationError("Nextcloud base URL is required")
	}
	if err := config.ValidateBaseURL(normalizedBaseURL); err != nil {
		return err
	}
	username, err := promptRequired(reader, stdout, "Nextcloud username: ")
	if err != nil {
		return err
	}
	securityURL := normalizedBaseURL + "/settings/user/security"
	if err := printLine(stdout, "Open this URL to create an app password: %s", securityURL); err != nil {
		return err
	}
	appPassword, err := promptRequired(reader, stdout, "Nextcloud app password: ")
	if err != nil {
		return err
	}
	cfg := config.Config{BaseURL: normalizedBaseURL, Username: username, Password: appPassword}
	if err := cfg.SaveProfile(profile); err != nil {
		return err
	}
	if strings.TrimSpace(profile) != "" && strings.TrimSpace(profile) != "default" {
		return printLine(stdout, "Saved local auth profile %q.", strings.TrimSpace(profile))
	}
	return printLine(stdout, "Saved local auth config.")
}

func runAuthProfiles(args []string, stdout io.Writer, output outputFormat) error {
	if len(args) > 0 {
		return validationf("unknown auth profiles argument %q", args[0])
	}
	profiles, err := config.ListProfiles()
	if err != nil {
		return err
	}
	if output == outputJSON {
		return printJSON(stdout, profiles)
	}
	for _, profile := range profiles {
		if err := printLine(stdout, "%s\t%s\t%s", profile.Name, profile.BaseURL, profile.Username); err != nil {
			return err
		}
	}
	return nil
}

func promptRequired(reader *bufio.Reader, stdout io.Writer, prompt string) (string, error) {
	if _, err := io.WriteString(stdout, prompt); err != nil {
		return "", err
	}
	value, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("read prompt: %w", err)
	}
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", validationf("%s is required", strings.TrimSuffix(prompt, ": "))
	}
	return trimmed, nil
}
