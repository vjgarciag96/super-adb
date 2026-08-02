package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vicgarci/sadb/adb"
	"github.com/vicgarci/sadb/internal/search"
)

// parseDeviceSerials parses the output of `adb devices` and returns a slice of
// connected device serials, excluding offline devices.
func parseDeviceSerials(output string) []string {
	var serials []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "List of devices") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[1] == "offline" {
			continue
		}
		serials = append(serials, fields[0])
	}
	return serials
}

// completeDeviceSerials is a cobra ValidArgsFunction that suggests connected device serials.
func completeDeviceSerials(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	out, err := adb.ShellRunner{}.Run("", "devices")
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return parseDeviceSerials(out), cobra.ShellCompDirectiveNoFileComp
}

// makeCompletePackages returns a cobra ValidArgsFunction that fetches the installed
// package list directly — no Package Search TUI — and returns it for shell completion.
func makeCompletePackages(runner adb.Runner) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		serial, _ := cmd.Flags().GetString("serial")
		if serial == "" {
			serial = os.Getenv("SADB_DEVICE")
		}
		packages, err := search.FetchPackages(serial, runner)
		if err != nil {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return packages, cobra.ShellCompDirectiveNoFileComp
	}
}

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for sadb.

To load completions for the current session:

  bash:       source <(sadb completion bash)
  zsh:        source <(sadb completion zsh)
  fish:       sadb completion fish | source
  powershell: sadb completion powershell | Out-String | Invoke-Expression

To load completions automatically on login, use 'sadb completion install'
or follow the per-shell instructions printed by each subcommand.`,
}

var completionBashCmd = &cobra.Command{
	Use:   "bash",
	Short: "Generate bash completion script",
	Long: `Generate the bash completion script for sadb.

To load completions for every new bash session:

  Linux:  sadb completion bash > /etc/bash_completion.d/sadb
  macOS:  sadb completion bash > $(brew --prefix)/etc/bash_completion.d/sadb`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return rootCmd.GenBashCompletion(cmd.OutOrStdout())
	},
}

var completionZshCmd = &cobra.Command{
	Use:   "zsh",
	Short: "Generate zsh completion script",
	Long: `Generate the zsh completion script for sadb.

To load completions for every new zsh session:

  sadb completion zsh > "${fpath[1]}/_sadb"

You will need to start a new shell for this to take effect, or run:

  autoload -U compinit && compinit`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return rootCmd.GenZshCompletion(cmd.OutOrStdout())
	},
}

var completionFishCmd = &cobra.Command{
	Use:   "fish",
	Short: "Generate fish completion script",
	Long: `Generate the fish completion script for sadb.

To load completions for every new fish session:

  sadb completion fish > ~/.config/fish/completions/sadb.fish`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
	},
}

var completionPowerShellCmd = &cobra.Command{
	Use:   "powershell",
	Short: "Generate PowerShell completion script",
	Long: `Generate the PowerShell completion script for sadb.

To load completions for every new PowerShell session, add this to your profile:

  sadb completion powershell | Out-String | Invoke-Expression`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return rootCmd.GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
	},
}

var completionInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install completion script for the current shell",
	Long:  `Detect the current shell and install the sadb completion script to the appropriate location.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		shell := filepath.Base(os.Getenv("SHELL"))
		if shell == "" || shell == "." {
			return fmt.Errorf("could not detect shell: $SHELL is not set")
		}

		switch shell {
		case "zsh":
			home, _ := os.UserHomeDir()
			dir := filepath.Join(home, ".zsh", "completions")
			dest, err := writeCompletion(dir, "_sadb", func(w io.Writer) error {
				return rootCmd.GenZshCompletion(w)
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Installed zsh completion to %s\n", dest)
			fmt.Fprintf(cmd.OutOrStdout(), "Add this to your ~/.zshrc if not already present:\n\n  fpath=(%s $fpath)\n  autoload -U compinit && compinit\n\n", dir)

		case "bash":
			var dir string
			if runtime.GOOS == "darwin" {
				dir = "/usr/local/etc/bash_completion.d"
			} else {
				home, _ := os.UserHomeDir()
				dir = filepath.Join(home, ".local", "share", "bash-completion", "completions")
			}
			dest, err := writeCompletion(dir, "sadb", func(w io.Writer) error {
				return rootCmd.GenBashCompletion(w)
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Installed bash completion to %s\n", dest)

		case "fish":
			home, _ := os.UserHomeDir()
			dir := filepath.Join(home, ".config", "fish", "completions")
			dest, err := writeCompletion(dir, "sadb.fish", func(w io.Writer) error {
				return rootCmd.GenFishCompletion(w, true)
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Installed fish completion to %s\n", dest)
			fmt.Fprintln(cmd.OutOrStdout(), "Fish will pick it up automatically in new sessions.")

		case "pwsh", "powershell":
			home, _ := os.UserHomeDir()
			dir := filepath.Join(home, ".config", "powershell")
			dest, err := writeCompletion(dir, "sadb.ps1", func(w io.Writer) error {
				return rootCmd.GenPowerShellCompletionWithDesc(w)
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Installed PowerShell completion to %s\n", dest)
			fmt.Fprintf(cmd.OutOrStdout(), "Add this to your PowerShell profile ($PROFILE):\n\n  . %s\n\n", dest)

		default:
			return fmt.Errorf("unsupported shell %q — run 'sadb completion --help' for manual instructions", shell)
		}
		return nil
	},
}

// writeCompletion creates dir if needed, writes the output of gen to filename inside
// dir, and returns the full destination path.
func writeCompletion(dir, filename string, gen func(io.Writer) error) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create completion dir: %w", err)
	}
	dest := filepath.Join(dir, filename)
	f, err := os.Create(dest)
	if err != nil {
		return "", fmt.Errorf("create completion file: %w", err)
	}
	defer f.Close()
	if err := gen(f); err != nil {
		return "", err
	}
	return dest, nil
}

func init() {
	// Disable cobra's auto-generated completion command; we provide our own.
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	completionCmd.AddCommand(completionBashCmd, completionZshCmd, completionFishCmd, completionPowerShellCmd, completionInstallCmd)
	rootCmd.AddCommand(completionCmd)

	// Suggest adb native subcommands when completing root-level (pass-through) args.
	// Cobra's shell filter handles prefix matching, but we filter ourselves since
	// DisableFlagParsing on root means cobra won't do it for us.
	rootCmd.ValidArgsFunction = func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var matches []string
		for _, c := range adbCommands {
			if strings.HasPrefix(c, toComplete) {
				matches = append(matches, c)
			}
		}
		return matches, cobra.ShellCompDirectiveNoFileComp
	}

	// Suggest device serials when completing the -s / --serial flag.
	// Cobra traverses parent commands when looking up flag completion functions,
	// so registering on root is sufficient for all curated subcommands that
	// inherit the persistent flag.
	_ = rootCmd.RegisterFlagCompletionFunc("serial", completeDeviceSerials)

	// Suggest package names when completing the uninstall argument.
	uninstallCmd.ValidArgsFunction = makeCompletePackages(adb.ShellRunner{})
}
