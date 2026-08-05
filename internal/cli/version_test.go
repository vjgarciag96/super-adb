package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/vicgarci/sadb/adb/adbtest"
)

func TestPrintADBVersion_WritesOutput(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("Android Debug Bridge version 1.0.41\n", nil)

	var buf bytes.Buffer
	if err := printADBVersion(&buf, f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(f.Calls) != 1 || f.Calls[0].Args[0] != "version" {
		t.Fatalf("expected one 'adb version' call, got %+v", f.Calls)
	}
	if !strings.Contains(buf.String(), "Android Debug Bridge") {
		t.Errorf("unexpected output: %q", buf.String())
	}
}

func TestPrintADBVersion_RunnerError(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", errFake("adb not found"))

	err := printADBVersion(&bytes.Buffer{}, f)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRootCmd_VersionWired(t *testing.T) {
	if rootCmd.Version != Version {
		t.Errorf("rootCmd.Version = %q, want %q — --version flag will not show the correct version", rootCmd.Version, Version)
	}
}
