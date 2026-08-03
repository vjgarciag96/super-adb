# Session persistence via SADB_DEVICE environment variable

After the Device Picker selects a device, sadb exports the chosen serial as `SADB_DEVICE`. Subsequent invocations in the same terminal session read this variable and skip the picker.

## Considered Options

- **Environment variable (`SADB_DEVICE`)**: visible to child processes, naturally scoped to the terminal session, no daemon required
- **File-based store (`~/.config/sadb/device`)**: persists across sessions, but "last used globally" is surprising when working with multiple terminals targeting different devices simultaneously
- **Eval-based shell integration**: a wrapper function captures picker output and sets the variable automatically, so the user never has to run `eval $(sadb device)` manually

## Decision

Environment variable scoped to the terminal session. A file-based global store was rejected because it causes the wrong device to be selected when two terminals target different devices — a common Android developer workflow. The variable is set either via `eval`-based shell integration or by the user running `export SADB_DEVICE=<serial>` manually; the integration script is the preferred path.

## Consequences

- Each terminal session is independent — opening a new tab starts with a fresh picker.
- `SADB_DEVICE` can be inspected, overridden, or unset manually at any time.
- Shell integration must be sourced in the user's shell profile for the eval mechanism to work automatically; without it, the picker result is printed but not applied.
