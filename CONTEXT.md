# sadb

A TUI-enhanced wrapper around ADB (Android Debug Bridge) that makes common Android development workflows faster and less error-prone. Invoked as `sadb` — a drop-in for `adb` — with curated subcommands for workflows that raw ADB handles awkwardly.

## Language

**sadb**:
The CLI tool itself. Invoked as `sadb <command>`, mirrors `adb` syntax and passes commands through transparently, intercepting only where it can add value (device selection, package discovery, etc.).
_Avoid_: super-adb (project directory name only)

**Target User**:
Android developers. The JVM is available on their machines, but sadb ships as a Go binary so no JVM is required.
_Avoid_: general developers, end users

**Curated Subcommand**:
A sadb-specific command (e.g. `sadb capture`, `sadb pkg`) that bundles a multi-step ADB workflow into a single ergonomic operation. Distinct from pass-through commands.
_Avoid_: shortcut, alias, macro

**Pass-through Command**:
A sadb invocation that mirrors an `adb` command exactly, with sadb adding only ambient improvements (device selection prompt, package name completion). The underlying ADB command runs unchanged.
_Avoid_: wrapper, proxy

**Active Device**:
The device currently targeted by sadb commands in a terminal session. Persisted via the `SADB_DEVICE` environment variable. Set automatically on first device pick, switchable via `sadb device`.
_Avoid_: selected device, current device, default device

**Device Picker**:
The interactive TUI that appears when multiple devices are connected and no Active Device is set. Shows connected devices and emulators; selection sets the Active Device for the session.
_Avoid_: device selector, device prompt

**Package Search**:
The interactive fuzzy-search TUI that appears when a Curated Subcommand requires a package name and none was provided. Fetches the installed package list from the Active Device and filters in real time as the user types.
_Avoid_: package picker, app search
