package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vicgarci/sadb/adb"
	"github.com/vicgarci/sadb/internal/device"
	"github.com/vicgarci/sadb/internal/picker"
	"github.com/vicgarci/sadb/internal/session"
)

// rootCmd is the top-level sadb command.
var rootCmd = &cobra.Command{
	Use:   "sadb",
	Short: "A TUI-enhanced wrapper around ADB for Android developers",
	Long: `sadb is a drop-in for adb with curated subcommands that bundle
multi-step ADB workflows into single ergonomic operations.

Pass-through: any sadb invocation that does not match a curated subcommand
is forwarded to adb verbatim, with automatic device selection applied.`,

	// DisableFlagParsing lets unknown flags (e.g. adb's own flags) pass through
	// to the underlying adb command without cobra rejecting them.
	DisableFlagParsing: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		runner := adb.ShellRunner{}
		envSerial := os.Getenv("SADB_DEVICE")

		// Extract -s <serial> from args if present, then strip it so the
		// remaining args are forwarded verbatim to adb (adb re-adds -s itself
		// via the runner).
		serial, remaining := extractSerial(args)

		cfg := device.ResolveConfig{
			Picker: picker.BubbleTeaPicker{Stderr: os.Stderr},
			Store:  session.FileStore{Path: session.DefaultPath()},
		}

		out, err := runPassThrough(runner, envSerial, serial, remaining, cfg)
		if out != "" {
			fmt.Fprintln(cmd.OutOrStdout(), out)
		}
		return err
	},
}

// extractSerial scans args for a "-s <serial>" pair and returns the serial
// and the remaining args with the pair removed.
func extractSerial(args []string) (serial string, remaining []string) {
	for i := 0; i < len(args); i++ {
		if args[i] == "-s" && i+1 < len(args) {
			serial = args[i+1]
			remaining = append(remaining, args[:i]...)
			remaining = append(remaining, args[i+2:]...)
			return
		}
	}
	return "", args
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
