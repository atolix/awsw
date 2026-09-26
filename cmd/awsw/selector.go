package main

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/atolix/awsw/theme"
)

func selectProfile(in io.Reader, out io.Writer, profiles []profileDetails) (string, error) {
	items := make([]list.Item, 0, len(profiles))
	for _, profile := range profiles {
		items = append(items, profileItem{name: profile.Name, current: profile.Current})
	}

	delegate := newStyleDelegate()
	delegate.ShowDescription = false
	selector := newProfileSelector(items, delegate, profiles)
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

type profileItem struct {
	name    string
	current bool
}

func (p profileItem) FilterValue() string {
	return p.name
}

func (p profileItem) Title() string {
	return p.name
}

func (p profileItem) Description() string {
	return ""
}

type profileSelector struct {
	list      list.Model
	profiles  map[string]profileDetails
	selected  string
	cancelled bool
	width     int
	height    int
}

func newProfileSelector(items []list.Item, delegate list.ItemDelegate, details []profileDetails) *profileSelector {
	profiles := list.New(items, delegate, 120, 20)
	profiles.SetShowTitle(false)
	profiles.SetShowStatusBar(false)
	profiles.SetShowPagination(false)
	profiles.SetFilteringEnabled(true)
	profiles.SetShowHelp(true)

	profileMap := make(map[string]profileDetails, len(details))
	for _, detail := range details {
		profileMap[detail.Name] = detail
	}
	return &profileSelector{list: profiles, profiles: profileMap}
}

func (m *profileSelector) Init() tea.Cmd {
	return nil
}

func (m *profileSelector) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = size.Width
		m.height = size.Height
		m.list.SetSize(m.listWidth(), m.listHeight())
	}

	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "enter":
			if selected, ok := m.list.SelectedItem().(profileItem); ok {
				m.selected = selected.name
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
	listWidth := m.listWidth()
	detailWidth := m.width - listWidth - 3
	if detailWidth < 30 {
		detailWidth = 30
	}

	listView := lipgloss.NewStyle().Width(listWidth).Render(m.list.View())
	detailView := profileDetailView(m.selectedProfile(), detailWidth, m.listHeight())
	view := tea.NewView(lipgloss.JoinHorizontal(lipgloss.Top, listView, detailView))
	view.AltScreen = true
	return view
}

func (m *profileSelector) selectedProfile() profileDetails {
	if item, ok := m.list.SelectedItem().(profileItem); ok {
		return m.profiles[item.name]
	}
	return profileDetails{}
}

func (m *profileSelector) listWidth() int {
	if m.width <= 0 {
		return 48
	}
	return m.width / 2
}

func (m *profileSelector) listHeight() int {
	if m.height <= 0 {
		return 20
	}
	return m.height
}

func profileDetailView(profile profileDetails, width, height int) string {
	if profile.Name == "" {
		return lipgloss.NewStyle().Width(width).Height(height).Render("")
	}

	label := lipgloss.NewStyle().Foreground(theme.Muted)
	value := lipgloss.NewStyle().Foreground(theme.Text)

	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(theme.Primary).Render(profile.Name),
		"",
		label.Render("Region") + "  " + value.Render(displayValue(profile.Region, "not set")),
		label.Render("Auth") + "    " + value.Render(profile.AuthType),
		label.Render("Output") + "  " + value.Render(displayValue(profile.Output, "json")),
	}
	if profile.AccountID != "" {
		lines = append(lines, label.Render("Account")+" "+value.Render(profile.AccountID))
	}
	if profile.RoleName != "" {
		lines = append(lines, label.Render("Role")+"    "+value.Render(profile.RoleName))
	}
	if profile.SSOStartURL != "" {
		lines = append(lines, label.Render("SSO")+"     "+value.Render(profile.SSOStartURL))
	}
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Render(strings.Join(lines, "\n"))
}

func displayValue(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
