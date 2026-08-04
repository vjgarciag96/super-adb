package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vicgarci/sadb/adb"
	"github.com/vicgarci/sadb/internal/search"
)

var clearDataCmd = &cobra.Command{
	Use:   "clear-data [package]",
	Short: "Clear app data for an app on the active device",
	Long: `clear-data clears the stored data for an app on the active device.

If a package name is provided it is passed directly to 'adb shell pm clear'.
If no package name is provided, the Package Search TUI opens so you can
find and select the app without needing to know its package name.`,
	Args: cobra.MaximumNArgs(1),
	RunE: withDevice(func(cmd *cobra.Command, args []string, runner adb.Runner, serial string) error {
		var pkg string
		if len(args) > 0 {
			pkg = args[0]
		}
		sel := search.BubbleTeaSelector{Stderr: os.Stderr}
		return runClearData(runner, serial, pkg, sel)
	}),
}

func init() {
	rootCmd.AddCommand(clearDataCmd)
}

func runClearData(runner adb.Runner, serial, pkg string, sel search.Selector) error {
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

	if _, err := runner.Run(serial, "shell", "pm", "clear", pkg); err != nil {
		return fmt.Errorf("clear data %s: %w", pkg, err)
	}
	return nil
}
