package cli

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveOutputPath_ExplicitPathReturned(t *testing.T) {
	explicit := "/tmp/screen.png"
	got, err := resolveOutputPath("", explicit, "auto.png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != explicit {
		t.Errorf("expected %q, got %q", explicit, got)
	}
}

func TestResolveOutputPath_AutoNameUnderOutputDir(t *testing.T) {
	dir := t.TempDir()
	got, err := resolveOutputPath(dir, "", "auto.png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != filepath.Join(dir, "auto.png") {
		t.Errorf("expected %q, got %q", filepath.Join(dir, "auto.png"), got)
	}
}

func TestResolveOutputPath_AutoNameUnderCwd(t *testing.T) {
	got, err := resolveOutputPath("", "", "auto.png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(got, "auto.png") {
		t.Errorf("expected path ending with auto.png, got %q", got)
	}
}

func TestValidateExclusive_BothSet_ReturnsError(t *testing.T) {
	err := validateExclusive("/some/dir", "/tmp/screen.png")
	if err == nil {
		t.Fatal("expected error when both --output and explicit path are set, got nil")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("expected 'mutually exclusive' in error, got: %v", err)
	}
}

func TestValidateExclusive_OnlyOutputDir_NoError(t *testing.T) {
	if err := validateExclusive("/some/dir", ""); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateExclusive_OnlyExplicit_NoError(t *testing.T) {
	if err := validateExclusive("", "/tmp/screen.png"); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateExclusive_NeitherSet_NoError(t *testing.T) {
	if err := validateExclusive("", ""); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}
