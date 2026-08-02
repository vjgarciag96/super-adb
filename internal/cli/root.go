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

// withDevice wraps a curated command's RunE so that device resolution happens
// automatically before the command body runs. The resolved serial and the
// ADB runner are injected into fn; no command needs to call device.Resolve,
// read SADB_DEVICE, or define its own -s flag.
//
// Resolution priority (highest first):
//
//	-s flag  >  SADB_DEVICE env  >  auto-select / picker
// defaultResolveConfig returns the standard device resolution config using the
// real BubbleTeaPicker and file-backed session store.
func defaultResolveConfig() device.ResolveConfig {
	return device.ResolveConfig{
		Picker: picker.BubbleTeaPicker{Stderr: os.Stderr},
		Store:  session.FileStore{Path: session.DefaultPath()},
	}
}

// withDevice wraps a curated command's RunE so that device resolution happens
// automatically before the command body runs. The resolved serial and the
// ADB runner are injected into fn; no command needs to call device.Resolve,
// read SADB_DEVICE, or define its own -s flag.
//
// Resolution priority (highest first):
//
//	-s flag  >  SADB_DEVICE env  >  auto-select / picker
func withDevice(fn func(cmd *cobra.Command, args []string, runner adb.Runner, serial string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		flagSerial, _ := cmd.Flags().GetString("serial")
		runner := adb.ShellRunner{}
		serial, err := device.Resolve(os.Getenv("SADB_DEVICE"), flagSerial, runner, defaultResolveConfig())
		if err != nil {
			return err
		}
		return fn(cmd, args, runner, serial)
	}
}

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

func init() {
	// -s is a persistent flag so every curated subcommand inherits it without
	// having to declare it themselves. The root pass-through command has
	// DisableFlagParsing=true so it never sees this flag; extractSerial handles
	// -s for that path instead.
	rootCmd.PersistentFlags().StringP("serial", "s", "", "Target a specific device by serial (overrides SADB_DEVICE)")
}
