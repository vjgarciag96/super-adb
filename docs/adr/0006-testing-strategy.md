# Testing strategy: assert on external behaviour via ADBRunner fake

Tests assert on what ADB commands were issued — arguments, order, device serial — not on internal state or which functions were called.

## Context

The system's observable output is the sequence of ADB commands it produces. Internal details (how the device list is stored, which struct holds state) are implementation details that should be free to change.

## Decision

Each test sets up a fake `ADBRunner`, runs a sadb operation, and asserts on the recorded call log and the return value. This maps directly to the user-visible contract: "when I run `sadb shot`, it should issue `screencap`, then `pull`, then delete the temp file — in that order, on the correct device."

Asserting on internal state (e.g., checking a field on a struct after a picker interaction) was rejected because it couples tests to implementation, making refactors break tests without the behaviour actually regressing.

## Consequences

- Refactoring internals does not break tests as long as the ADB call sequence is preserved.
- Tests document the expected ADB command sequences explicitly, making them useful as a spec.
- Any behaviour that doesn't produce an ADB call (pure computation, string formatting) is tested directly without the fake.
