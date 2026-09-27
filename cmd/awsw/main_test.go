package main

import (
	"errors"
	"io"
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

func TestLoadProfileDetailsAuthType(t *testing.T) {
	config := map[string]profileConfig{
		"legacy-sso":  {SSOStartURL: "https://example.awsapps.com/start"},
		"session-sso": {SSOSession: "example"},
		"static":      {Region: "ap-northeast-1"},
	}

	tests := map[string]string{
		"legacy-sso":  authTypeSSO,
		"session-sso": authTypeSSO,
		"static":      authTypeCredentials,
		"missing":     authTypeCredentials,
	}
	for name, want := range tests {
		if got := loadProfileDetails(name, "", config).AuthType; got != want {
			t.Errorf("%s: auth type = %q, want %q", name, got, want)
		}
	}
}

func TestFindProfile(t *testing.T) {
	profiles := []profileDetails{
		{Name: "dev", AuthType: authTypeCredentials},
		{Name: "prod", AuthType: authTypeSSO},
	}

	if got := findProfile(profiles, "prod"); got.AuthType != authTypeSSO {
		t.Fatalf("findProfile(prod) = %#v, want SSO profile", got)
	}
	if got := findProfile(profiles, "unknown"); got.Name != "unknown" || got.AuthType != "" {
		t.Fatalf("findProfile(unknown) = %#v, want bare profile", got)
	}
}

func TestExecuteArguments(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout bool
		wantStderr string
	}{
		{name: "help", args: []string{"--help"}, wantCode: 0, wantStderr: "usage: awsw"},
		{name: "short help", args: []string{"-h"}, wantCode: 0, wantStderr: "usage: awsw"},
		{name: "clear without shell integration", args: []string{"--clear"}, wantCode: 2, wantStderr: "shell integration"},
		{name: "unknown argument", args: []string{"--bogus"}, wantCode: 2, wantStderr: `unknown argument "--bogus"`},
		{name: "init without shell", args: []string{"init"}, wantCode: 2, wantStderr: "usage: awsw init"},
		{name: "init unknown shell", args: []string{"init", "fish"}, wantCode: 2, wantStderr: "unsupported shell"},
		{name: "init zsh", args: []string{"init", "zsh"}, wantCode: 0, wantStdout: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr strings.Builder
			code := execute(tt.args, strings.NewReader(""), &stdout, &stderr)
			if code != tt.wantCode {
				t.Fatalf("exit code = %d, want %d (stderr: %q)", code, tt.wantCode, stderr.String())
			}
			if (stdout.Len() > 0) != tt.wantStdout {
				t.Fatalf("stdout written = %v, want %v: %q", stdout.Len() > 0, tt.wantStdout, stdout.String())
			}
			if tt.wantStderr != "" && !strings.Contains(stderr.String(), tt.wantStderr) {
				t.Fatalf("stderr = %q, want it to contain %q", stderr.String(), tt.wantStderr)
			}
		})
	}
}

func TestExitCode(t *testing.T) {
	var stderr strings.Builder
	if code := exitCode(nil, &stderr); code != 0 {
		t.Fatalf("exitCode(nil) = %d, want 0", code)
	}
	if code := exitCode(errCancelled, &stderr); code != 0 || stderr.Len() != 0 {
		t.Fatalf("exitCode(cancelled) = %d with stderr %q, want 0 and silence", code, stderr.String())
	}
	if code := exitCode(errors.New("boom"), &stderr); code != 1 || !strings.Contains(stderr.String(), "awsw: boom") {
		t.Fatalf("exitCode(error) = %d with stderr %q, want 1 and message", code, stderr.String())
	}
}

// fakeAWS replaces the aws invocations made by ensureAuthenticated. Each call
// to get-caller-identity consumes the next entry of identity; login returns
// loginErr.
type fakeAWS struct {
	identity []bool
	loginErr error
	calls    []string
}

func (f *fakeAWS) install(t *testing.T) {
	t.Helper()
	origSucceeds, origInteractive := commandSucceeds, runInteractive
	commandSucceeds = func(name string, args ...string) bool {
		f.calls = append(f.calls, strings.Join(append([]string{name}, args...), " "))
		if len(f.identity) == 0 {
			t.Fatalf("unexpected quiet command: %s %v", name, args)
		}
		result := f.identity[0]
		f.identity = f.identity[1:]
		return result
	}
	runInteractive = func(errOut io.Writer, name string, args ...string) error {
		f.calls = append(f.calls, strings.Join(append([]string{name}, args...), " "))
		return f.loginErr
	}
	t.Cleanup(func() {
		commandSucceeds, runInteractive = origSucceeds, origInteractive
	})
}

func TestEnsureAuthenticated(t *testing.T) {
	sso := profileDetails{Name: "prod", AuthType: authTypeSSO}
	static := profileDetails{Name: "dev", AuthType: authTypeCredentials}
	login := "aws sso login --profile prod"

	tests := []struct {
		name      string
		profile   profileDetails
		aws       fakeAWS
		wantErr   string
		wantLogin bool
	}{
		{
			name:    "already authenticated",
			profile: sso,
			aws:     fakeAWS{identity: []bool{true}},
		},
		{
			name:    "credentials profile never logs in",
			profile: static,
			aws:     fakeAWS{identity: []bool{false}},
			wantErr: `AWS authentication failed for profile "dev"`,
		},
		{
			name:      "sso profile logs in and succeeds",
			profile:   sso,
			aws:       fakeAWS{identity: []bool{false, true}},
			wantLogin: true,
		},
		{
			name:      "sso login fails",
			profile:   sso,
			aws:       fakeAWS{identity: []bool{false}, loginErr: errors.New("exit status 1")},
			wantErr:   `AWS SSO login failed for profile "prod"`,
			wantLogin: true,
		},
		{
			name:      "sso login succeeds but identity still fails",
			profile:   sso,
			aws:       fakeAWS{identity: []bool{false, false}},
			wantErr:   `still unavailable for profile "prod"`,
			wantLogin: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aws := tt.aws
			aws.install(t)

			var stderr strings.Builder
			err := ensureAuthenticated(tt.profile, &stderr)
			if tt.wantErr == "" && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("error = %v, want it to contain %q", err, tt.wantErr)
			}

			gotLogin := false
			for _, call := range aws.calls {
				if call == login {
					gotLogin = true
				}
			}
			if gotLogin != tt.wantLogin {
				t.Fatalf("login called = %v, want %v (calls: %v)", gotLogin, tt.wantLogin, aws.calls)
			}
			if tt.wantLogin && !strings.Contains(stderr.String(), "Starting AWS SSO login") {
				t.Fatalf("stderr = %q, want login notice", stderr.String())
			}
		})
	}
}
