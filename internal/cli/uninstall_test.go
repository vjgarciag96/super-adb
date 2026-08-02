package cli

import (
	"testing"

	"github.com/vicgarci/sadb/adb/adbtest"
	"github.com/vicgarci/sadb/internal/search"
)

// stubSelector is a Selector that returns a preset package or error without opening a TUI.
type stubSelector struct {
	pkg string
	err error
}

func (s *stubSelector) Select(_ []string) (string, error) {
	return s.pkg, s.err
}

func TestRunUninstall_DirectPackage(t *testing.T) {
	f := &adbtest.FakeRunner{}
	// Device resolution: one device, auto-select.
	f.QueueResponse("List of devices attached\nemulator-5554\tdevice\n", nil)
	// adb uninstall com.example.app
	f.QueueResponse("Success", nil)

	sel := &stubSelector{}
	err := runUninstall(f, "", "com.example.app", sel, noopCfg(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expect exactly 2 calls: devices + uninstall (no pm list packages).
	if len(f.Calls) != 2 {
		t.Fatalf("expected 2 ADB calls, got %d: %+v", len(f.Calls), f.Calls)
	}

	uninstall := f.Calls[1]
	if uninstall.Serial != "emulator-5554" {
		t.Errorf("uninstall: expected serial %q, got %q", "emulator-5554", uninstall.Serial)
	}
	if len(uninstall.Args) != 2 || uninstall.Args[0] != "uninstall" || uninstall.Args[1] != "com.example.app" {
		t.Errorf("uninstall: unexpected args %v", uninstall.Args)
	}
}

func TestRunUninstall_DirectPackage_UsesEnvSerial(t *testing.T) {
	f := &adbtest.FakeRunner{}
	// No devices call when envSerial is set.
	f.QueueResponse("Success", nil)

	sel := &stubSelector{}
	err := runUninstall(f, "env-device-123", "com.example.app", sel, noopCfg(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(f.Calls) != 1 {
		t.Fatalf("expected 1 ADB call (no devices listing), got %d", len(f.Calls))
	}
	if f.Calls[0].Serial != "env-device-123" {
		t.Errorf("expected serial %q, got %q", "env-device-123", f.Calls[0].Serial)
	}
}

func TestRunUninstall_SearchPath_PmListThenUninstall(t *testing.T) {
	f := &adbtest.FakeRunner{}
	// Device resolution: one device, auto-select.
	f.QueueResponse("List of devices attached\nemulator-5554\tdevice\n", nil)
	// pm list packages
	f.QueueResponse("package:com.example.foo\npackage:com.example.bar\n", nil)
	// adb uninstall (selected package)
	f.QueueResponse("Success", nil)

	sel := &stubSelector{pkg: "com.example.bar"}
	err := runUninstall(f, "", "", sel, noopCfg(t))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expect 3 calls: devices, pm list packages, uninstall.
	if len(f.Calls) != 3 {
		t.Fatalf("expected 3 ADB calls, got %d: %+v", len(f.Calls), f.Calls)
	}

	// Call 1: pm list packages
	pmList := f.Calls[1]
	if pmList.Serial != "emulator-5554" {
		t.Errorf("pm list: expected serial %q, got %q", "emulator-5554", pmList.Serial)
	}
	wantPmArgs := []string{"shell", "pm", "list", "packages"}
	if len(pmList.Args) != len(wantPmArgs) {
		t.Fatalf("pm list: expected args %v, got %v", wantPmArgs, pmList.Args)
	}
	for i, a := range wantPmArgs {
		if pmList.Args[i] != a {
			t.Errorf("pm list: arg[%d]: expected %q, got %q", i, a, pmList.Args[i])
		}
	}

	// Call 2: uninstall selected package
	uninstall := f.Calls[2]
	if uninstall.Serial != "emulator-5554" {
		t.Errorf("uninstall: expected serial %q, got %q", "emulator-5554", uninstall.Serial)
	}
	if len(uninstall.Args) != 2 || uninstall.Args[0] != "uninstall" || uninstall.Args[1] != "com.example.bar" {
		t.Errorf("uninstall: unexpected args %v", uninstall.Args)
	}
}

func TestRunUninstall_SearchAborted_NoUninstallCall(t *testing.T) {
	f := &adbtest.FakeRunner{}
	// Device resolution: one device, auto-select.
	f.QueueResponse("List of devices attached\nemulator-5554\tdevice\n", nil)
	// pm list packages
	f.QueueResponse("package:com.example.foo\n", nil)
	// No uninstall response queued — if it's called the test will catch it.

	sel := &stubSelector{err: search.ErrAborted}
	err := runUninstall(f, "", "", sel, noopCfg(t))
	if err != nil {
		t.Fatalf("expected nil error on abort, got %v", err)
	}

	// Only devices + pm list packages should have been called; no uninstall.
	if len(f.Calls) != 2 {
		t.Fatalf("expected 2 ADB calls (devices + pm list), got %d: %+v", len(f.Calls), f.Calls)
	}
	for _, call := range f.Calls {
		if len(call.Args) > 0 && call.Args[0] == "uninstall" {
			t.Errorf("uninstall must not be called when search is aborted: %+v", call)
		}
	}
}

func TestRunUninstall_UninstallError_ReturnsError(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("List of devices attached\nemulator-5554\tdevice\n", nil)
	f.QueueResponse("", errFake("INSTALL_FAILED_USER_RESTRICTED"))

	sel := &stubSelector{}
	err := runUninstall(f, "", "com.example.app", sel, noopCfg(t))
	if err == nil {
		t.Fatal("expected error when uninstall fails, got nil")
	}
}
