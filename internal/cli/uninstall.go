package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vicgarci/sadb/adb"
	"github.com/vicgarci/sadb/internal/device"
	"github.com/vicgarci/sadb/internal/search"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall [package]",
	Short: "Uninstall an app from the active device",
	Long: `uninstall removes an app from the active device.

If a package name is provided it is passed directly to 'adb uninstall'.
If no package name is provided, the Package Search TUI opens so you can
find and select the app without needing to know its package name.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var pkg string
		if len(args) > 0 {
			pkg = args[0]
		}
		runner := adb.ShellRunner{}
		cfg := defaultResolveConfig()
		sel := search.BubbleTeaSelector{Stderr: os.Stderr}
		return runUninstall(runner, os.Getenv("SADB_DEVICE"), pkg, sel, cfg)
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

// runUninstall resolves the active device and uninstalls pkg from it.
//
// If pkg is empty, FetchPackages is called on the active device and the
// Package Search TUI (sel) is opened so the user can select a package.
// If the user cancels the search, the function returns nil with no action taken.
func runUninstall(runner adb.Runner, envSerial, pkg string, sel search.Selector, cfg device.ResolveConfig) error {
	serial, err := device.Resolve(envSerial, "", runner, cfg)
	if err != nil {
		return fmt.Errorf("device resolution: %w", err)
	}

	if pkg == "" {
		packages, err := search.FetchPackages(serial, runner)
		if err != nil {
			return fmt.Errorf("listing packages: %w", err)
		}
		pkg, err = sel.Select(packages)
		if err != nil {
			if errors.Is(err, search.ErrAborted) {
				return nil
			}
			return fmt.Errorf("package search: %w", err)
		}
	}

	if _, err := runner.Run(serial, "uninstall", pkg); err != nil {
		return fmt.Errorf("uninstall %s: %w", pkg, err)
	}
	return nil
}
