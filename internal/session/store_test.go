package session_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vicgarci/sadb/internal/session"
)

func TestFileStore_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "device")
	s := session.FileStore{Path: path}

	// Load returns "" when file doesn't exist.
	if got := s.Load(); got != "" {
		t.Errorf("expected empty load on missing file, got %q", got)
	}

	// Save then Load round-trips the serial.
	if err := s.Save("emulator-5554"); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if got := s.Load(); got != "emulator-5554" {
		t.Errorf("expected %q after save, got %q", "emulator-5554", got)
	}
}

func TestFileStore_Save_CreatesParentDirs(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "dir", "device")
	s := session.FileStore{Path: path}

	if err := s.Save("serial-123"); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file to exist after Save: %v", err)
	}
}

func TestFileStore_Load_TrimsWhitespace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "device")
	if err := os.WriteFile(path, []byte("  emulator-5554\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	s := session.FileStore{Path: path}
	if got := s.Load(); got != "emulator-5554" {
		t.Errorf("expected trimmed serial, got %q", got)
	}
}

func TestNoopStore(t *testing.T) {
	s := session.NoopStore{}

	if got := s.Load(); got != "" {
		t.Errorf("NoopStore.Load should return empty, got %q", got)
	}
	if err := s.Save("anything"); err != nil {
		t.Errorf("NoopStore.Save should not error, got %v", err)
	}
}
