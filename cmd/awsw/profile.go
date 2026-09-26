package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func listProfiles() ([]string, error) {
	cmd := exec.Command("aws", "configure", "list-profiles")
	output, err := cmd.Output()
	if err != nil {
		if _, ok := err.(*exec.Error); ok {
			return nil, errors.New("AWS CLI was not found; install it before using awsw")
		}
		return nil, fmt.Errorf("could not list AWS profiles: %w", err)
	}

	var profiles []string
	for _, line := range strings.Split(string(output), "\n") {
		if profile := strings.TrimSpace(line); profile != "" {
			profiles = append(profiles, profile)
		}
	}
	return profiles, nil
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
