# 04 — `sadb device` command

**What to build:** `sadb device` opens the Device Picker at any time, letting the user explicitly switch the Active Device mid-session without starting a new terminal. Reuses the Bubble Tea Device Picker component from ticket 03.

**Blocked by:** 03 — Device Picker TUI + session persistence.

**Status:** done

- [x] `sadb device` subcommand registered with Cobra
- [x] Running `sadb device` opens the Device Picker regardless of the current `SADB_DEVICE` value
- [x] Selecting a device updates the Active Device for the session via the same persistence mechanism established in ticket 03
- [x] `sadb device --help` describes the command correctly
