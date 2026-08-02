# 10 — Guard against stale stored serial suppressing the picker

**What to fix:** `device.Resolve` reads `store.Load()` at priority 3 — before querying ADB — and uses whatever serial is stored as the Active Device without checking whether that device is currently connected. If the user previously picked `emulator-5554` but has since unplugged it and plugged in two different devices, sadb will silently target the stale serial rather than showing the picker again.

The spec says the picker fires when *"multiple devices are connected and no Active Device is set."* A stored-but-disconnected serial should not count as "Active Device set."

**Proposed fix:** After loading a serial from the store, verify it appears in the live `adb devices` output before treating it as the Active Device. If it is not present:
- If exactly one device is connected, auto-select it (and update the store).
- If multiple devices are connected, run the picker.
- If no devices are connected, return `ErrNoDevices`.

Alternatively, only use the store when the stored device appears in the connected list, and fall through to normal resolution otherwise.

**Blocked by:** 03 — Device Picker TUI (done)

**Status:** ready-for-agent

- [ ] A stored serial that is not in the live device list does not suppress the picker
- [ ] Resolution falls through gracefully when stored serial is stale
- [ ] Store is updated after a new pick triggered by a stale value
- [ ] Tests cover stale-serial scenario (fake runner returns device list that excludes the stored serial)
- [ ] Single-device auto-select still works when stored serial is stale
