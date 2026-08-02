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
	f.QueueResponse("Success", nil)

	err := runUninstall(f, "emulator-5554", "com.example.app", &stubSelector{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(f.Calls) != 1 {
		t.Fatalf("expected 1 ADB call, got %d: %+v", len(f.Calls), f.Calls)
	}
	call := f.Calls[0]
	if call.Serial != "emulator-5554" {
		t.Errorf("expected serial %q, got %q", "emulator-5554", call.Serial)
	}
	if len(call.Args) != 2 || call.Args[0] != "uninstall" || call.Args[1] != "com.example.app" {
		t.Errorf("unexpected args %v", call.Args)
	}
}

func TestRunUninstall_SearchPath_PmListThenUninstall(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("package:com.example.foo\npackage:com.example.bar\n", nil) // pm list packages
	f.QueueResponse("Success", nil)                                             // uninstall

	err := runUninstall(f, "emulator-5554", "", &stubSelector{pkg: "com.example.bar"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(f.Calls) != 2 {
		t.Fatalf("expected 2 ADB calls, got %d: %+v", len(f.Calls), f.Calls)
	}

	// Call 0: pm list packages
	pmList := f.Calls[0]
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

	// Call 1: uninstall selected package
	uninstall := f.Calls[1]
	if uninstall.Serial != "emulator-5554" {
		t.Errorf("uninstall: expected serial %q, got %q", "emulator-5554", uninstall.Serial)
	}
	if len(uninstall.Args) != 2 || uninstall.Args[0] != "uninstall" || uninstall.Args[1] != "com.example.bar" {
		t.Errorf("uninstall: unexpected args %v", uninstall.Args)
	}
}

func TestRunUninstall_SearchAborted_NoUninstallCall(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("package:com.example.foo\n", nil) // pm list packages

	err := runUninstall(f, "emulator-5554", "", &stubSelector{err: search.ErrAborted})
	if err != nil {
		t.Fatalf("expected nil error on abort, got %v", err)
	}

	if len(f.Calls) != 1 {
		t.Fatalf("expected 1 ADB call (pm list only), got %d: %+v", len(f.Calls), f.Calls)
	}
	if f.Calls[0].Args[0] == "uninstall" {
		t.Errorf("uninstall must not be called when search is aborted")
	}
}

func TestRunUninstall_UninstallError_ReturnsError(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", errFake("INSTALL_FAILED_USER_RESTRICTED"))

	err := runUninstall(f, "emulator-5554", "com.example.app", &stubSelector{})
	if err == nil {
		t.Fatal("expected error when uninstall fails, got nil")
	}
}
