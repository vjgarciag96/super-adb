# sadb — Spec

## Problem Statement

ADB is powerful but friction-heavy for day-to-day Android development. Three pain points recur constantly:

1. **Multi-device ambiguity**: you type a full command, then ADB refuses to run because multiple devices are connected. You have to re-invoke with `-s <serial>`, which requires knowing the serial upfront.
2. **Media capture is two commands**: taking a screenshot or recording a screen capture requires one command to capture on-device and a separate `adb pull` to get the file onto your machine.
3. **Package names are opaque**: commands like `adb uninstall` require the exact package name, with no way to search or browse what's installed.

## Solution

`sadb` is a Go CLI binary that wraps ADB transparently. For pass-through commands, it behaves identically to `adb` — same flags, same output — but intercepts at the points where ADB falls short:

- When multiple devices are connected and no Active Device is set, a Device Picker TUI appears before the command runs.
- When a Curated Subcommand needs a package name and none is given, a Package Search TUI appears to let the user filter and select from installed packages.
- `sadb capture` takes a photo or video and pulls it to the current directory in a single command.

The Active Device persists for the terminal session via `SADB_DEVICE`, eliminating repeated device selection across invocations.

## User Stories

1. As an Android developer, I want to run `sadb install app.apk` instead of `adb install app.apk`, so that I get automatic device selection without changing my existing command habits.
2. As an Android developer, I want a Device Picker TUI to appear automatically when I run any sadb command with multiple devices connected, so that I don't have to look up device serials and re-run the command.
3. As an Android developer, I want my device selection to persist for the terminal session, so that I only have to pick once per session rather than on every command.
4. As an Android developer, I want to run `sadb device` to interactively switch the Active Device at any time, so that I can change target without starting a new terminal session.
5. As an Android developer, I want `sadb` to fail immediately with a clear error when no device is connected, so that I know what's wrong without digging through ADB output.
6. As an Android developer, I want to run `sadb capture photo` to take a screenshot and have it pulled to my current directory automatically, so that I don't have to run two separate ADB commands.
7. As an Android developer, I want to run `sadb capture video` to record the screen and have the file pulled to my current directory when I stop recording, so that I don't have to pull it manually.
8. As an Android developer, I want `sadb capture` to save files to my current working directory by default, so that captured files land where I'm working.
9. As an Android developer, I want to pass `--output <path>` to `sadb capture` to override the save location, so that I can direct captures to a specific folder when needed.
10. As an Android developer, I want to run `sadb uninstall` without a package name and have a Package Search TUI open, so that I can find and select the app to uninstall without memorising its package name.
11. As an Android developer, I want the Package Search TUI to filter installed packages in real time as I type, so that I can quickly narrow down to the package I want.
12. As an Android developer, I want any sadb Curated Subcommand that requires a package name to support Package Search when no package is provided, so that I never have to memorise or copy-paste a package name.
13. As an Android developer, I want `sadb` pass-through commands to produce identical output to the equivalent `adb` command, so that I can trust the tool without learning new output formats.
14. As an Android developer, I want to pass `-s <serial>` to `sadb` to target a specific device explicitly, so that I can override the Active Device when needed, consistent with how `adb` works.
15. As an Android developer, I want Tab completion to suggest sadb's own subcommands and all native adb subcommands, so that I don't have to memorise the full command surface.
16. As an Android developer, I want Tab completion on `-s` to list my connected device serials, so that I can select a device without running `adb devices` first.
17. As an Android developer, I want Tab completion on `sadb uninstall` to list installed packages on the Active Device, so that I can find the package without memorising its name.
18. As an Android developer, I want shell completions to work in zsh, bash, fish, and PowerShell, so that I'm not forced into a specific shell.
19. As an Android developer, I want completions to activate automatically when I install sadb via Homebrew, so that I don't have to run any setup commands.

## Implementation Decisions

