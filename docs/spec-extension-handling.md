# Spec: Extension Handling for `shot` and `record`

## Problem

`adb screencap` requires a `.png` filename and `adb screenrecord` requires a `.mp4`
filename on the device side. Today `sadb` passes the user-supplied positional path
straight through to `resolveOutputPath`, so users who omit the extension or use the
wrong one get a silently misnamed or corrupt file on disk.

## Behaviour Rules

These rules apply to the **positional `<path>` argument only**. The `--output` flag
(a directory, not a filename) is unaffected.

| User input | Result |
|---|---|
| No extension (`myscreen`) | Append correct extension → `myscreen.png` / `myscreen.mp4` |
| Correct extension, any case (`.png`, `.PNG`, `.mp4`, `.MP4`) | Pass through unchanged, no warning |
| Wrong extension (`myscreen.jpg`, `myscreen.avi`) | Replace extension, print warning to stderr |

**Warning format** (stderr, wrong-extension case):

```
Warning: .jpg is not valid for shot, saving as myscreen.png
```

— The wrong extension is named, the command is named, the corrected filename is shown.

## Interface Change

`resolveCapturePath` in `internal/cli/output_path.go` gains one new parameter:
`expectedExt string` (e.g. `".png"`, `".mp4"`).

```go
func resolveCapturePath(
    cmd *cobra.Command,
    args []string,
    autoName string,
    expectedExt string,
) (string, error)
```

The extension normalisation logic lives in a new unexported helper in the same file:

```go
// normaliseExtension checks whether path has the expected extension (case-insensitive).
// If the path has no extension, it appends expectedExt and returns (newPath, "").
// If the path has the wrong extension, it replaces it and returns (newPath, warning).
// If the extension is correct (any case), it returns (path, "").
func normaliseExtension(path, expectedExt string) (string, string)
```

The warning string returned by `normaliseExtension` is written to `os.Stderr` by
`resolveCapturePath` when non-empty. The caller (`shot.go`, `record.go`) sees no
change beyond the extra argument.

## Callers

| Command | `expectedExt` |
|---|---|
| `shot` | `".png"` |
| `record` | `".mp4"` |

## Test Seams

All tests live in `internal/cli/`. The confirmed seams are:

### `normaliseExtension` (unit tests — `output_path_test.go`)

Test the pure function directly; no filesystem or cobra dependency.

| Case | Input path | Input ext | Expected output path | Expected warning |
|---|---|---|---|---|
| No extension | `myscreen` | `.png` | `myscreen.png` | `""` |
| Correct extension, lowercase | `myscreen.png` | `.png` | `myscreen.png` | `""` |
| Correct extension, uppercase | `myscreen.PNG` | `.png` | `myscreen.PNG` | `""` |
| Wrong extension | `myscreen.jpg` | `.png` | `myscreen.png` | `"Warning: .jpg is not valid for shot, saving as myscreen.png"` |
| Path with directories, no ext | `~/Desktop/myscreen` | `.png` | `~/Desktop/myscreen.png` | `""` |
| Path with directories, wrong ext | `~/Desktop/myscreen.jpg` | `.png` | `~/Desktop/myscreen.png` | `"Warning: .jpg is not valid for shot, saving as myscreen.png"` |
| Multi-dot name, wrong ext | `my.screen.jpg` | `.png` | `my.screen.png` | `"Warning: .jpg is not valid for shot, saving as my.screen.png"` |

> The warning string for the wrong-extension case must include the original extension,
> the command name, and the corrected filename. The command name is not known inside
> `normaliseExtension`; it is passed in as a `cmdName string` parameter.

Revised signature:

```go
func normaliseExtension(path, expectedExt, cmdName string) (string, string)
```

### `resolveCapturePath` (integration — existing `output_path_test.go`)

Add one test confirming the warning is written to stderr when a wrong extension is
supplied. Use `os.Pipe` or redirect `os.Stderr` for assertion.

### `shot` / `record` commands (CLI layer — existing `*_test.go`)

No new CLI-layer tests required for extension handling; the unit tests on
`normaliseExtension` cover the logic. Existing CLI tests that pass an explicit path
with a correct extension must continue to pass unchanged.

## Out of Scope

- Validating that the *directory* portion of an explicit path exists.
- Normalising case on the accepted extension (`.PNG` stays `.PNG`).
- Any change to `--output` flag behaviour.
- Any change to auto-generated filenames (they already carry the right extension).
