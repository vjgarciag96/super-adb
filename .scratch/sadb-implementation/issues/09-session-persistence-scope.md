# 09 — Fix session persistence: terminal-session scope + SADB_DEVICE

**What to fix:** The implementation in ticket 03 chose env-file (`~/.sadb/device`) for session persistence, but the file outlives terminal sessions — all concurrent and future terminals share one global device selection. The spec requires *terminal-session* scope: *"The Active Device persists for the terminal session via `SADB_DEVICE`"* and *"sadb prints an `export SADB_DEVICE=<serial>` instruction for the user's shell to evaluate, or uses a shell integration script."* Neither is satisfied: the env var is never set, and there is no export instruction.

**Decision to make:** Decide definitively between:
1. **Eval-based**: After picking, print `export SADB_DEVICE=<serial>` to stdout. Requires a shell wrapper (`sadb() { eval "$(command sadb "$@")"; }`) that users add to their profile. Clean terminal-session scope; env var is actually set.
2. **Env-file with TTY scoping**: Write to a file keyed by TTY/session ID (e.g. `~/.sadb/sessions/<tty-name>`) so concurrent terminals don't clobber each other. Closer to terminal-session scope without requiring shell setup.
3. **Accept global file with documentation**: Keep `~/.sadb/device` but explicitly document that it is a last-used device, not a per-session value, and update the spec wording accordingly.

**Blocked by:** 03 — Device Picker TUI (done)

**Status:** ready-for-agent

- [ ] Decision documented (ADR or inline comment) on which persistence mechanism to use
- [ ] After device pick, `SADB_DEVICE` is set for the current terminal session (or clear docs explain why not)
- [ ] Users are not silently surprised by cross-session state leaking between terminals
- [ ] `sadb device` command (ticket 04) uses the same persistence mechanism
- [ ] Tests cover the chosen persistence path
