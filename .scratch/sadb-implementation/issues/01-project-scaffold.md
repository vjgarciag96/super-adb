# 01 — Project scaffold + ADBRunner seam

**What to build:** A `sadb` binary that compiles and runs. Wire up Cobra as the CLI framework, define the `ADBRunner` interface with a real shell-out implementation that delegates to the `adb` binary, and establish a reusable fake `ADBRunner` test helper that records calls for assertions. Running `sadb --help` should work; no other behaviour is implemented yet.

**Blocked by:** None — can start immediately.

**Status:** ready-for-agent

- [ ] Go module initialised (`go.mod`, `go.sum`), Cobra wired as the root command
- [ ] `ADBRunner` interface defined: accepts a device serial and a list of arguments, returns output and an error
- [ ] Real `ADBRunner` implementation shells out to the `adb` binary on PATH
- [ ] Fake `ADBRunner` test helper records all calls (serial + args) and returns configurable output/errors; lives in a shared test package so all modules can import it
- [ ] `sadb --help` runs without error
- [ ] CI (or `go test ./...`) passes with at least one smoke test exercising the fake
