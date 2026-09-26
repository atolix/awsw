package main

import (
	"testing"

	"charm.land/bubbles/v2/list"
)

func TestProfileItemFilterValue(t *testing.T) {
	item := profileItem("staging")
	if item.FilterValue() != "staging" {
		t.Fatalf("filter value = %q, want %q", item.FilterValue(), "staging")
	}
}

func TestProfileSelectorStartsWithFirstItem(t *testing.T) {
	selector := newProfileSelector([]list.Item{profileItem("dev")}, list.NewDefaultDelegate())
	selected, ok := selector.list.SelectedItem().(profileItem)
	if !ok || selected != profileItem("dev") {
		t.Fatalf("selected item = %q, want %q", selected, "dev")
	}
}
