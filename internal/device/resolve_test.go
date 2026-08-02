package device_test

import (
	"errors"
	"testing"

	"github.com/vicgarci/sadb/adb/adbtest"
	"github.com/vicgarci/sadb/internal/device"
)

// devicesOutput builds the output of "adb devices" with the given serials.
func devicesOutput(serials ...string) string {
	out := "List of devices attached\n"
	for _, s := range serials {
		out += s + "\tdevice\n"
	}
	return out
}

func TestResolve_EnvVarSet(t *testing.T) {
	f := &adbtest.FakeRunner{}
	// No responses queued — "adb devices" should never be called.

	serial, err := device.Resolve("emulator-5554", "", f)
	if len(f.Calls) != 0 {
		t.Errorf("expected no ADB calls when SADB_DEVICE is set, got %d", len(f.Calls))
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if serial != "emulator-5554" {
		t.Errorf("expected serial %q, got %q", "emulator-5554", serial)
	}
}

func TestResolve_FlagOverride(t *testing.T) {
	f := &adbtest.FakeRunner{}

	serial, err := device.Resolve("", "explicit-serial", f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if serial != "explicit-serial" {
		t.Errorf("expected serial %q, got %q", "explicit-serial", serial)
	}
	if len(f.Calls) != 0 {
		t.Errorf("expected no ADB calls with -s override, got %d", len(f.Calls))
	}
}

func TestResolve_NoDevices(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse(devicesOutput(), nil) // empty device list

	_, err := device.Resolve("", "", f)
	if err == nil {
		t.Fatal("expected error for no connected devices, got nil")
	}
	if !errors.Is(err, device.ErrNoDevices) {
		t.Errorf("expected ErrNoDevices, got %v", err)
	}
	if len(f.Calls) != 1 || f.Calls[0].Args[0] != "devices" {
		t.Errorf("expected one 'devices' call, got %+v", f.Calls)
	}
}

func TestResolve_OneDevice_AutoSelect(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse(devicesOutput("emulator-5554"), nil)

	serial, err := device.Resolve("", "", f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if serial != "emulator-5554" {
		t.Errorf("expected auto-selected serial %q, got %q", "emulator-5554", serial)
	}
}

func TestResolve_MultipleDevices_Error(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse(devicesOutput("emulator-5554", "192.168.1.1:5555"), nil)

	_, err := device.Resolve("", "", f)
	if err == nil {
		t.Fatal("expected error for multiple devices, got nil")
	}
	if !errors.Is(err, device.ErrMultipleDevices) {
		t.Errorf("expected ErrMultipleDevices, got %v", err)
	}
}

func TestResolve_ADBError(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", errors.New("adb not found"))

	_, err := device.Resolve("", "", f)
	if err == nil {
		t.Fatal("expected error when adb fails, got nil")
	}
}
