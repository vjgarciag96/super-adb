# Shell completions via Cobra with Homebrew distribution

sadb uses Cobra's built-in completion system to generate shell-specific Completion Scripts, distributed automatically through the Homebrew formula.

## Context

adb itself has no built-in completion support. Tab completions for `adb` exist only because zsh ships a community-written `_adb` completion function as part of its standard library — a 624-line script maintained separately from the adb binary. bash has no equivalent.

sadb is a new command that no shell knows about. Without completions, users must memorise both sadb's curated subcommands and adb's native pass-through subcommands.

Three things needed a decision:

1. **How to generate Completion Scripts** — hand-written shell scripts vs. Cobra's built-in generator
2. **How to supply adb's native subcommands** — dynamic parsing of `adb help` vs. a hardcoded list
3. **How to distribute Completion Scripts** — manual user setup vs. automatic via package manager

## Decisions

### 1. Cobra's built-in completion generator

Cobra ships with `cobra.Command` completion support out of the box. Registering `sadb completion <shell>` requires minimal code and produces correct, idiomatic Completion Scripts for zsh, bash, fish, and PowerShell from a single source. Dynamic completions (device serials, package names) are registered per-command via `ValidArgsFunction`.

Hand-writing shell scripts (as zsh's `_adb` does) was rejected: it requires separate maintenance per shell, is harder to keep in sync with command changes, and provides no benefit over Cobra's generator for a Cobra-based CLI.

### 2. Hardcoded list of adb native subcommands

adb's subcommand surface is stable and well-known. Parsing `adb help` at Tab-press time would add a process-spawn delay on every completion invocation — poor UX. The precedent set by zsh's own `_adb` function (which uses a static `ALL_ADB_COMMANDS` array with no version detection) confirms that hardcoding is the accepted approach for this class of tool.

The list lives in a dedicated `adb_commands.go` source file so that future additions are easy to locate and update.

### 3. Homebrew formula handles distribution

`sadb completion <shell>` and `sadb completion install` are available for manual setup, but the primary distribution path is the Homebrew formula. Homebrew installs Completion Scripts to the correct shell-specific location as part of `brew install sadb` — users get completions with no extra steps.

This decision is consistent with ADR 0001, which chose Go specifically because it "integrates naturally with Homebrew." Shell completions are the first concrete use of that integration.

## Consequences

- sadb exposes `sadb completion <shell>` and `sadb completion install` as user-facing commands.
- adb's native subcommands are maintained as a hardcoded list in `adb_commands.go`; the list will occasionally need manual updates as adb evolves.
- Homebrew distribution must be built before zero-setup completions are available to users. Until then, `sadb completion install` serves as a bridge.
- Dynamic completions (device serials, package names) spawn child processes at Tab-press time. This is acceptable latency for intentional completions on typed flags.
