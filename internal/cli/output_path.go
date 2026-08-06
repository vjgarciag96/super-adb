package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// resolveCapturePath resolves the full local file path for a capture command
// (shot or record). It reads the --output flag and optional positional
// argument from cmd/args, enforces their mutual exclusion, and falls back to
// autoName under cwd or --output dir when no explicit path is given.
// expectedExt is the required file extension (e.g. ".png", ".mp4"); when an
// explicit path is supplied with the wrong extension it is replaced and a
// warning is written to os.Stderr.
func resolveCapturePath(cmd *cobra.Command, args []string, autoName, expectedExt string) (string, error) {
	outputDir, _ := cmd.Flags().GetString("output")
	var explicitPath string
	if len(args) > 0 {
		explicitPath = args[0]
	}
	if err := validateExclusive(outputDir, explicitPath); err != nil {
		return "", err
	}
	if explicitPath != "" {
		cmdName := cmd.Name()
		fixed, warn := normaliseExtension(explicitPath, expectedExt, cmdName)
		if warn != "" {
			fmt.Fprintln(os.Stderr, warn)
		}
		explicitPath = fixed
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

// normaliseExtension checks whether path has the expected extension (case-insensitive).
// If the path has no extension, it appends expectedExt and returns (newPath, "").
// If the path has the wrong extension, it replaces it and returns (newPath, warning).
// If the extension is correct (any case), it returns (path, "").
func normaliseExtension(path, expectedExt, cmdName string) (string, string) {
	ext := filepath.Ext(path)
	if ext == "" {
		return path + expectedExt, ""
	}
	if strings.EqualFold(ext, expectedExt) {
		return path, ""
	}
	newPath := path[:len(path)-len(ext)] + expectedExt
	warning := fmt.Sprintf("Warning: %s is not valid for %s, saving as %s", ext, cmdName, filepath.Base(newPath))
	return newPath, warning
}

// validateExclusive returns an error when both --output and a positional <path>
// are provided, since the two flags are mutually exclusive.
func validateExclusive(outputDir, explicitPath string) error {
	if outputDir != "" && explicitPath != "" {
		return fmt.Errorf("--output and positional <path> are mutually exclusive")
	}
	return nil
}
