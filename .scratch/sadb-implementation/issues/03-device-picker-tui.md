# 03 — Device Picker TUI + session persistence

**What to build:** When multiple devices are connected and `SADB_DEVICE` is unset, an interactive Bubble Tea list appears showing all connected devices and emulators. The user selects one and the Active Device is set for the terminal session. The session persistence mechanism (eval-based `export SADB_DEVICE=<serial>` vs. env-file) should be decided and implemented here. This ticket replaces the "multiple devices" error introduced in ticket 02 with the real picker.

**Blocked by:** 02 — Device resolution + pass-through forwarding.

**Status:** ready-for-agent

- [ ] Bubble Tea Device Picker component renders a list of connected devices and emulators
- [ ] Selecting a device sets `SADB_DEVICE` for the terminal session
- [ ] Session persistence mechanism decided and implemented (eval-based or env-file)
- [ ] The "multiple devices" error from ticket 02 is replaced by the Device Picker
- [ ] Pressing Ctrl+C or Escape aborts the picker and exits without running the command
- [ ] Device list is fetched via `ADBRunner` (testable with the fake)
