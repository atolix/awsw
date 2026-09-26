package main

import (
	"strings"
	"testing"
)

func TestSelectProfile(t *testing.T) {
	profile, err := selectProfile(strings.NewReader("2\n"), &strings.Builder{}, []string{"dev", "staging"})
	if err != nil {
		t.Fatalf("selectProfile returned an error: %v", err)
	}
	if profile != "staging" {
		t.Fatalf("selected profile = %q, want %q", profile, "staging")
	}
}

func TestSelectProfileCancellation(t *testing.T) {
	_, err := selectProfile(strings.NewReader("\n"), &strings.Builder{}, []string{"dev"})
	if err != errCancelled {
		t.Fatalf("error = %v, want cancellation", err)
	}
}

func TestSelectProfileRejectsOutOfRange(t *testing.T) {
	_, err := selectProfile(strings.NewReader("3\n"), &strings.Builder{}, []string{"dev", "staging"})
	if err == nil {
		t.Fatal("expected invalid selection error")
	}
}
