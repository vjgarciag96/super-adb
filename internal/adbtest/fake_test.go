package adbtest_test

import (
	"errors"
	"testing"

	"github.com/vicgarci/sadb/internal/adbtest"
)

// Smoke test: FakeRunner satisfies the adb.Runner contract and records calls correctly.
// This test exercises the shared test helper used by all other test modules.
func TestFakeRunner_RecordsCalls(t *testing.T) {
	f := &adbtest.FakeRunner{}
	f.QueueResponse("emulator-5554", nil)
	f.QueueResponse("", errors.New("device offline"))

	// First call: expect queued output, no error.
	out, err := f.Run("emulator-5554", "devices")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if out != "emulator-5554" {
		t.Errorf("expected output %q, got %q", "emulator-5554", out)
	}

	// Second call: expect queued error.
	_, err = f.Run("emulator-5554", "shell", "getprop")
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	// Third call: queue exhausted — should return zero values.
	out, err = f.Run("", "devices")
	if err != nil {
		t.Fatalf("expected no error on empty queue, got %v", err)
	}
	if out != "" {
		t.Errorf("expected empty output on empty queue, got %q", out)
	}

	// Assert all three calls were recorded.
	if len(f.Calls) != 3 {
		t.Fatalf("expected 3 recorded calls, got %d", len(f.Calls))
	}

	first := f.Calls[0]
	if first.Serial != "emulator-5554" || len(first.Args) != 1 || first.Args[0] != "devices" {
		t.Errorf("unexpected first call: %+v", first)
	}

	second := f.Calls[1]
	if second.Serial != "emulator-5554" || second.Args[0] != "shell" || second.Args[1] != "getprop" {
		t.Errorf("unexpected second call: %+v", second)
	}

	third := f.Calls[2]
	if third.Serial != "" || len(third.Args) != 1 || third.Args[0] != "devices" {
		t.Errorf("unexpected third call: %+v", third)
	}
}
