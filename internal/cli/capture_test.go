package cli

import (
	"strings"
	"testing"

	"github.com/vicgarci/sadb/adb/adbtest"
)

// ── Photo tests ───────────────────────────────────────────────────────────────

func TestRunCapturePhoto_SavesToOutputDir(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", nil)              // screencap
	f.QueueResponse("1 file pulled", nil) // pull
	f.QueueResponse("", nil)              // rm

	destDir := t.TempDir()
	localPath, err := runCapturePhoto(f, "emulator-5554", destDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(localPath, destDir) {
		t.Errorf("expected path under %s, got %q", destDir, localPath)
	}
	if !strings.HasSuffix(localPath, ".png") {
		t.Errorf("expected .png file, got %q", localPath)
	}
}

func TestRunCapturePhoto_UsesResolvedSerial(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", nil)              // screencap
	f.QueueResponse("1 file pulled", nil) // pull
	f.QueueResponse("", nil)              // rm

	_, err := runCapturePhoto(f, "env-device-123", t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All three ADB calls must use the provided serial.
	if len(f.Calls) != 3 {
		t.Fatalf("expected 3 ADB calls, got %d", len(f.Calls))
	}
	for _, call := range f.Calls {
		if call.Serial != "env-device-123" {
			t.Errorf("expected serial %q, got %q", "env-device-123", call.Serial)
		}
	}
}

func TestRunCapturePhoto_ScreencapError_ReturnsError(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", errFake("screencap failed"))

	_, err := runCapturePhoto(f, "emulator-5554", t.TempDir())
	if err == nil {
		t.Fatal("expected error on screencap failure, got nil")
	}
}

// ── Video tests ───────────────────────────────────────────────────────────────

func TestRunCaptureVideo_SavesToOutputDir(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", nil)              // screenrecord
	f.QueueResponse("1 file pulled", nil) // pull
	f.QueueResponse("", nil)              // rm

	destDir := t.TempDir()
	localPath, err := runCaptureVideo(f, "emulator-5554", destDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(localPath, destDir) {
		t.Errorf("expected path under %s, got %q", destDir, localPath)
	}
	if !strings.HasSuffix(localPath, ".mp4") {
		t.Errorf("expected .mp4 file, got %q", localPath)
	}
}

func TestRunCaptureVideo_UsesResolvedSerial(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", nil)              // screenrecord
	f.QueueResponse("1 file pulled", nil) // pull
	f.QueueResponse("", nil)              // rm

	_, err := runCaptureVideo(f, "env-device-456", t.TempDir())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(f.Calls) != 3 {
		t.Fatalf("expected 3 ADB calls, got %d", len(f.Calls))
	}
	for _, call := range f.Calls {
		if call.Serial != "env-device-456" {
			t.Errorf("expected serial %q, got %q", "env-device-456", call.Serial)
		}
	}
}

func TestRunCaptureVideo_ScreenrecordError_ReturnsError(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", errFake("device offline")) // screenrecord fails (not a signal)

	_, err := runCaptureVideo(f, "emulator-5554", t.TempDir())
	if err == nil {
		t.Fatal("expected error on screenrecord failure, got nil")
	}
}

// errFake is a minimal error type for test scenarios.
type errFake string

func (e errFake) Error() string { return string(e) }
