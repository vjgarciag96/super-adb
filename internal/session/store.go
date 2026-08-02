// Package session manages the last-used device serial across sadb invocations.
//
// # Persistence semantics
//
// sadb stores the most-recently-picked device serial in a plain file at
// ~/.sadb/device. This is intentionally a user-global "last used" value —
// it is not scoped to a single terminal session. Concurrently open terminals
// share the same last-picked device.
//
// If per-terminal scoping is required in the future, the eval-based approach
// (printing "export SADB_DEVICE=<serial>" and wrapping sadb in a shell
// function) is the recommended path, as it would set SADB_DEVICE in the
// shell's environment and take priority over the file (see device.Resolve).
package session

import (
	"os"
	"path/filepath"
	"strings"
)

// Store persists the most-recently-selected device serial.
type Store interface {
	// Load returns the saved serial, or "" if none is saved.
	Load() string
	// Save persists serial for future invocations.
	Save(serial string) error
}

// FileStore persists the serial to a plain-text file at Path.
type FileStore struct {
	Path string
}

// DefaultPath returns the default store path (~/.sadb/device).
func DefaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".sadb", "device")
}

// Load reads and returns the saved serial. Returns "" if the file does not exist.
func (f FileStore) Load() string {
	data, err := os.ReadFile(f.Path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// Save writes serial to the file, creating parent directories as needed.
func (f FileStore) Save(serial string) error {
	if err := os.MkdirAll(filepath.Dir(f.Path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(f.Path, []byte(serial+"\n"), 0o600)
}

// NoopStore is a Store that does nothing. Useful in tests.
type NoopStore struct{}

func (NoopStore) Load() string          { return "" }
func (NoopStore) Save(_ string) error   { return nil }
