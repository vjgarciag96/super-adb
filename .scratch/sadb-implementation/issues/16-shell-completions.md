# 16 — Shell Completions

**What to build:** Shell completions for sadb, bundled in the binary via Cobra's built-in completion system. Supports zsh, bash, fish, and PowerShell. Completes sadb's curated subcommands, adb's native pass-through subcommands (hardcoded list), device serials for `-s`, and package names for `uninstall`.

**Blocked by:** nothing — standalone feature.

**See also:** ADR 0002 — Shell completions via Cobra with Homebrew distribution.

**Status:** done

- [x] Add `adb_commands.go` with a hardcoded slice of all known adb native subcommands
- [x] Register `ValidArgsFunction` on the root command to suggest adb native subcommands for pass-through
- [x] Register `ValidArgsFunction` on `-s` / `--serial` flag to suggest device serials via `adb devices`
- [x] Register `ValidArgsFunction` on `uninstall` to suggest package names via `adb shell pm list packages`
- [x] Add `sadb completion <shell>` subcommand (Cobra built-in — enable via `rootCmd.CompletionOptions`)
- [x] Add `sadb completion install` subcommand that detects the current shell and writes the Completion Script to the correct location
- [x] Verify completions work end-to-end in zsh and bash
- [x] Update Homebrew formula (when built) to install the zsh Completion Script automatically
