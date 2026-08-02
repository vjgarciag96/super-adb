# 05 — `sadb capture photo`

**What to build:** `sadb capture photo` takes a screenshot in a single command. It runs `screencap` to a temporary path on the device, pulls the file to the current working directory (or to `--output <path>` if provided), then deletes the temporary file from the device. The user ends up with a screenshot file locally without running any additional commands.

**Blocked by:** 02 — Device resolution + pass-through forwarding.

**Status:** done

- [x] `sadb capture photo` subcommand registered under a `capture` parent command
- [x] Runs `adb shell screencap` to a generated temp path on the device
- [x] Pulls the captured file to the current working directory
- [x] `--output <path>` flag overrides the save location
- [x] Cleans up the temp file on the device after a successful pull
- [x] Reports a clear error if screencap or pull fails; temp file cleanup is still attempted
- [x] Tests assert the correct ADB call sequence (screencap → pull → rm) using the fake ADBRunner
