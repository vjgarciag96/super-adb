package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/vicgarci/sadb/adb/adbtest"
)

// errFake is a minimal error type shared across cli tests.
type errFake string

func (e errFake) Error() string { return string(e) }

func TestShotRunE_WrongExtension_CorrectedToPng(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", nil)              // screencap
	f.QueueResponse("1 file pulled", nil) // pull
	f.QueueResponse("", nil)              // rm

	orig := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stderr = w

	err = shotRunE(shotCmd, []string{"myscreen.jpg"}, f, "emulator-5554")

	w.Close()
	os.Stderr = orig
	var buf bytes.Buffer
	io.Copy(&buf, r)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantWarn := "Warning: .jpg is not valid for shot, saving as myscreen.png"
	if !bytes.Contains(buf.Bytes(), []byte(wantWarn)) {
		t.Errorf("stderr: expected %q, got %q", wantWarn, buf.String())
	}
	// Confirm the corrected path reached the ADB pull call.
	if got := f.Calls[1].Args[len(f.Calls[1].Args)-1]; got != "myscreen.png" {
		t.Errorf("pull destination: expected %q, got %q", "myscreen.png", got)
	}
}

func TestRunShot_SavesToExplicitPath(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", nil)              // screencap
	f.QueueResponse("1 file pulled", nil) // pull
	f.QueueResponse("", nil)              // rm

	explicitPath := filepath.Join(t.TempDir(), "screen.png")
	localPath, err := runShot(f, "emulator-5554", explicitPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if localPath != explicitPath {
		t.Errorf("expected path %q, got %q", explicitPath, localPath)
	}
}

func TestRunShot_UsesResolvedSerial(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", nil)              // screencap
	f.QueueResponse("1 file pulled", nil) // pull
	f.QueueResponse("", nil)              // rm

	_, err := runShot(f, "env-device-123", filepath.Join(t.TempDir(), "photo.png"))
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

	_, err := runShot(f, "emulator-5554", filepath.Join(t.TempDir(), "photo.png"))
	if err == nil {
		t.Fatal("expected error on screencap failure, got nil")
	}
}
