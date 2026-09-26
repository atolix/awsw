package main

import (
	"github.com/atolix/awsw/theme"

	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
)

func newStyleDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.PaddingLeft(2)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.PaddingLeft(2)

	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(theme.Primary).
		MarginLeft(0).
		PaddingLeft(2).
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(theme.Border).
		Bold(true)

	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Foreground(theme.Primary).
		MarginLeft(0).
		PaddingLeft(2).
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(theme.Border)

	return delegate
}
