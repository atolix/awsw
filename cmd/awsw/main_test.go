package main

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
)

func TestProfileItemFilterValue(t *testing.T) {
	item := profileItem{name: "staging"}
	if item.FilterValue() != "staging" {
		t.Fatalf("filter value = %q, want %q", item.FilterValue(), "staging")
	}
}

func TestProfileSelectorStartsWithFirstItem(t *testing.T) {
	selector := newProfileSelector(
		[]list.Item{profileItem{name: "dev"}},
		list.NewDefaultDelegate(),
		[]profileDetails{{Name: "dev", Region: "ap-northeast-1", AuthType: "credentials"}},
	)
	selected, ok := selector.list.SelectedItem().(profileItem)
	if !ok || selected.name != "dev" {
		t.Fatalf("selected item = %q, want %q", selected.name, "dev")
	}
	if !strings.Contains(selector.list.View(), "dev") {
		t.Fatalf("list view does not contain selected profile: %q", selector.list.View())
	}
}
