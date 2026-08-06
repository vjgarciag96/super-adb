package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/vicgarci/sadb/adb/adbtest"
)

func TestRecordRunE_WrongExtension_CorrectedToMp4(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", nil)              // screenrecord
	f.QueueResponse("", nil)              // pgrep screenrecord → empty, process gone
	f.QueueResponse("1 file pulled", nil) // pull
	f.QueueResponse("", nil)              // rm

	orig := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stderr = w

	err = recordRunE(recordCmd, []string{"demo.avi"}, f, "emulator-5554")

	w.Close()
	os.Stderr = orig
	var buf bytes.Buffer
	io.Copy(&buf, r)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantWarn := "Warning: .avi is not valid for record, saving as demo.mp4"
	if !bytes.Contains(buf.Bytes(), []byte(wantWarn)) {
		t.Errorf("stderr: expected %q, got %q", wantWarn, buf.String())
	}
	// Confirm the corrected path reached the ADB pull call.
	pullCall := f.Calls[2]
	if got := pullCall.Args[len(pullCall.Args)-1]; got != "demo.mp4" {
		t.Errorf("pull destination: expected %q, got %q", "demo.mp4", got)
	}
}

func TestRunRecord_SavesToExplicitPath(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", nil)              // screenrecord
	f.QueueResponse("", nil)              // pgrep screenrecord → empty, process gone
	f.QueueResponse("1 file pulled", nil) // pull
	f.QueueResponse("", nil)              // rm

	explicitPath := filepath.Join(t.TempDir(), "demo.mp4")
	localPath, err := runRecord(f, "emulator-5554", explicitPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if localPath != explicitPath {
		t.Errorf("expected path %q, got %q", explicitPath, localPath)
	}
}

func TestRunRecord_UsesResolvedSerial(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", nil)              // screenrecord
	f.QueueResponse("", nil)              // pgrep screenrecord → empty, process gone
	f.QueueResponse("1 file pulled", nil) // pull
	f.QueueResponse("", nil)              // rm

	_, err := runRecord(f, "env-device-456", filepath.Join(t.TempDir(), "video.mp4"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Calls) != 4 {
		t.Fatalf("expected 4 ADB calls, got %d", len(f.Calls))
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

	_, err := runRecord(f, "emulator-5554", filepath.Join(t.TempDir(), "video.mp4"))
	if err == nil {
		t.Fatal("expected error on screenrecord failure, got nil")
	}
}
