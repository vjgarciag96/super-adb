# sadb

A TUI-enhanced wrapper around ADB for Android developers. Drop-in replacement for `adb` — same flags, same output — with three friction points removed:

- **Multi-device selection** — when multiple devices are connected, an interactive picker appears instead of an error. The choice persists for the session via `SADB_DEVICE`.
- **Single-command capture** — `sadb shot` and `sadb record` take a screenshot or screen recording and pull it to your current directory in one step.
- **Package search** — commands that need a package name (e.g. `sadb uninstall`) open a live-filtered TUI when none is provided, so you never have to memorise or copy-paste package names.

Everything else passes through to `adb` verbatim.

## Install

```sh
brew tap vjgarciag96/super-adb
brew install sadb
```

Shell completions (zsh, bash, fish) are activated automatically.

## Commands

| Command | Description |
|---|---|
| `sadb <any adb command>` | Pass-through to adb with automatic device selection |
| `sadb device` | Switch the active device for the current session |
| `sadb shot [--output path]` | Take a screenshot and pull it locally |
| `sadb record [--output path]` | Record the screen and pull it locally (Ctrl+C to stop) |
| `sadb uninstall [package]` | Uninstall a package, with fuzzy search if no package given |
| `sadb clear-data [package]` | Clear app data, with fuzzy search if no package given |
| `sadb launch [package]` | Launch an app, with fuzzy search if no package given |
| `sadb completion install` | Install shell completions manually |

## Requirements

- macOS (arm64 or amd64)
- `adb` installed and on your `PATH`
