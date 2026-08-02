package capture_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/vicgarci/sadb/adb/adbtest"
	"github.com/vicgarci/sadb/internal/capture"
)

// ── Photo tests ───────────────────────────────────────────────────────────────

func TestRunPhoto_Success(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", nil)           // screencap
	f.QueueResponse("1 file pulled", nil) // pull
	f.QueueResponse("", nil)           // rm

	destDir := t.TempDir()
	localPath, err := capture.RunPhoto("emulator-5554", f, destDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(f.Calls) != 3 {
		t.Fatalf("expected 3 ADB calls, got %d: %+v", len(f.Calls), f.Calls)
	}

	// Call 0: adb shell screencap -p /sdcard/sadb_photo_<ts>.png
	screencap := f.Calls[0]
	if screencap.Serial != "emulator-5554" {
		t.Errorf("screencap: expected serial %q, got %q", "emulator-5554", screencap.Serial)
	}
	if len(screencap.Args) < 4 || screencap.Args[0] != "shell" || screencap.Args[1] != "screencap" || screencap.Args[2] != "-p" {
		t.Errorf("screencap: unexpected args %v", screencap.Args)
	}
	devicePath := screencap.Args[3]
	if !strings.HasPrefix(devicePath, "/sdcard/") || !strings.HasSuffix(devicePath, ".png") {
		t.Errorf("screencap: unexpected device path %q", devicePath)
	}

	// Call 1: adb pull <device_path> <local_path>
	pull := f.Calls[1]
	if pull.Serial != "emulator-5554" {
		t.Errorf("pull: expected serial %q, got %q", "emulator-5554", pull.Serial)
	}
	if len(pull.Args) < 3 || pull.Args[0] != "pull" {
		t.Errorf("pull: unexpected args %v", pull.Args)
	}
	if pull.Args[1] != devicePath {
		t.Errorf("pull: expected device path %q, got %q", devicePath, pull.Args[1])
	}
	if !strings.HasPrefix(pull.Args[2], destDir) || !strings.HasSuffix(pull.Args[2], ".png") {
		t.Errorf("pull: unexpected local path %q (expected under %s, .png)", pull.Args[2], destDir)
	}
	if localPath != pull.Args[2] {
		t.Errorf("RunPhoto returned %q but pull used %q", localPath, pull.Args[2])
	}

	// Call 2: adb shell rm <device_path>
	rm := f.Calls[2]
	if rm.Serial != "emulator-5554" {
		t.Errorf("rm: expected serial %q, got %q", "emulator-5554", rm.Serial)
	}
	if len(rm.Args) < 3 || rm.Args[0] != "shell" || rm.Args[1] != "rm" {
		t.Errorf("rm: unexpected args %v", rm.Args)
	}
	if rm.Args[2] != devicePath {
		t.Errorf("rm: expected to remove %q, got %q", devicePath, rm.Args[2])
	}
}

func TestRunPhoto_ScreencapError_ReturnsError(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", errors.New("permission denied"))

	_, err := capture.RunPhoto("emulator-5554", f, t.TempDir())
	if err == nil {
		t.Fatal("expected error when screencap fails, got nil")
	}
	if len(f.Calls) != 1 {
		t.Errorf("expected exactly 1 ADB call on screencap failure, got %d", len(f.Calls))
	}
}

func TestRunPhoto_PullError_CleanupStillRuns(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", nil)                            // screencap succeeds
	f.QueueResponse("", errors.New("connection reset")) // pull fails
	f.QueueResponse("", nil)                            // rm cleanup

	_, err := capture.RunPhoto("emulator-5554", f, t.TempDir())
	if err == nil {
		t.Fatal("expected error when pull fails, got nil")
	}
	if len(f.Calls) != 3 {
		t.Fatalf("expected 3 calls (screencap, pull, rm), got %d", len(f.Calls))
	}
	rm := f.Calls[2]
	if len(rm.Args) < 2 || rm.Args[0] != "shell" || rm.Args[1] != "rm" {
		t.Errorf("expected rm cleanup after pull failure, got %v", rm.Args)
	}
}

// ── Video tests ───────────────────────────────────────────────────────────────

func TestRunVideo_Success(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", nil)           // screenrecord (completes naturally or via signal)
	f.QueueResponse("1 file pulled", nil) // pull
	f.QueueResponse("", nil)           // rm

	destDir := t.TempDir()
	localPath, err := capture.RunVideo("emulator-5554", f, destDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(f.Calls) != 3 {
		t.Fatalf("expected 3 ADB calls, got %d: %+v", len(f.Calls), f.Calls)
	}

	// Call 0: adb shell screenrecord /sdcard/sadb_video_<ts>.mp4
	rec := f.Calls[0]
	if rec.Serial != "emulator-5554" {
		t.Errorf("screenrecord: expected serial %q, got %q", "emulator-5554", rec.Serial)
	}
	if len(rec.Args) < 3 || rec.Args[0] != "shell" || rec.Args[1] != "screenrecord" {
		t.Errorf("screenrecord: unexpected args %v", rec.Args)
	}
	devicePath := rec.Args[2]
	if !strings.HasPrefix(devicePath, "/sdcard/") || !strings.HasSuffix(devicePath, ".mp4") {
		t.Errorf("screenrecord: unexpected device path %q", devicePath)
	}

	// Call 1: adb pull <device_path> <local_path>
	pull := f.Calls[1]
	if pull.Args[0] != "pull" || pull.Args[1] != devicePath {
		t.Errorf("pull: unexpected args %v", pull.Args)
	}
	if !strings.HasPrefix(pull.Args[2], destDir) || !strings.HasSuffix(pull.Args[2], ".mp4") {
		t.Errorf("pull: unexpected local path %q", pull.Args[2])
	}
	if localPath != pull.Args[2] {
		t.Errorf("RunVideo returned %q but pull used %q", localPath, pull.Args[2])
	}

	// Call 2: adb shell rm <device_path>
	rm := f.Calls[2]
	if len(rm.Args) < 3 || rm.Args[0] != "shell" || rm.Args[1] != "rm" || rm.Args[2] != devicePath {
		t.Errorf("rm: unexpected args %v", rm.Args)
	}
}

func TestRunVideo_PullError_CleanupStillRuns(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", nil)                            // screenrecord
	f.QueueResponse("", errors.New("connection reset")) // pull fails
	f.QueueResponse("", nil)                            // rm cleanup

	_, err := capture.RunVideo("emulator-5554", f, t.TempDir())
	if err == nil {
		t.Fatal("expected error when pull fails, got nil")
	}
	if len(f.Calls) != 3 {
		t.Fatalf("expected 3 calls (screenrecord, pull, rm), got %d", len(f.Calls))
	}
	rm := f.Calls[2]
	if len(rm.Args) < 2 || rm.Args[0] != "shell" || rm.Args[1] != "rm" {
		t.Errorf("expected rm cleanup after pull failure, got %v", rm.Args)
	}
}

func TestRunVideo_ScreenrecordError_ReturnsError(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", errors.New("device offline")) // screenrecord fails (not a signal)

	_, err := capture.RunVideo("emulator-5554", f, t.TempDir())
	if err == nil {
		t.Fatal("expected error when screenrecord fails hard, got nil")
	}
	// No pull or cleanup should run when recording never started.
	if len(f.Calls) != 1 {
		t.Errorf("expected only 1 ADB call on screenrecord failure, got %d", len(f.Calls))
	}
}
