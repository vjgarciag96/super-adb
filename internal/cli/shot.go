package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vicgarci/sadb/adb"
	"github.com/vicgarci/sadb/internal/capture"
)

var shotCmd = &cobra.Command{
	Use:   "shot",
	Short: "Take a screenshot and pull it to the current directory",
	Long: `shot runs 'adb shell screencap' on the active device, pulls the result
to the output directory, and removes the temp file from the device.`,
	RunE: withDevice(func(cmd *cobra.Command, _ []string, runner adb.Runner, serial string) error {
		outputDir, _ := cmd.Flags().GetString("output")
		if outputDir == "" {
			var err error
			outputDir, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
		}
		localPath, err := runShot(runner, serial, outputDir)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Saved: %s\n", localPath)
		return nil
	}),
}

// runShot takes a screenshot on the given device and saves it to outputDir.
// Device resolution has already happened; serial is the resolved device serial.
func runShot(runner adb.Runner, serial, outputDir string) (string, error) {
	return capture.RunPhoto(serial, runner, outputDir)
}

func init() {
	shotCmd.Flags().String("output", "", "Directory to save the screenshot (default: current working directory)")
	rootCmd.AddCommand(shotCmd)
}
