package main

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

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
