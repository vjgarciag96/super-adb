package cli

import (
	"strings"
	"testing"

	"github.com/vicgarci/sadb/adb/adbtest"
)

// errFake is a minimal error type shared across cli tests.
type errFake string

func (e errFake) Error() string { return string(e) }

func TestRunShot_SavesToOutputDir(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", nil)              // screencap
	f.QueueResponse("1 file pulled", nil) // pull
	f.QueueResponse("", nil)              // rm

	destDir := t.TempDir()
	localPath, err := runShot(f, "emulator-5554", destDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(localPath, destDir) {
		t.Errorf("expected path under %s, got %q", destDir, localPath)
	}
	if !strings.HasSuffix(localPath, ".png") {
		t.Errorf("expected .png extension, got %q", localPath)
	}
}

func TestRunShot_UsesResolvedSerial(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", nil)              // screencap
	f.QueueResponse("1 file pulled", nil) // pull
	f.QueueResponse("", nil)              // rm

	_, err := runShot(f, "env-device-123", t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(f.Calls) != 3 {
		t.Fatalf("expected 3 ADB calls, got %d", len(f.Calls))
	}
	for _, call := range f.Calls {
		if call.Serial != "env-device-123" {
			t.Errorf("expected serial %q, got %q", "env-device-123", call.Serial)
		}
	}
}

func TestRunShot_ScreencapError_ReturnsError(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", errFake("screencap failed"))

	_, err := runShot(f, "emulator-5554", t.TempDir())
	if err == nil {
		t.Fatal("expected error on screencap failure, got nil")
	}
}
