// Package capture implements the sadb capture subcommands: photo and video.
//
// Both operations follow the same three-step pattern:
//  1. Capture to a temp path on the device.
//  2. Pull the file to the local destination directory.
//  3. Remove the temp file from the device (best-effort).
package capture

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/vicgarci/sadb/adb"
)

// RunPhoto takes a screenshot of the active device and saves it to destDir.
// Returns the full local path of the saved file.
//
// ADB sequence:
//
//	adb shell screencap -p /sdcard/sadb_photo_<ts>.png
//	adb pull /sdcard/sadb_photo_<ts>.png <destDir>/photo_<ts>.png
//	adb shell rm /sdcard/sadb_photo_<ts>.png
func RunPhoto(serial string, runner adb.Runner, destDir string) (string, error) {
	ts := time.Now().Format("20060102_150405")
	devicePath := fmt.Sprintf("/sdcard/sadb_photo_%s.png", ts)
	localPath := filepath.Join(destDir, fmt.Sprintf("photo_%s.png", ts))

	if _, err := runner.Run(serial, "shell", "screencap", "-p", devicePath); err != nil {
		return "", fmt.Errorf("screencap: %w", err)
	}

	return pullAndClean(serial, runner, devicePath, localPath)
}

// RunVideo records the device screen and saves the recording to destDir.
// Recording stops when the underlying adb process exits (e.g. when the user
// presses Ctrl+C — the terminal delivers SIGINT to the child adb process via
// the foreground process group, and the Go process intercepts the signal via
// the caller so it can proceed with pull and cleanup).
//
// A non-zero exit from screenrecord that was caused by a signal (SIGINT/SIGTERM)
// is treated as a normal stop, not an error.
//
// ADB sequence:
//
//	adb shell screenrecord /sdcard/sadb_video_<ts>.mp4
//	adb pull /sdcard/sadb_video_<ts>.mp4 <destDir>/video_<ts>.mp4
//	adb shell rm /sdcard/sadb_video_<ts>.mp4
func RunVideo(serial string, runner adb.Runner, destDir string) (string, error) {
	ts := time.Now().Format("20060102_150405")
	devicePath := fmt.Sprintf("/sdcard/sadb_video_%s.mp4", ts)
	localPath := filepath.Join(destDir, fmt.Sprintf("video_%s.mp4", ts))

	_, err := runner.Run(serial, "shell", "screenrecord", devicePath)
	if err != nil && !isSignalExit(err) {
		return "", fmt.Errorf("screenrecord: %w", err)
	}

	return pullAndClean(serial, runner, devicePath, localPath)
}

// pullAndClean pulls devicePath to localPath then removes devicePath from the
// device (best-effort). Returns localPath on success.
func pullAndClean(serial string, runner adb.Runner, devicePath, localPath string) (string, error) {
	if _, err := runner.Run(serial, "pull", devicePath, localPath); err != nil {
		_, _ = runner.Run(serial, "shell", "rm", devicePath)
		return "", fmt.Errorf("pull: %w", err)
	}
	_, _ = runner.Run(serial, "shell", "rm", devicePath)
	return localPath, nil
}

// isSignalExit reports whether err represents a process that was killed by a
// signal (e.g. SIGINT from Ctrl+C). Such exits are expected for screenrecord
// and should not be treated as failures.
func isSignalExit(err error) bool {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if ws, ok := exitErr.ProcessState.Sys().(syscall.WaitStatus); ok {
			return ws.Signaled()
		}
	}
	return false
}
