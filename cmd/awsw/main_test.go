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

func TestParseProfileConfig(t *testing.T) {
	config := parseProfileConfig(strings.NewReader(`
[default]
region = ap-northeast-1
output = json

[profile awsw-sso]
sso_start_url = https://example.awsapps.com/start
sso_session = example
sso_account_id = 123456789012
sso_role_name = Developer
region = us-west-2
`))

	if config["default"].Region != "ap-northeast-1" {
		t.Fatalf("default region = %q, want %q", config["default"].Region, "ap-northeast-1")
	}
	if config["awsw-sso"].SSOStartURL == "" || config["awsw-sso"].SSOSession == "" {
		t.Fatalf("SSO profile was not parsed: %#v", config["awsw-sso"])
	}
	if config["awsw-sso"].AccountID != "123456789012" || config["awsw-sso"].RoleName != "Developer" {
		t.Fatalf("SSO identity was not parsed: %#v", config["awsw-sso"])
	}
}

func TestPrintShellIntegration(t *testing.T) {
	var output strings.Builder
	if err := printShellIntegration(&output, "zsh"); err != nil {
		t.Fatalf("printShellIntegration returned an error: %v", err)
	}
	if !strings.Contains(output.String(), `profile="$(command awsw "$@")"`) {
		t.Fatalf("shell integration does not call the binary: %q", output.String())
	}
}

func TestPrintShellIntegrationRejectsUnknownShell(t *testing.T) {
	if err := printShellIntegration(&strings.Builder{}, "fish"); err == nil {
		t.Fatal("expected unsupported shell error")
	}
}
