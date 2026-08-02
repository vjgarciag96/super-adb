package cli

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/vicgarci/sadb/adb"
	"github.com/vicgarci/sadb/internal/capture"
)

var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record the screen and pull the video to the current directory",
	Long: `record runs 'adb shell screenrecord' on the active device. Press Ctrl+C to
stop recording. The video is then pulled to the output directory and the temp
file is removed from the device.`,
	RunE: withDevice(func(cmd *cobra.Command, _ []string, runner adb.Runner, serial string) error {
		outputDir, _ := cmd.Flags().GetString("output")
		if outputDir == "" {
			var err error
			outputDir, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
		}

		// Intercept Ctrl+C so the Go process stays alive after the user stops
		// recording. The SIGINT propagates to the child adb process via the
		// terminal's process group, which stops screenrecord. We then proceed
		// to pull and clean up.
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		defer signal.Stop(sigCh)

		fmt.Fprintln(cmd.OutOrStdout(), "Recording… press Ctrl+C to stop.")

		recordDone := make(chan error, 1)
		var localPath string
		go func() {
			var captureErr error
			localPath, captureErr = runRecord(runner, serial, outputDir)
			recordDone <- captureErr
		}()

		select {
		case <-sigCh:
			// Signal received — adb child also got it; wait for the goroutine.
			if err := <-recordDone; err != nil {
				return err
			}
		case err := <-recordDone:
			if err != nil {
				return err
			}
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Saved: %s\n", localPath)
		return nil
	}),
}

// runRecord records the screen on the given device and saves it to outputDir.
// Device resolution has already happened; serial is the resolved device serial.
func runRecord(runner adb.Runner, serial, outputDir string) (string, error) {
	return capture.RunVideo(serial, runner, outputDir)
}

func init() {
	recordCmd.Flags().String("output", "", "Directory to save the recording (default: current working directory)")
	rootCmd.AddCommand(recordCmd)
}
