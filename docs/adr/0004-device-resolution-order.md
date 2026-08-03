# Device resolution order

Before any command runs, sadb resolves the target device using this priority order:

1. `SADB_DEVICE` set in environment → use it, skip picker
2. Exactly one device connected → use it automatically
3. Multiple devices connected → show Device Picker TUI
4. No devices connected → exit immediately with a descriptive error

## Context

ADB's native behaviour on multi-device setups is to error and ask the user to re-run with `-s <serial>`. This is the core friction sadb is designed to eliminate.

## Decision

Check the environment variable first so that session persistence (ADR 0005) can short-circuit the picker on repeat invocations. Auto-select on a single device removes a redundant interaction. The picker only fires when genuinely needed. Error fast on no devices rather than letting ADB produce its own cryptic output.

## Consequences

- Commands behave identically to `adb -s <serial> ...` once a device is resolved — no other part of the system needs to know how it was chosen.
- `SADB_DEVICE` can go stale if a device disconnects between sessions; this is handled separately (clear the store on disconnect).
