package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// resolveCapturePath resolves the full local file path for a capture command
// (shot or record). It reads the --output flag and optional positional
// argument from cmd/args, enforces their mutual exclusion, and falls back to
// autoName under cwd or --output dir when no explicit path is given.
func resolveCapturePath(cmd *cobra.Command, args []string, autoName string) (string, error) {
	outputDir, _ := cmd.Flags().GetString("output")
	var explicitPath string
	if len(args) > 0 {
		explicitPath = args[0]
	}
	if err := validateExclusive(outputDir, explicitPath); err != nil {
		return "", err
	}
	return resolveOutputPath(outputDir, explicitPath, autoName)
}

// resolveOutputPath returns the full local file path for a capture command.
//
// If explicitPath is non-empty it is returned directly. Otherwise autoName is
// joined with outputDir (or the current working directory when outputDir is
// empty).
func resolveOutputPath(outputDir, explicitPath, autoName string) (string, error) {
	if explicitPath != "" {
		return explicitPath, nil
	}
	if outputDir == "" {
		var err error
		outputDir, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("getting working directory: %w", err)
		}
	}
	return filepath.Join(outputDir, autoName), nil
}

// validateExclusive returns an error when both --output and a positional <path>
// are provided, since the two flags are mutually exclusive.
func validateExclusive(outputDir, explicitPath string) error {
	if outputDir != "" && explicitPath != "" {
		return fmt.Errorf("--output and positional <path> are mutually exclusive")
	}
	return nil
}
