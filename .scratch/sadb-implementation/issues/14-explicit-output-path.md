# 14 — Explicit output path argument for `sadb shot` / `sadb record`

**What to build:** Allow the user to pass an explicit file path as a positional argument to override the auto-generated filename. Today the output name is always derived from the current timestamp (e.g. `photo_20060102_150405.png`). With this change, running `sadb shot /tmp/screen.png` or `sadb record ~/Desktop/demo.mp4` saves directly to that path instead.

If no positional argument is given, the existing auto-generated name in the current working directory (or `--output` directory) is used unchanged.

**Blocked by:** 15 — Rename `capture photo`/`video` to `sadb shot`/`sadb record` (renames the commands this issue targets).

**Status:** done

- [x] `sadb shot [<path>]` accepts an optional positional argument
- [x] `sadb record [<path>]` accepts an optional positional argument
- [x] When `<path>` is provided it is used as the full local file path for the pull destination
- [x] When `<path>` is omitted behaviour is identical to today (auto-generated name under cwd or `--output` dir)
- [x] `--output` and positional `<path>` are mutually exclusive; passing both returns a clear error
- [x] `RunPhoto` / `RunVideo` signatures updated to accept the resolved local path rather than computing it internally (so the CLI layer owns path resolution and the core functions remain testable)
- [x] Tests cover: explicit path used; auto-generated fallback; mutual-exclusion error
