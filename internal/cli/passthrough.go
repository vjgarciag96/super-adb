package cli

import (
	"fmt"

	"github.com/vicgarci/sadb/adb"
	"github.com/vicgarci/sadb/internal/device"
)

// runPassThrough resolves the active device and forwards args to adb verbatim.
//
// envSerial is the value of the SADB_DEVICE environment variable (empty if unset).
// flagSerial is the value of the -s flag (empty if not provided).
// args are the raw adb arguments to forward (e.g. ["shell", "getprop"]).
//
// Returns the combined output from adb and any error.
func runPassThrough(runner adb.Runner, envSerial, flagSerial string, args []string, cfg device.ResolveConfig) (string, error) {
	serial, err := device.Resolve(envSerial, flagSerial, runner, cfg)
	if err != nil {
		return "", fmt.Errorf("device resolution: %w", err)
	}

	out, err := runner.Run(serial, args...)
	if err != nil {
		return out, fmt.Errorf("adb %v: %w", args, err)
	}
	return out, nil
}
