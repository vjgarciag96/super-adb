package cli

import (
	"testing"

	"github.com/vicgarci/sadb/adb/adbtest"
	"github.com/vicgarci/sadb/internal/search"
)

func TestRunClearData_DirectPackage(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("Success", nil)

	err := runClearData(f, "emulator-5554", "com.example.app", &stubSelector{})
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
	wantArgs := []string{"shell", "pm", "clear", "com.example.app"}
	if len(call.Args) != len(wantArgs) {
		t.Fatalf("expected args %v, got %v", wantArgs, call.Args)
	}
	for i, a := range wantArgs {
		if call.Args[i] != a {
			t.Errorf("arg[%d]: expected %q, got %q", i, a, call.Args[i])
		}
	}
}

func TestRunClearData_SearchPath_PmListThenClear(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("package:com.example.foo\npackage:com.example.bar\n", nil) // pm list packages
	f.QueueResponse("Success", nil)                                             // pm clear

	err := runClearData(f, "emulator-5554", "", &stubSelector{pkg: "com.example.bar"})
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

	// Call 1: pm clear selected package
	clear := f.Calls[1]
	if clear.Serial != "emulator-5554" {
		t.Errorf("pm clear: expected serial %q, got %q", "emulator-5554", clear.Serial)
	}
	wantClearArgs := []string{"shell", "pm", "clear", "com.example.bar"}
	if len(clear.Args) != len(wantClearArgs) {
		t.Fatalf("pm clear: expected args %v, got %v", wantClearArgs, clear.Args)
	}
	for i, a := range wantClearArgs {
		if clear.Args[i] != a {
			t.Errorf("pm clear: arg[%d]: expected %q, got %q", i, a, clear.Args[i])
		}
	}
}

func TestRunClearData_SearchAborted_NoClearCall(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("package:com.example.foo\n", nil) // pm list packages

	err := runClearData(f, "emulator-5554", "", &stubSelector{err: search.ErrAborted})
	if err != nil {
		t.Fatalf("expected nil error on abort, got %v", err)
	}

	if len(f.Calls) != 1 {
		t.Fatalf("expected 1 ADB call (pm list only), got %d: %+v", len(f.Calls), f.Calls)
	}
	if f.Calls[0].Args[0] == "shell" && len(f.Calls[0].Args) > 2 && f.Calls[0].Args[2] == "clear" {
		t.Errorf("pm clear must not be called when search is aborted")
	}
}

func TestRunClearData_ClearError_ReturnsError(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", errFake("failed to clear data"))

	err := runClearData(f, "emulator-5554", "com.example.app", &stubSelector{})
	if err == nil {
		t.Fatal("expected error when pm clear fails, got nil")
	}
}
