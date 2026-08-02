package adb

import (
	"os/exec"
	"strings"
)

// Runner executes ADB commands against a specific device.
// It is the single seam across the system: every ADB invocation goes through here.
type Runner interface {
	// Run executes the given ADB arguments against the specified device serial.
	// serial may be empty when the command does not require a specific device (e.g. "devices").
	// Returns combined stdout+stderr output and any execution error.
	Run(serial string, args ...string) (string, error)
}

// ShellRunner is the real Runner implementation. It delegates to the adb binary on PATH.
type ShellRunner struct{}

// Run shells out to adb, injecting -s <serial> when a serial is provided.
func (r ShellRunner) Run(serial string, args ...string) (string, error) {
	var cmdArgs []string
	if serial != "" {
		cmdArgs = append(cmdArgs, "-s", serial)
	}
	cmdArgs = append(cmdArgs, args...)

	out, err := exec.Command("adb", cmdArgs...).CombinedOutput()
	return strings.TrimRight(string(out), "\n"), err
}
