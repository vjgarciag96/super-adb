# 06 — `sadb capture video`

**What to build:** `sadb capture video` records the screen in a single command. It runs `screenrecord` on the device in the foreground; the user presses Ctrl+C to stop recording. On stop, the recorded file is pulled to the current working directory (or `--output <path>`) and the temporary file on the device is deleted. The user ends up with a video file locally without any manual pull step.

**Blocked by:** 05 — `sadb capture photo` (shares the `capture` subcommand scaffold).

**Status:** done

- [x] `sadb capture video` subcommand registered under the existing `capture` parent command
- [x] Runs `adb shell screenrecord` to a generated temp path on the device in the foreground
- [x] Ctrl+C stops the recording and triggers the pull + cleanup sequence
- [x] Pulls the recorded file to the current working directory
- [x] `--output <path>` flag overrides the save location
- [x] Cleans up the temp file on the device after a successful pull
- [x] Reports a clear error if pull fails after recording stops; cleanup is still attempted
- [x] Tests assert the correct ADB call sequence (screenrecord → pull → rm) using the fake ADBRunner
