package cli

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"github.com/vicgarci/sadb/adb"
)

// Version is set at build time via -ldflags "-X github.com/vicgarci/sadb/internal/cli.Version=x.y.z".
// It falls back to "dev" for local builds.
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print sadb and adb versions",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintf(cmd.OutOrStdout(), "super-adb version %s\n", Version)
		return printADBVersion(cmd.OutOrStdout(), adb.ShellRunner{})
	},
}

func printADBVersion(w io.Writer, runner adb.Runner) error {
	out, err := runner.Run("", "version")
	if err != nil {
		return err
	}
	fmt.Fprint(w, out)
	return nil
}

func init() {
	rootCmd.Version = Version
	rootCmd.AddCommand(versionCmd)
}
