// Package device handles device resolution: determining which Android device
// a sadb command should target before execution.
package device

import (
	"errors"
	"fmt"
	"strings"

	"github.com/vicgarci/sadb/adb"
	"github.com/vicgarci/sadb/internal/session"
)

// ErrNoDevices is returned when no Android devices are connected.
var ErrNoDevices = errors.New("no devices connected")

// Picker interactively selects one serial from the given device list.
// It is called when multiple devices are connected and no Active Device is set.
type Picker interface {
	Pick(devices []string) (string, error)
}

// ResolveConfig bundles the interactive and persistence dependencies of Resolve.
type ResolveConfig struct {
	Picker Picker
	Store  session.Store
}

// Resolve determines the device serial to use for the current command.
//
// Resolution order (highest priority first):
//  1. flagSerial — the value of -s <serial> passed explicitly on the command line.
//  2. envSerial  — the value of the SADB_DEVICE environment variable.
//  3. Auto-select — if exactly one device is connected, use it.
//  4. Stored serial — if the last-picked serial is still connected, use it.
//  5. Picker — if multiple devices are connected (and no stored serial matches), prompt the user;
//     persist the pick via cfg.Store.
//  6. Error — zero devices → ErrNoDevices.
//
// The stored serial (cfg.Store) is validated against the live device list before use,
// so a disconnected device never silently suppresses the picker.
//
// Both flagSerial and envSerial may be empty strings to signal "not set".
func Resolve(envSerial, flagSerial string, runner adb.Runner, cfg ResolveConfig) (string, error) {
	if flagSerial != "" {
		return flagSerial, nil
	}
	if envSerial != "" {
		return envSerial, nil
	}

	// Query connected devices.
	out, err := runner.Run("", "devices")
	if err != nil {
		return "", fmt.Errorf("listing devices: %w", err)
	}
	serials := ParseDevices(out)

	// If a previously-picked serial is still connected, use it without prompting.
	if stored := cfg.Store.Load(); stored != "" {
		for _, s := range serials {
			if s == stored {
				return stored, nil
			}
		}
		// Stored serial is no longer connected — clear it so the file never
		// holds a serial that isn't currently attached.
		_ = cfg.Store.Save("")
		// Fall through to normal resolution.
	}

	switch len(serials) {
	case 0:
		return "", ErrNoDevices
	case 1:
		return serials[0], nil
	default:
		serial, err := cfg.Picker.Pick(serials)
		if err != nil {
			return "", err
		}
		_ = cfg.Store.Save(serial) // best-effort; don't fail the command on store error
		return serial, nil
	}
}

// ParseDevices extracts device serials from the output of `adb devices`.
// It skips the header line and any lines that are not in "serial\tstate" format.
func ParseDevices(output string) []string {
	var serials []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "List of devices") {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) == 2 && parts[1] == "device" {
			serials = append(serials, parts[0])
		}
	}
	return serials
}