- **Language and tooling**: Go, with Bubble Tea + Lip Gloss for TUI components and Cobra for CLI argument parsing.
- **ADBRunner interface**: All ADB execution goes through a single `ADBRunner` interface that takes a device serial and a list of arguments and returns output and an error. The real implementation shells out to the `adb` binary. Tests inject a fake implementation. This is the single seam across the entire system.
- **Device resolution**: Before any command runs, sadb checks for `SADB_DEVICE` in the environment. If unset and multiple devices are connected, the Device Picker fires. If unset and exactly one device is connected, that device is used automatically. If no devices are connected, sadb exits immediately with a descriptive error.
- **Session persistence**: After a Device Picker selection, sadb prints an `export SADB_DEVICE=<serial>` instruction for the user's shell to evaluate, or uses a shell integration script that automatically captures and exports the value. The exact mechanism (eval-based vs. env-file) is an implementation detail to be decided during build.
- **`sadb device` command**: Lists all connected devices and emulators in an interactive TUI, allows selection, and updates `SADB_DEVICE` for the session using the same mechanism as the Device Picker.
- **`sadb capture photo`**: Runs `adb shell screencap` to a temp path on the device, then `adb pull` to the current directory (or `--output` path). Cleans up the temp file on the device after pulling.
- **`sadb capture video`**: Runs `adb shell screenrecord` in the foreground; the user presses Ctrl+C to stop. On stop, pulls the file to the current directory and cleans up the temp file on the device.
- **Package Search**: Invoked when a Curated Subcommand requires a package and none is given. Runs `adb shell pm list packages` on the Active Device, parses the output into a list, and renders a live-filtered TUI list. On selection, the chosen package name is passed to the underlying command.
- **Pass-through behaviour**: Any `sadb` invocation that does not match a Curated Subcommand is forwarded to `adb` verbatim after device resolution, with the resolved serial injected as `-s <serial>`.
- **Shell Completion**: sadb uses Cobra's built-in completion system to generate shell-specific Completion Scripts for zsh, bash, fish, and PowerShell. The binary exposes `sadb completion <shell>` to emit the script and `sadb completion install` to write it to the correct location automatically. The Homebrew formula installs the Completion Script as part of `brew install sadb` so no user action is required. adb's native subcommands are included in completions via a hardcoded list in a dedicated source file (`adb_commands.go`), mirroring the approach used by zsh's own `_adb` completion function. Device serials (for `-s`) and package names (for `uninstall`) are completed dynamically at Tab-press time by shelling out to `adb devices` and `adb shell pm list packages` respectively.

## Testing Decisions

- **What makes a good test**: tests should assert on external behaviour — what ADB commands were issued, in what order, with what arguments — not on internal implementation details like how device lists are stored or which functions were called. Each test sets up a fake `ADBRunner`, runs a `sadb` operation, and asserts on the recorded calls and the return value.
- **Modules to test**:
  - Device resolution logic (no device → error; one device → auto-select; multiple devices → picker; `SADB_DEVICE` set → skip picker)
  - Capture sequencing (correct `screencap`/`screenrecord` + `pull` + cleanup sequence)
  - Package Search input/output (correct `pm list packages` call; correct package passed to downstream command)
  - Pass-through forwarding (command forwarded verbatim with correct `-s` injection)
- **Prior art**: none yet — greenfield project. Establish the `ADBRunner` fake in a shared test helper so all modules can reuse it.

## Out of Scope

- A persistent TUI dashboard or always-on UI — `sadb` is invoked per-command, not launched as a long-running process.
- Support for non-Android devices or non-ADB transports.
- Wireless ADB pairing or setup flows.
- Log viewing, logcat filtering, or other debugging workflows (can be added later as Curated Subcommands).
- Plugin or extension system.
- Windows support in the first version (macOS and Linux first).

## Further Notes

- The project directory is `super-adb`; the binary and CLI entry point is `sadb`.
- The target user is Android developers, who have ADB already installed and on their `PATH`.
- Bubble Tea's model-update-view architecture maps naturally onto the Device Picker and Package Search components — each is a self-contained Bubble Tea program returned to the caller with a result.
