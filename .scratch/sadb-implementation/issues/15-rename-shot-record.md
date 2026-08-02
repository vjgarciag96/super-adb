# 15 — Rename `sadb capture photo` / `sadb capture video` to `sadb shot` / `sadb record`

**What to build:** Promote the capture subcommands from `sadb capture photo` / `sadb capture video` to top-level commands `sadb shot` and `sadb record`. Remove the `capture` parent command. The `--output` flag and all behaviour stay the same; only the command names and their registration change.

**Blocked by:** none — standalone rename on top of the existing #05/#06 implementation.

**Status:** done

- [x] `sadb shot` registered as a top-level Cobra command (replaces `sadb capture photo`)
- [x] `sadb record` registered as a top-level Cobra command (replaces `sadb capture video`)
- [x] `captureCmd` parent and `internal/cli/capture.go` removed or repurposed; new files `internal/cli/shot.go` and `internal/cli/record.go` created
- [x] `--output` flag moved to each new top-level command (no longer needs `PersistentFlags` on a parent)
- [x] `sadb capture` no longer appears in `sadb --help` output
- [x] No changes to `internal/capture/capture.go` — `RunPhoto` / `RunVideo` are unchanged
