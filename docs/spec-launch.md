# Spec: `launch` Curated Subcommand

## Problem

Launching an installed app by package name requires knowing the exact monkey invocation
(`adb shell monkey -p <pkg> -c android.intent.category.LAUNCHER 1`), which is
non-obvious and hard to remember. `sadb uninstall` and `sadb clear-data` already solve
the analogous package-discovery problem with Package Search. `launch` applies the same
pattern to app launching.

## Behaviour Rules

| Scenario | Behaviour |
|---|---|
| Package provided (`sadb launch com.example.app`) | Passes directly to monkey; no TUI |
| No package provided (`sadb launch`) | Opens Package Search TUI; user selects |
| User cancels Package Search (Esc / Ctrl-C) | Exits silently, no launch |
| App already running | Brought to foreground (default monkey behaviour; no force-stop) |
| Success | Silent — no output to stdout or stderr |
| Monkey returns non-zero (e.g. bad package) | Returns error to the user |

## Launch Mechanism

```
adb shell monkey -p <package> -c android.intent.category.LAUNCHER 1
```

Monkey is preferred over `am start` because it requires no activity-name resolution
and is the idiomatic single-call way to fire the launcher intent for a package.

## Files to Touch

| File | Change |
|---|---|
| `internal/cli/launch.go` | New file — command definition and `runLaunch` |
| `internal/cli/launch_test.go` | New file — tests for `runLaunch` |
| `internal/cli/adb_commands.go` | Add `"launch"` between `"kill-server"` and `"logcat"` |
| `internal/cli/completion.go` | Register `ValidArgsFunction` on `launchCmd` in `init()` |
| `README.md` | Add row to command table |

## README Row

```markdown
| `sadb launch [package]` | Launch an app, with fuzzy search if no package given |
```

## Completion

In `completion.go`, append to the end of `init()`, alongside the existing lines for
`uninstallCmd` and `clearDataCmd`:

```go
launchCmd.ValidArgsFunction = makeCompletePackages(adb.ShellRunner{})
```

## Test Seams

All tests live in `internal/cli/launch_test.go`, package `cli`.
The confirmed seam is `runLaunch(runner, serial, pkg, sel)` — the same public boundary
used by `runUninstall` and `runClearData`.

The `stubSelector` type defined in `uninstall_test.go` is already visible within the
`cli` package test files; no redefinition needed.

### Test cases

#### `TestRunLaunch_DirectPackage`

Package provided — skips TUI, issues exactly one ADB call with the correct monkey args.

- Runner queue: one success response (`""`, `nil`)
- Call: `runLaunch(f, "emulator-5554", "com.example.app", &stubSelector{})`
- Assert: no error
- Assert: `len(f.Calls) == 1`
- Assert: `f.Calls[0].Serial == "emulator-5554"`
- Assert: `f.Calls[0].Args == ["shell", "monkey", "-p", "com.example.app", "-c", "android.intent.category.LAUNCHER", "1"]`

---

#### `TestRunLaunch_SearchPath_PmListThenMonkey`

No package provided — fetches package list, opens TUI, then launches selected package.

- Runner queue:
  1. `"package:com.example.foo\npackage:com.example.bar\n"`, `nil` — pm list packages
  2. `""`, `nil` — monkey
- Call: `runLaunch(f, "emulator-5554", "", &stubSelector{pkg: "com.example.bar"})`
- Assert: no error
- Assert: `len(f.Calls) == 2`
- Assert call 0 (pm list packages):
  - `Serial == "emulator-5554"`
  - `Args == ["shell", "pm", "list", "packages"]`
- Assert call 1 (monkey):
  - `Serial == "emulator-5554"`
  - `Args == ["shell", "monkey", "-p", "com.example.bar", "-c", "android.intent.category.LAUNCHER", "1"]`

---

#### `TestRunLaunch_SearchAborted_NoMonkeyCall`

User cancels Package Search — exits silently, monkey is never called.

- Runner queue: `"package:com.example.foo\n"`, `nil` — pm list packages
- Call: `runLaunch(f, "emulator-5554", "", &stubSelector{err: search.ErrAborted})`
- Assert: no error
- Assert: `len(f.Calls) == 1` (pm list only, no monkey)
- Assert: `f.Calls[0].Args[0] != "monkey"` (first arg of first call is not monkey)

---

#### `TestRunLaunch_MonkeyError_ReturnsError`

Monkey call fails — error is propagated to the caller.

- Runner queue: `""`, `errFake("error")` — monkey
- Call: `runLaunch(f, "emulator-5554", "com.example.app", &stubSelector{})`
- Assert: `err != nil`

## Out of Scope

- Restarting the app (force-stop before launch). Users who want a clean start can run `sadb clear-data` first.
- Launching a specific activity within a package.
- Any output on successful launch.
