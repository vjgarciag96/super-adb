package cli

import (
	"errors"
	"testing"

	"github.com/vicgarci/sadb/adb/adbtest"
	"github.com/vicgarci/sadb/internal/device"
	"github.com/vicgarci/sadb/internal/picker"
)

// stubPicker records which devices were offered and returns a preset response.
type stubPicker struct {
	offered []string
	serial  string
	err     error
}

func (p *stubPicker) Pick(devices []string) (string, error) {
	p.offered = devices
	return p.serial, p.err
}

// recordingStore records Save calls.
type recordingStore struct {
	saved string
}

func (s *recordingStore) Load() string            { return "" }
func (s *recordingStore) Save(serial string) error { s.saved = serial; return nil }

const devicesHeader = "List of devices attached\n"

func TestRunDeviceCommand_NoDevices(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse(devicesHeader, nil)

	p := &stubPicker{}
	store := &recordingStore{}

	_, err := runDeviceCommand(f, p, store)
	if !errors.Is(err, device.ErrNoDevices) {
		t.Fatalf("want ErrNoDevices, got %v", err)
	}
	if p.offered != nil {
		t.Error("picker must not be called when no devices are connected")
	}
	if store.saved != "" {
		t.Error("store must not be written when no devices are connected")
	}
}

func TestRunDeviceCommand_SingleDevice_AlwaysShowsPicker(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse(devicesHeader+"emulator-5554\tdevice\n", nil)

	p := &stubPicker{serial: "emulator-5554"}
	store := &recordingStore{}

	got, err := runDeviceCommand(f, p, store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.offered == nil {
		t.Fatal("picker must be called even with a single device")
	}
	if got != "emulator-5554" {
		t.Errorf("got %q, want %q", got, "emulator-5554")
	}
	if store.saved != "emulator-5554" {
		t.Errorf("store.saved = %q, want %q", store.saved, "emulator-5554")
	}
}

func TestRunDeviceCommand_MultipleDevices_PickerOfferedAll(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse(devicesHeader+"phone-1\tdevice\nphone-2\tdevice\n", nil)

	p := &stubPicker{serial: "phone-2"}
	store := &recordingStore{}

	got, err := runDeviceCommand(f, p, store)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(p.offered) != 2 {
		t.Fatalf("picker offered %d devices, want 2", len(p.offered))
	}
	if got != "phone-2" {
		t.Errorf("got %q, want %q", got, "phone-2")
	}
	if store.saved != "phone-2" {
		t.Errorf("store.saved = %q, want %q", store.saved, "phone-2")
	}
}

func TestRunDeviceCommand_PickerAborted_NothingSaved(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse(devicesHeader+"phone-1\tdevice\n", nil)

	p := &stubPicker{err: picker.ErrAborted}
	store := &recordingStore{}

	_, err := runDeviceCommand(f, p, store)
	if !errors.Is(err, picker.ErrAborted) {
		t.Fatalf("want ErrAborted, got %v", err)
	}
	if store.saved != "" {
		t.Errorf("store must not be written on abort, got %q", store.saved)
	}
}
