// Package device handles device resolution: determining which Android device
// a sadb command should target before execution.
package device

import (
	"errors"
	"fmt"
	"strings"

	"github.com/vicgarci/sadb/adb"
)

// Sentinel errors returned by Resolve.
var (
	// ErrNoDevices is returned when no Android devices are connected.
	ErrNoDevices = errors.New("no devices connected")

	// ErrMultipleDevices is returned when multiple devices are connected and
	// no Active Device is set. The caller should prompt the user to run `sadb device`.
	ErrMultipleDevices = errors.New("multiple devices connected")
)

// Resolve determines the device serial to use for the current command.
//
// Resolution order (highest priority first):
//  1. flagSerial — the value of -s <serial> passed explicitly on the command line.
//  2. envSerial  — the value of the SADB_DEVICE environment variable.
//  3. Auto-select — if exactly one device is connected, use it.
//  4. Error       — zero devices → ErrNoDevices; multiple devices → ErrMultipleDevices.
//
// Both flagSerial and envSerial may be empty strings to signal "not set".
func Resolve(envSerial, flagSerial string, runner adb.Runner) (string, error) {
	if flagSerial != "" {
		return flagSerial, nil
	}
	if envSerial != "" {
		return envSerial, nil
	}

	// Neither override is set — query connected devices.
	out, err := runner.Run("", "devices")
	if err != nil {
		return "", fmt.Errorf("listing devices: %w", err)
	}

	serials := parseDevices(out)
	switch len(serials) {
	case 0:
		return "", ErrNoDevices
	case 1:
		return serials[0], nil
	default:
		return "", fmt.Errorf("%w: run `sadb device` to select one", ErrMultipleDevices)
	}
}

// parseDevices extracts device serials from the output of `adb devices`.
// It skips the header line and any lines that are not in "serial\tstate" format.
func parseDevices(output string) []string {
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
