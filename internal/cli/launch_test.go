package cli

import (
	"testing"

	"github.com/vicgarci/sadb/adb/adbtest"
	"github.com/vicgarci/sadb/internal/search"
)

func TestRunLaunch_DirectPackage(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", nil)

	err := runLaunch(f, "emulator-5554", "com.example.app", &stubSelector{})
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
	wantArgs := []string{"shell", "monkey", "-p", "com.example.app", "-c", "android.intent.category.LAUNCHER", "1"}
	if len(call.Args) != len(wantArgs) {
		t.Fatalf("expected args %v, got %v", wantArgs, call.Args)
	}
	for i, a := range wantArgs {
		if call.Args[i] != a {
			t.Errorf("arg[%d]: expected %q, got %q", i, a, call.Args[i])
		}
	}
}

func TestRunLaunch_SearchPath_PmListThenMonkey(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("package:com.example.foo\npackage:com.example.bar\n", nil) // pm list packages
	f.QueueResponse("", nil)                                                    // monkey

	err := runLaunch(f, "emulator-5554", "", &stubSelector{pkg: "com.example.bar"})
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

	// Call 1: monkey
	monkey := f.Calls[1]
	if monkey.Serial != "emulator-5554" {
		t.Errorf("monkey: expected serial %q, got %q", "emulator-5554", monkey.Serial)
	}
	wantMonkeyArgs := []string{"shell", "monkey", "-p", "com.example.bar", "-c", "android.intent.category.LAUNCHER", "1"}
	if len(monkey.Args) != len(wantMonkeyArgs) {
		t.Fatalf("monkey: expected args %v, got %v", wantMonkeyArgs, monkey.Args)
	}
	for i, a := range wantMonkeyArgs {
		if monkey.Args[i] != a {
			t.Errorf("monkey: arg[%d]: expected %q, got %q", i, a, monkey.Args[i])
		}
	}
}

func TestRunLaunch_SearchAborted_NoMonkeyCall(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("package:com.example.foo\n", nil) // pm list packages

	err := runLaunch(f, "emulator-5554", "", &stubSelector{err: search.ErrAborted})
	if err != nil {
		t.Fatalf("expected nil error on abort, got %v", err)
	}

	if len(f.Calls) != 1 {
		t.Fatalf("expected 1 ADB call (pm list only), got %d: %+v", len(f.Calls), f.Calls)
	}
	if f.Calls[0].Args[0] == "monkey" {
		t.Errorf("monkey must not be called when search is aborted")
	}
}

func TestRunLaunch_MonkeyError_ReturnsError(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("", errFake("error"))

	err := runLaunch(f, "emulator-5554", "com.example.app", &stubSelector{})
	if err == nil {
		t.Fatal("expected error when monkey fails, got nil")
	}
}
