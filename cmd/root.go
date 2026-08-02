package cmd

import (
	"github.com/spf13/cobra"
)

// rootCmd is the top-level sadb command.
var rootCmd = &cobra.Command{
	Use:   "sadb",
	Short: "A TUI-enhanced wrapper around ADB for Android developers",
	Long: `sadb is a drop-in for adb with curated subcommands that bundle
multi-step ADB workflows into single ergonomic operations.

Pass-through: any sadb invocation that does not match a curated subcommand
is forwarded to adb verbatim, with automatic device selection applied.`,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
