# 12 — Tidy minor code smells from issue 03 review

**What to fix:** Three small smells flagged in the post-03 code review, each a one-step clean-up.

### 1. Middle Man — `BubbleTeaPicker.output()`

`output()` is a two-line private method whose only job is to null-check `p.Stderr` and return `io.Discard`. It has one call site and no reuse value. Inline the check directly into `Pick`.

### 2. Data Clump — `chosen` / `Chosen()` on `picker.Model`

Callers of `BubbleTeaPicker.Pick` never interact with `Model` directly (it's an internal implementation detail of the Bubble Tea program). But tests do drive the `Model` and must check `Aborted()` → `Chosen()` → `Selected()` in sequence; the type doesn't enforce ordering. Consider replacing the three exported state fields with a single `Result() (serial string, aborted bool)` accessor, or a small `PickResult` value type, so the caller can't accidentally read `Selected()` without knowing whether the pick was aborted.

### 3. Data Clumps — `picker` + `store` params in `runPassThrough`

`device.Picker` and `device.Store` travel together at every call site (`runPassThrough`, `root.go`, all tests). Bundle them into a `device.ResolveOptions` (or similar) struct so they are passed as one argument. Makes future additions (e.g. a timeout or a logger) a single-site change.

**Blocked by:** 11 — Consolidate Store interfaces (the struct in item 3 uses whichever `Store` type survives that ticket)

**Status:** ready-for-agent

- [ ] `output()` inlined into `Pick`
- [ ] `picker.Model` state query simplified (single accessor or result type)
- [ ] `picker` and `store` bundled into one options/config value passed to `runPassThrough` and `Resolve`
- [ ] All tests updated and passing
