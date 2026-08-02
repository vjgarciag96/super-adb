# 02 — Device resolution + pass-through forwarding

**What to build:** `sadb <any-adb-command>` works end-to-end. Before every command, device resolution runs: read `SADB_DEVICE` from the environment and use it if set; if unset and exactly one device is connected, use it automatically; if unset and no devices are connected, exit immediately with a clear descriptive error; if unset and multiple devices are connected, exit with an actionable error telling the user to run `sadb device` (the interactive picker is added in ticket 03). The `-s <serial>` flag overrides all of the above, consistent with `adb` behaviour. Any command that does not match a Curated Subcommand is forwarded to `adb` verbatim with the resolved serial injected as `-s <serial>`, producing output identical to running `adb` directly.

**Blocked by:** 01 — Project scaffold + ADBRunner seam.

**Status:** ready-for-agent

- [ ] Device resolution logic reads `SADB_DEVICE` and short-circuits to that device when set
- [ ] Auto-selects the single connected device when exactly one is present
- [ ] Exits with a clear error message when no devices are connected
- [ ] Exits with an actionable error (mentioning `sadb device`) when multiple devices are connected and `SADB_DEVICE` is unset
- [ ] `-s <serial>` flag overrides device resolution
- [ ] Pass-through forwarding injects `-s <resolved-serial>` and delegates to `adb` verbatim
- [ ] Pass-through output is identical to the equivalent `adb` command
- [ ] Tests cover all four device resolution paths using the fake ADBRunner
- [ ] Test asserts pass-through forwards the correct args with the correct serial
