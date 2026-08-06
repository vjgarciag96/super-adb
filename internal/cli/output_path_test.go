package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
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

func TestResolveCapturePath_WrongExtension_WritesWarningToStderr(t *testing.T) {
	// Redirect os.Stderr to capture the warning.
	orig := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stderr = w

	cmd := &cobra.Command{Use: "shot"}
	cmd.Flags().String("output", "", "")
	gotPath, err := resolveCapturePath(cmd, []string{"myscreen.jpg"}, "auto.png", ".png")

	w.Close()
	os.Stderr = orig

	var buf bytes.Buffer
	if _, err2 := io.Copy(&buf, r); err2 != nil {
		t.Fatalf("read pipe: %v", err2)
	}

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "myscreen.png" {
		t.Errorf("path: expected %q, got %q", "myscreen.png", gotPath)
	}
	wantWarn := "Warning: .jpg is not valid for shot, saving as myscreen.png"
	stderr := buf.String()
	if !strings.Contains(stderr, wantWarn) {
		t.Errorf("stderr: expected to contain %q, got %q", wantWarn, stderr)
	}
}

func TestNormaliseExtension_MultiDotName_WrongExt_ReplacesAndWarns(t *testing.T) {
	gotPath, gotWarn := normaliseExtension("my.screen.jpg", ".png", "shot")
	if gotPath != "my.screen.png" {
		t.Errorf("path: expected %q, got %q", "my.screen.png", gotPath)
	}
	wantWarn := "Warning: .jpg is not valid for shot, saving as my.screen.png"
	if gotWarn != wantWarn {
		t.Errorf("warning: expected %q, got %q", wantWarn, gotWarn)
	}
}

func TestNormaliseExtension_PathWithDirs_WrongExt_ReplacesAndWarns(t *testing.T) {
	gotPath, gotWarn := normaliseExtension("~/Desktop/myscreen.jpg", ".png", "shot")
	if gotPath != "~/Desktop/myscreen.png" {
		t.Errorf("path: expected %q, got %q", "~/Desktop/myscreen.png", gotPath)
	}
	wantWarn := "Warning: .jpg is not valid for shot, saving as myscreen.png"
	if gotWarn != wantWarn {
		t.Errorf("warning: expected %q, got %q", wantWarn, gotWarn)
	}
}

func TestNormaliseExtension_PathWithDirs_NoExt_AppendsExpected(t *testing.T) {
	gotPath, gotWarn := normaliseExtension("~/Desktop/myscreen", ".png", "shot")
	if gotPath != "~/Desktop/myscreen.png" {
		t.Errorf("path: expected %q, got %q", "~/Desktop/myscreen.png", gotPath)
	}
	if gotWarn != "" {
		t.Errorf("warning: expected empty, got %q", gotWarn)
	}
}

func TestNormaliseExtension_WrongExtension_ReplacesAndWarns(t *testing.T) {
	gotPath, gotWarn := normaliseExtension("myscreen.jpg", ".png", "shot")
	if gotPath != "myscreen.png" {
		t.Errorf("path: expected %q, got %q", "myscreen.png", gotPath)
	}
	wantWarn := "Warning: .jpg is not valid for shot, saving as myscreen.png"
	if gotWarn != wantWarn {
		t.Errorf("warning: expected %q, got %q", wantWarn, gotWarn)
	}
}

func TestNormaliseExtension_CorrectExtensionUppercase_PassThrough(t *testing.T) {
	gotPath, gotWarn := normaliseExtension("myscreen.PNG", ".png", "shot")
	if gotPath != "myscreen.PNG" {
		t.Errorf("path: expected %q, got %q", "myscreen.PNG", gotPath)
	}
	if gotWarn != "" {
		t.Errorf("warning: expected empty, got %q", gotWarn)
	}
}

func TestNormaliseExtension_CorrectExtensionLowercase_PassThrough(t *testing.T) {
	gotPath, gotWarn := normaliseExtension("myscreen.png", ".png", "shot")
	if gotPath != "myscreen.png" {
		t.Errorf("path: expected %q, got %q", "myscreen.png", gotPath)
	}
	if gotWarn != "" {
		t.Errorf("warning: expected empty, got %q", gotWarn)
	}
}

func TestNormaliseExtension_NoExtension_AppendsExpected(t *testing.T) {
	gotPath, gotWarn := normaliseExtension("myscreen", ".png", "shot")
	if gotPath != "myscreen.png" {
		t.Errorf("path: expected %q, got %q", "myscreen.png", gotPath)
	}
	if gotWarn != "" {
		t.Errorf("warning: expected empty, got %q", gotWarn)
	}
}
