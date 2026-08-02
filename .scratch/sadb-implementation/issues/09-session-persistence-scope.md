# 09 — Fix session persistence: terminal-session scope + SADB_DEVICE

**What to fix:** The implementation in ticket 03 chose env-file (`~/.sadb/device`) for session persistence, but the file outlives terminal sessions — all concurrent and future terminals share one global device selection. The spec requires *terminal-session* scope: *"The Active Device persists for the terminal session via `SADB_DEVICE`"* and *"sadb prints an `export SADB_DEVICE=<serial>` instruction for the user's shell to evaluate, or uses a shell integration script."* Neither is satisfied: the env var is never set, and there is no export instruction.

**Decision to make:** Decide definitively between:
1. **Eval-based**: After picking, print `export SADB_DEVICE=<serial>` to stdout. Requires a shell wrapper (`sadb() { eval "$(command sadb "$@")"; }`) that users add to their profile. Clean terminal-session scope; env var is actually set.
2. **Env-file with TTY scoping**: Write to a file keyed by TTY/session ID (e.g. `~/.sadb/sessions/<tty-name>`) so concurrent terminals don't clobber each other. Closer to terminal-session scope without requiring shell setup.
3. **Accept global file with documentation**: Keep `~/.sadb/device` but explicitly document that it is a last-used device, not a per-session value, and update the spec wording accordingly.

**Blocked by:** 03 — Device Picker TUI (done)

**Status:** decided — option 3 (global file, documented)

**Decision:** Keep `~/.sadb/device` as a user-global "last-used device" rather than a per-terminal-session value. Rationale: eval-based export requires shell function setup (user friction); TTY-keyed files add implementation complexity for limited gain. The behavior is documented in the `session` package doc comment. If per-terminal scoping becomes a real user need, the eval path remains available and takes priority (the env var is checked before the file in `device.Resolve`).

Fix 10 (stale serial guard) means the file never silently causes problems: if the stored device is no longer connected, resolution falls through to the picker automatically.

- [x] Decision documented in `internal/session/store.go` package comment
- [x] After device pick, the serial is available to subsequent commands via file (not env var — documented tradeoff)
- [x] Stale file values do not suppress the picker (fixed in issue 10)
- [ ] `sadb device` command (ticket 04) uses the same persistence mechanism
- [x] Eval-based path documented as future upgrade option if per-terminal scope is needed
