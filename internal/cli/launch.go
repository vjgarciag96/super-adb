package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vicgarci/sadb/adb"
	"github.com/vicgarci/sadb/internal/search"
)

var launchCmd = &cobra.Command{
	Use:   "launch [package]",
	Short: "Launch an app on the active device",
	Long: `launch starts an installed app on the active device.

If a package name is provided it is passed directly to the monkey launcher.
If no package name is provided, the Package Search TUI opens so you can
find and select the app without needing to know its package name.`,
	Args: cobra.MaximumNArgs(1),
	RunE: withDevice(func(cmd *cobra.Command, args []string, runner adb.Runner, serial string) error {
		var pkg string
		if len(args) > 0 {
			pkg = args[0]
		}
		sel := search.BubbleTeaSelector{Stderr: os.Stderr}
		return runLaunch(runner, serial, pkg, sel)
	}),
}

func init() {
	rootCmd.AddCommand(launchCmd)
}

// runLaunch launches pkg on the given device using the monkey launcher.
//
// Device resolution has already happened; serial is the resolved device serial.
// If pkg is empty, FetchPackages is called and the Package Search TUI (sel) is
// opened so the user can select a package. If the user cancels, returns nil.
func runLaunch(runner adb.Runner, serial, pkg string, sel search.Selector) error {
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

	if _, err := runner.Run(serial, "shell", "monkey", "-p", pkg, "-c", "android.intent.category.LAUNCHER", "1"); err != nil {
		return fmt.Errorf("launch %s: %w", pkg, err)
	}
	return nil
}
