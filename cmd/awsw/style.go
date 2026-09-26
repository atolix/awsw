package main

import (
	"fmt"
	"io"

	"github.com/atolix/awsw/theme"

	"charm.land/bubbles/v2/list"
	"charm.land/lipgloss/v2"
)

type profileDelegate struct {
	list.DefaultDelegate
}

func newStyleDelegate() profileDelegate {
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

	return profileDelegate{DefaultDelegate: delegate}
}

func (d profileDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	profile, ok := item.(profileItem)
	if !ok {
		d.DefaultDelegate.Render(w, m, index, item)
		return
	}

	style := d.Styles.NormalTitle
	if index == m.Index() {
		style = d.Styles.SelectedTitle
	}
	if profile.current {
		style = style.Foreground(theme.Current)
	}
	fmt.Fprint(w, style.Render(profile.name))
}
