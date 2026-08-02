# 08 — `sadb uninstall` with Package Search

**What to build:** `sadb uninstall` works two ways: if a package name is provided as an argument it forwards directly to `adb uninstall <package>`; if no package name is provided it opens the Package Search TUI so the user can find and select the app to uninstall without knowing the package name upfront.

**Blocked by:** 07 — Package Search TUI.

**Status:** ready-for-agent

- [ ] `sadb uninstall` subcommand registered with Cobra
- [ ] When a package name is provided, forwards `adb uninstall <package>` with the resolved serial
- [ ] When no package name is provided, opens the Package Search TUI and uses the selected package
- [ ] Exits cleanly with no action if the user cancels the Package Search
- [ ] Test asserts direct-path: correct `uninstall <package>` call issued when package is provided
- [ ] Test asserts search-path: `pm list packages` called first, then `uninstall <selected-package>` issued
