# 07 — Package Search TUI

**What to build:** A self-contained Bubble Tea component that fetches the installed package list from the Active Device, renders it as a live-filtered list as the user types, and returns the selected package name to its caller. This component is the building block for any Curated Subcommand that requires a package name — ticket 08 wires it into `sadb uninstall`.

**Blocked by:** 02 — Device resolution + pass-through forwarding.

**Status:** ready-for-agent

- [ ] Runs `adb shell pm list packages` on the Active Device and parses the output into a package list
- [ ] Bubble Tea component renders the list and filters it in real time as the user types
- [ ] Selecting a package returns the package name to the caller
- [ ] Pressing Ctrl+C or Escape aborts the search and returns no selection
- [ ] Package list is fetched via `ADBRunner` (testable with the fake)
- [ ] Test asserts that the correct `pm list packages` call is made
- [ ] Test asserts that the selected package name is returned correctly
