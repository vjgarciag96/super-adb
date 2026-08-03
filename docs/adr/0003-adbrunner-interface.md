# ADBRunner as the single test seam

All ADB execution goes through a single `ADBRunner` interface: takes a device serial and a list of arguments, returns output and an error. The real implementation shells out to the `adb` binary. Tests inject a fake.

## Context

The tool needs to be testable without a physical device. The question was where to draw the seam: wrap individual commands, wrap the `adb` binary, or wrap at the OS exec level.

## Decision

One interface, one seam, as low as possible without hitting the OS. `ADBRunner` wraps the `adb` binary invocation rather than individual command logic, so tests can assert on what ADB commands were issued — arguments, order, serial — without caring how the code assembled them.

A single shared fake in a test helper means every module reuses the same seam rather than growing its own mocking layer.

## Consequences

- All ADB calls are auditable in tests via the recorded call log on the fake.
- No test requires a connected device or a real `adb` binary.
- Any code path that bypasses `ADBRunner` (e.g., a direct `exec.Command("adb", ...)`) is untestable and should be treated as a bug.
