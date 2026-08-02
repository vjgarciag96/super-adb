package cli

import (
	"strings"
	"testing"

	"github.com/vicgarci/sadb/adb/adbtest"
)

func TestRunRecord_SavesToOutputDir(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", nil)              // screenrecord
	f.QueueResponse("1 file pulled", nil) // pull
	f.QueueResponse("", nil)              // rm

	destDir := t.TempDir()
	localPath, err := runRecord(f, "emulator-5554", destDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(localPath, destDir) {
		t.Errorf("expected path under %s, got %q", destDir, localPath)
	}
	if !strings.HasSuffix(localPath, ".mp4") {
		t.Errorf("expected .mp4 extension, got %q", localPath)
	}
}

func TestRunRecord_UsesResolvedSerial(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", nil)              // screenrecord
	f.QueueResponse("1 file pulled", nil) // pull
	f.QueueResponse("", nil)              // rm

	_, err := runRecord(f, "env-device-456", t.TempDir())
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

func TestRunRecord_ScreenrecordError_ReturnsError(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", errFake("device offline"))

	_, err := runRecord(f, "emulator-5554", t.TempDir())
	if err == nil {
		t.Fatal("expected error on screenrecord failure, got nil")
	}
}
