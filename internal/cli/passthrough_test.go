package cli

import (
	"testing"

	"github.com/vicgarci/sadb/adb/adbtest"
)

func TestPassThrough_ForwardsCommandWithSerial(t *testing.T) {
	f := &adbtest.FakeRunner{}
	// Device list returns one device so auto-select fires.
	f.QueueResponse("List of devices attached\nemulator-5554\tdevice\n", nil)
	// The actual pass-through command.
	f.QueueResponse("success", nil)

	out, err := runPassThrough(f, "", "", []string{"shell", "getprop", "ro.product.model"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "success" {
		t.Errorf("expected output %q, got %q", "success", out)
	}

	// Expect: first call = "devices", second call = the forwarded command with serial injected.
	if len(f.Calls) != 2 {
		t.Fatalf("expected 2 ADB calls, got %d: %+v", len(f.Calls), f.Calls)
	}
	pt := f.Calls[1]
	if pt.Serial != "emulator-5554" {
		t.Errorf("expected serial %q on pass-through call, got %q", "emulator-5554", pt.Serial)
	}
	if len(pt.Args) != 3 || pt.Args[0] != "shell" || pt.Args[1] != "getprop" || pt.Args[2] != "ro.product.model" {
		t.Errorf("expected forwarded args, got %v", pt.Args)
	}
}

func TestPassThrough_FlagSerialOverridesResolution(t *testing.T) {
	f := &adbtest.FakeRunner{}
	// No devices queued — resolution should be skipped due to -s flag.
	f.QueueResponse("overridden output", nil)

	out, err := runPassThrough(f, "", "device-123", []string{"install", "app.apk"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "overridden output" {
		t.Errorf("unexpected output: %q", out)
	}

	if len(f.Calls) != 1 {
		t.Fatalf("expected 1 ADB call (no devices listing), got %d", len(f.Calls))
	}
	if f.Calls[0].Serial != "device-123" {
		t.Errorf("expected serial %q, got %q", "device-123", f.Calls[0].Serial)
	}
}

func TestPassThrough_EnvDeviceOverridesResolution(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("env output", nil)

	_, err := runPassThrough(f, "env-device-456", "", []string{"logcat"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(f.Calls) != 1 {
		t.Fatalf("expected 1 ADB call, got %d", len(f.Calls))
	}
	if f.Calls[0].Serial != "env-device-456" {
		t.Errorf("expected serial %q, got %q", "env-device-456", f.Calls[0].Serial)
	}
}
