package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/vicgarci/sadb/adb"
	"github.com/vicgarci/sadb/internal/capture"
)

var shotCmd = &cobra.Command{
	Use:   "shot [<path>]",
	Short: "Take a screenshot and pull it to the current directory",
	Long: `shot runs 'adb shell screencap' on the active device, pulls the result
to the output directory, and removes the temp file from the device.

An optional positional argument overrides the auto-generated filename:

  sadb shot /tmp/screen.png

If omitted, the file is saved as photo_<timestamp>.png under the current
working directory (or the directory given by --output).`,
	Args: cobra.MaximumNArgs(1),
	RunE: withDevice(func(cmd *cobra.Command, args []string, runner adb.Runner, serial string) error {
		autoName := fmt.Sprintf("photo_%s.png", time.Now().Format("20060102_150405"))
		localPath, err := resolveCapturePath(cmd, args, autoName)
		if err != nil {
			return err
		}

		localPath, err = runShot(runner, serial, localPath)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Saved: %s\n", localPath)
		return nil
	}),
}

// runShot takes a screenshot on the given device and saves it to localPath.
// Device resolution has already happened; serial is the resolved device serial.
func runShot(runner adb.Runner, serial, localPath string) (string, error) {
	return capture.RunPhoto(serial, runner, localPath)
}

func init() {
	shotCmd.Flags().String("output", "", "Directory to save the screenshot (default: current working directory)")
	rootCmd.AddCommand(shotCmd)
}
