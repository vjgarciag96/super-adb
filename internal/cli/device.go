package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/vicgarci/sadb/adb"
	"github.com/vicgarci/sadb/internal/device"
	"github.com/vicgarci/sadb/internal/session"
)

var deviceCmd = &cobra.Command{
	Use:   "device",
	Short: "Switch the active Android device",
	Long: `device opens the interactive device picker and sets the active device
for the current session. The selection is persisted so subsequent sadb
commands target the chosen device without requiring -s each time.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		runner := adb.ShellRunner{}
		cfg := defaultResolveConfig()
		p := cfg.Picker
		store := cfg.Store

		serial, err := runDeviceCommand(runner, p, store)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Active device: %s\n", serial)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deviceCmd)
}

// runDeviceCommand queries connected devices, opens the picker unconditionally,
// persists the selection, and returns the chosen serial.
//
// Unlike device.Resolve, this function always invokes the picker — even when
// only a single device is connected — so the user can consciously confirm or
// switch their active device.
func runDeviceCommand(runner adb.Runner, p device.Picker, store session.Store) (string, error) {
	out, err := runner.Run("", "devices")
	if err != nil {
		return "", fmt.Errorf("listing devices: %w", err)
	}
	serials := device.ParseDevices(out)

	if len(serials) == 0 {
		return "", device.ErrNoDevices
	}

	serial, err := p.Pick(serials)
	if err != nil {
		return "", err
	}

	_ = store.Save(serial) // best-effort; don't fail the command on store error
	return serial, nil
}

