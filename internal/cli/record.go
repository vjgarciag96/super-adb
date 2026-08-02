package cli

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"
	"github.com/vicgarci/sadb/adb"
	"github.com/vicgarci/sadb/internal/capture"
)

var recordCmd = &cobra.Command{
	Use:   "record [<path>]",
	Short: "Record the screen and pull the video to the current directory",
	Long: `record runs 'adb shell screenrecord' on the active device. Press Ctrl+C to
stop recording. The video is then pulled to the output directory and the temp
file is removed from the device.

An optional positional argument overrides the auto-generated filename:

  sadb record ~/Desktop/demo.mp4

If omitted, the file is saved as video_<timestamp>.mp4 under the current
working directory (or the directory given by --output).`,
	Args: cobra.MaximumNArgs(1),
	RunE: withDevice(func(cmd *cobra.Command, args []string, runner adb.Runner, serial string) error {
		autoName := fmt.Sprintf("video_%s.mp4", time.Now().Format("20060102_150405"))
		localPath, err := resolveCapturePath(cmd, args, autoName)
		if err != nil {
			return err
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
		var savedPath string
		go func() {
			var captureErr error
			savedPath, captureErr = runRecord(runner, serial, localPath)
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

		fmt.Fprintf(cmd.OutOrStdout(), "Saved: %s\n", savedPath)
		return nil
	}),
}

// runRecord records the screen on the given device and saves it to localPath.
// Device resolution has already happened; serial is the resolved device serial.
func runRecord(runner adb.Runner, serial, localPath string) (string, error) {
	return capture.RunVideo(serial, runner, localPath)
}

func init() {
	recordCmd.Flags().String("output", "", "Directory to save the recording (default: current working directory)")
	rootCmd.AddCommand(recordCmd)
}
