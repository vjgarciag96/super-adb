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

// fakePicker always returns the configured serial, recording whether it was called.
type fakePicker struct {
	returnSerial string
	returnErr    error
	called       bool
	gotDevices   []string
}

func (p *fakePicker) Pick(devices []string) (string, error) {
	p.called = true
	p.gotDevices = devices
	return p.returnSerial, p.returnErr
}

// fakeStore is an in-memory Store for tests.
// saves records every value passed to Save in order.
type fakeStore struct {
	saves  []string
	loaded string
}

func (s *fakeStore) Load() string { return s.loaded }
func (s *fakeStore) Save(serial string) error {
	s.saves = append(s.saves, serial)
	return nil
}

// lastSave returns the most recent value passed to Save, or "" if Save was never called.
func (s *fakeStore) lastSave() string {
	if len(s.saves) == 0 {
		return ""
	}
	return s.saves[len(s.saves)-1]
}

func cfg(p *fakePicker, s *fakeStore) device.ResolveConfig {
	return device.ResolveConfig{Picker: p, Store: s}
}

func TestResolve_EnvVarSet(t *testing.T) {
	f := &adbtest.FakeRunner{}
	p := &fakePicker{}
	store := &fakeStore{}

	serial, err := device.Resolve("emulator-5554", "", f, cfg(p, store))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if serial != "emulator-5554" {
		t.Errorf("expected serial %q, got %q", "emulator-5554", serial)
	}
	if len(f.Calls) != 0 {
		t.Errorf("expected no ADB calls when SADB_DEVICE is set, got %d", len(f.Calls))
	}
	if p.called {
		t.Error("picker should not be called when env var is set")
	}
}

func TestResolve_FlagOverride(t *testing.T) {
	f := &adbtest.FakeRunner{}
	p := &fakePicker{}
	store := &fakeStore{}

	serial, err := device.Resolve("", "explicit-serial", f, cfg(p, store))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if serial != "explicit-serial" {
		t.Errorf("expected serial %q, got %q", "explicit-serial", serial)
	}
	if len(f.Calls) != 0 {
		t.Errorf("expected no ADB calls with -s override, got %d", len(f.Calls))
	}
	if p.called {
		t.Error("picker should not be called when -s flag is set")
	}
}

func TestResolve_StoredSerial_UsedWhenConnected(t *testing.T) {
	f := &adbtest.FakeRunner{}
	// Device list includes the stored serial.
	f.QueueResponse(devicesOutput("emulator-5554", "192.168.1.1:5555"), nil)
	p := &fakePicker{}
	store := &fakeStore{loaded: "emulator-5554"}

	serial, err := device.Resolve("", "", f, cfg(p, store))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if serial != "emulator-5554" {
		t.Errorf("expected stored serial %q, got %q", "emulator-5554", serial)
	}
	if p.called {
		t.Error("picker should not be called when stored serial is connected")
	}
}

func TestResolve_StoredSerial_ClearedWhenStale(t *testing.T) {
	f := &adbtest.FakeRunner{}
	// Single device so no picker needed — isolates just the clear behaviour.
	f.QueueResponse(devicesOutput("new-device-1"), nil)
	p := &fakePicker{}
	store := &fakeStore{loaded: "old-disconnected-device"}

	serial, err := device.Resolve("", "", f, cfg(p, store))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if serial != "new-device-1" {
		t.Errorf("expected auto-selected serial, got %q", serial)
	}
	// Store must have been explicitly cleared (Save("")) when stale serial detected.
	if len(store.saves) == 0 || store.saves[0] != "" {
		t.Errorf("expected first Save call to clear the store with empty string, got saves=%v", store.saves)
	}
	if p.called {
		t.Error("picker should not be called with a single device")
	}
}

func TestResolve_StoredSerial_IgnoredWhenStale(t *testing.T) {
	f := &adbtest.FakeRunner{}
	// Device list does NOT include the stored serial.
	f.QueueResponse(devicesOutput("new-device-1", "new-device-2"), nil)
	p := &fakePicker{returnSerial: "new-device-1"}
	store := &fakeStore{loaded: "old-disconnected-device"}

	serial, err := device.Resolve("", "", f, cfg(p, store))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !p.called {
		t.Error("picker should be called when stored serial is stale")
	}
	if serial != "new-device-1" {
		t.Errorf("expected picker-chosen serial, got %q", serial)
	}
	if store.lastSave() != "new-device-1" {
		t.Errorf("expected new serial saved after stale pick, got %q", store.lastSave())
	}
}

func TestResolve_NoDevices(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse(devicesOutput(), nil)
	p := &fakePicker{}
	store := &fakeStore{}

	_, err := device.Resolve("", "", f, cfg(p, store))
	if err == nil {
		t.Fatal("expected error for no connected devices, got nil")
	}
	if !errors.Is(err, device.ErrNoDevices) {
		t.Errorf("expected ErrNoDevices, got %v", err)
	}
}

func TestResolve_OneDevice_AutoSelect(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse(devicesOutput("emulator-5554"), nil)
	p := &fakePicker{}
	store := &fakeStore{}

	serial, err := device.Resolve("", "", f, cfg(p, store))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if serial != "emulator-5554" {
		t.Errorf("expected auto-selected serial %q, got %q", "emulator-5554", serial)
	}
	if p.called {
		t.Error("picker should not be called with a single device")
	}
}

func TestResolve_MultipleDevices_InvokesPicker(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse(devicesOutput("emulator-5554", "192.168.1.1:5555"), nil)
	p := &fakePicker{returnSerial: "192.168.1.1:5555"}
	store := &fakeStore{}

	serial, err := device.Resolve("", "", f, cfg(p, store))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if serial != "192.168.1.1:5555" {
		t.Errorf("expected picker-selected serial, got %q", serial)
	}
	if !p.called {
		t.Error("picker should have been called for multiple devices")
	}
	if len(p.gotDevices) != 2 {
		t.Errorf("expected 2 devices passed to picker, got %d", len(p.gotDevices))
	}
}

func TestResolve_MultipleDevices_SavesPickedSerial(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse(devicesOutput("a", "b"), nil)
	p := &fakePicker{returnSerial: "a"}
	store := &fakeStore{}

	_, err := device.Resolve("", "", f, cfg(p, store))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.lastSave() != "a" {
		t.Errorf("expected store to save serial %q, got %q", "a", store.lastSave())
	}
}

func TestResolve_MultipleDevices_PickerAborted(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse(devicesOutput("a", "b"), nil)
	abortErr := errors.New("picker aborted")
	p := &fakePicker{returnErr: abortErr}
	store := &fakeStore{}

	_, err := device.Resolve("", "", f, cfg(p, store))
	if err == nil {
		t.Fatal("expected error when picker aborts")
	}
	if !errors.Is(err, abortErr) {
		t.Errorf("expected picker abort error, got %v", err)
	}
}

func TestResolve_ADBError(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", errors.New("adb not found"))
	p := &fakePicker{}
	store := &fakeStore{}

	_, err := device.Resolve("", "", f, cfg(p, store))
	if err == nil {
		t.Fatal("expected error when adb fails, got nil")
	}
}
