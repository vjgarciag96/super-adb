# 11 — Consolidate duplicate Store interfaces

**What to fix:** `device.Store` and `session.Store` are two separate interface declarations with identical method sets (`Load() string`, `Save(serial string) error`). The concrete implementations (`session.FileStore`, `session.NoopStore`) satisfy both, but any future change to the contract requires two edits in lockstep — a Shotgun Surgery smell.

**Proposed fix:** Define the canonical `Store` interface once in `internal/session` (it owns the persistence concern). Have `internal/device` import and reference `session.Store` directly, removing the re-declaration in `device`. Update `Resolve` signature and all call sites.

**Blocked by:** 03 — Device Picker TUI (done)

**Status:** done

- [ ] `Store` interface declared exactly once, in `internal/session`
- [ ] `device.Resolve` accepts `session.Store` (or the interface is re-exported from `session` as the canonical type)
- [ ] No duplicate interface declarations remain
- [ ] All tests continue to pass
