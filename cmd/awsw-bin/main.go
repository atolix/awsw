package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

var errCancelled = errors.New("selection cancelled")

func main() {
	if err := run(os.Stdin, os.Stdout, os.Stderr); err != nil {
		if !errors.Is(err, errCancelled) {
			fmt.Fprintf(os.Stderr, "awsw: %v\n", err)
		}
		if errors.Is(err, errCancelled) {
			os.Exit(0)
		}
		os.Exit(1)
	}
}

func run(in io.Reader, out, errOut io.Writer) error {
	profiles, err := listProfiles()
	if err != nil {
		return err
	}
	if len(profiles) == 0 {
		return errors.New("no AWS profiles found")
	}

	profile, err := selectProfile(in, errOut, profiles)
	if err != nil {
		return err
	}

	if err := ensureAuthenticated(profile, errOut); err != nil {
		return err
	}

	// stdout is intentionally reserved for the shell integration.
	_, err = fmt.Fprintln(out, profile)
	return err
}

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

func selectProfile(in io.Reader, out io.Writer, profiles []string) (string, error) {
	items := make([]list.Item, 0, len(profiles))
	for _, profile := range profiles {
		items = append(items, profileItem(profile))
	}

	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	selector := newProfileSelector(items, delegate)
	program := tea.NewProgram(
		selector,
		tea.WithInput(in),
		// stdout is reserved for the selected profile. Bubble Tea renders the
		// interactive UI to stderr so shell command substitution stays safe.
		tea.WithOutput(out),
	)

	finalModel, err := program.Run()
	if err != nil {
		return "", fmt.Errorf("profile selector failed: %w", err)
	}
	finalSelector, ok := finalModel.(*profileSelector)
	if !ok || finalSelector.cancelled || finalSelector.selected == "" {
		return "", errCancelled
	}
	return finalSelector.selected, nil
}

type profileItem string

func (p profileItem) FilterValue() string {
	return string(p)
}

type profileSelector struct {
	list      list.Model
	selected  string
	cancelled bool
}

func newProfileSelector(items []list.Item, delegate list.ItemDelegate) *profileSelector {
	profiles := list.New(items, delegate, 64, len(items)+8)
	profiles.Title = "AWS Profile"
	profiles.SetShowStatusBar(false)
	profiles.SetShowPagination(false)
	profiles.SetShowHelp(true)
	return &profileSelector{list: profiles}
}

func (m *profileSelector) Init() tea.Cmd {
	return nil
}

func (m *profileSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "enter":
			if selected, ok := m.list.SelectedItem().(profileItem); ok {
				m.selected = string(selected)
				return m, tea.Quit
			}
		case "esc", "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *profileSelector) View() tea.View {
	return tea.NewView(m.list.View())
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
