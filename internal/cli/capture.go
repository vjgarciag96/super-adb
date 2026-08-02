package cli

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/vicgarci/sadb/adb"
	"github.com/vicgarci/sadb/internal/capture"
	"github.com/vicgarci/sadb/internal/device"
	"github.com/vicgarci/sadb/internal/picker"
	"github.com/vicgarci/sadb/internal/session"
)

var captureCmd = &cobra.Command{
	Use:   "capture",
	Short: "Capture screenshots or screen recordings from the active device",
}

var capturePhotoCmd = &cobra.Command{
	Use:   "photo",
	Short: "Take a screenshot and pull it to the current directory",
	Long: `photo runs 'adb shell screencap' on the active device, pulls the result
to the output directory, and removes the temp file from the device.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		outputDir, _ := cmd.Flags().GetString("output")
		if outputDir == "" {
			var err error
			outputDir, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
		}

		runner := adb.ShellRunner{}
		cfg := defaultResolveConfig()
		localPath, err := runCapturePhoto(runner, os.Getenv("SADB_DEVICE"), cfg, outputDir)
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Saved: %s\n", localPath)
		return nil
	},
}

var captureVideoCmd = &cobra.Command{
	Use:   "video",
	Short: "Record the screen and pull the video to the current directory",
	Long: `video runs 'adb shell screenrecord' on the active device. Press Ctrl+C to
stop recording. The video is then pulled to the output directory and the temp
file is removed from the device.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		outputDir, _ := cmd.Flags().GetString("output")
		if outputDir == "" {
			var err error
			outputDir, err = os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
		}

		runner := adb.ShellRunner{}
		cfg := defaultResolveConfig()

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
			localPath, captureErr = runCaptureVideo(runner, os.Getenv("SADB_DEVICE"), cfg, outputDir)
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
	},
}

// runCapturePhoto resolves the active device and takes a screenshot, saving it
// to outputDir. envSerial is the value of SADB_DEVICE (empty if unset).
// Returns the local path of the saved file.
func runCapturePhoto(runner adb.Runner, envSerial string, cfg device.ResolveConfig, outputDir string) (string, error) {
	serial, err := device.Resolve(envSerial, "", runner, cfg)
	if err != nil {
		return "", fmt.Errorf("device resolution: %w", err)
	}
	return capture.RunPhoto(serial, runner, outputDir)
}

// runCaptureVideo resolves the active device and records the screen, saving
// the video to outputDir. envSerial is the value of SADB_DEVICE (empty if
// unset). Returns the local path of the saved file.
func runCaptureVideo(runner adb.Runner, envSerial string, cfg device.ResolveConfig, outputDir string) (string, error) {
	serial, err := device.Resolve(envSerial, "", runner, cfg)
	if err != nil {
		return "", fmt.Errorf("device resolution: %w", err)
	}
	return capture.RunVideo(serial, runner, outputDir)
}

// defaultResolveConfig returns the standard device resolution config using the
// real BubbleTeaPicker and file-backed session store.
func defaultResolveConfig() device.ResolveConfig {
	return device.ResolveConfig{
		Picker: picker.BubbleTeaPicker{Stderr: os.Stderr},
		Store:  session.FileStore{Path: session.DefaultPath()},
	}
}

func init() {
	// --output is a persistent flag on the parent so both subcommands inherit it,
	// matching the spec: "pass --output <path> to sadb capture".
	captureCmd.PersistentFlags().String("output", "", "Directory to save captured files (default: current working directory)")
	captureCmd.AddCommand(capturePhotoCmd, captureVideoCmd)
	rootCmd.AddCommand(captureCmd)
}
