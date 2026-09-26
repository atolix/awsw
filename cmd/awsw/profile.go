package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type profileConfig struct {
	Region      string
	Output      string
	SSOStartURL string
	SSOSession  string
	SSORegion   string
	AccountID   string
	RoleName    string
}

type profileDetails struct {
	Name        string
	Region      string
	Output      string
	AuthType    string
	Current     bool
	SSOStartURL string
	SSORegion   string
	AccountID   string
	RoleName    string
}

func listProfiles() ([]profileDetails, error) {
	cmd := exec.Command("aws", "configure", "list-profiles")
	output, err := cmd.Output()
	if err != nil {
		if _, ok := err.(*exec.Error); ok {
			return nil, errors.New("AWS CLI was not found; install it before using awsw")
		}
		return nil, fmt.Errorf("could not list AWS profiles: %w", err)
	}

	currentProfile := os.Getenv("AWS_PROFILE")
	config := readProfileConfig()
	var profiles []profileDetails
	for _, line := range strings.Split(string(output), "\n") {
		if profile := strings.TrimSpace(line); profile != "" {
			profiles = append(profiles, loadProfileDetails(profile, currentProfile, config))
		}
	}
	return profiles, nil
}

func loadProfileDetails(name, currentProfile string, config map[string]profileConfig) profileDetails {
	values := config[name]
	authType := "credentials"
	if values.SSOStartURL != "" || values.SSOSession != "" {
		authType = "SSO"
	}

	return profileDetails{
		Name:        name,
		Region:      values.Region,
		Output:      values.Output,
		AuthType:    authType,
		Current:     name == currentProfile,
		SSOStartURL: values.SSOStartURL,
		SSORegion:   values.SSORegion,
		AccountID:   values.AccountID,
		RoleName:    values.RoleName,
	}
}

func readProfileConfig() map[string]profileConfig {
	path := os.Getenv("AWS_CONFIG_FILE")
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return map[string]profileConfig{}
		}
		path = filepath.Join(home, ".aws", "config")
	}

	file, err := os.Open(path)
	if err != nil {
		return map[string]profileConfig{}
	}
	defer file.Close()
	return parseProfileConfig(file)
}

func parseProfileConfig(input io.Reader) map[string]profileConfig {
	profiles := make(map[string]profileConfig)
	current := ""
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section := strings.TrimSpace(line[1 : len(line)-1])
			switch {
			case section == "default":
				current = "default"
			case strings.HasPrefix(section, "profile "):
				current = strings.TrimSpace(strings.TrimPrefix(section, "profile "))
			default:
				current = ""
			}
			if current != "" {
				if _, ok := profiles[current]; !ok {
					profiles[current] = profileConfig{}
				}
			}
			continue
		}

		if current == "" {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		config := profiles[current]
		switch key {
		case "region":
			config.Region = value
		case "output":
			config.Output = value
		case "sso_start_url":
			config.SSOStartURL = value
		case "sso_session":
			config.SSOSession = value
		case "sso_region":
			config.SSORegion = value
		case "sso_account_id":
			config.AccountID = value
		case "sso_role_name":
			config.RoleName = value
		}
		profiles[current] = config
	}
	return profiles
}

func ensureAuthenticated(profile string, errOut io.Writer) error {
	if commandSucceeds("aws", "sts", "get-caller-identity", "--profile", profile) {
		return nil
	}

	// Only SSO profiles should trigger sso login. For other profiles, surface
	// the authentication error instead of unexpectedly starting a login flow.
	if !commandSucceeds("aws", "configure", "get", "sso_start_url", "--profile", profile) {
		return fmt.Errorf("AWS authentication failed for profile %q", profile)
	}

	fmt.Fprintf(errOut, "Authentication required for profile %q. Starting AWS SSO login...\n", profile)
	login := exec.Command("aws", "sso", "login", "--profile", profile)
	login.Stdin = os.Stdin
	login.Stdout = errOut
	login.Stderr = errOut
	if err := login.Run(); err != nil {
		return fmt.Errorf("AWS SSO login failed for profile %q: %w", profile, err)
	}

	if !commandSucceeds("aws", "sts", "get-caller-identity", "--profile", profile) {
		return fmt.Errorf("AWS authentication is still unavailable for profile %q after login", profile)
	}
	return nil
}

func commandSucceeds(name string, args ...string) bool {
	cmd := exec.Command(name, args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run() == nil
}
