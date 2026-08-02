# 17 — Homebrew Distribution

**What to build:** A Homebrew tap and formula so users can install sadb via `brew install`. The formula must install the Completion Script for zsh automatically so users get completions with no extra steps.

**Blocked by:** A tagged release with a published binary (GitHub Releases or similar) is needed before the formula can reference a download URL.

**See also:** ADR 0001 — Go over Kotlin (chose Go partly for Homebrew compatibility). ADR 0002 — Shell completions (Homebrew is the primary zero-setup distribution path for completions).

**Status:** in progress

- [x] Set up GitHub Actions release workflow: tag triggers a build, produces a Darwin arm64 + amd64 binary, uploads to GitHub Releases
- [ ] Create a Homebrew tap repository (`homebrew-sadb` or similar) — must be done manually on GitHub
- [x] Write the Homebrew formula: `url`, `sha256`, `bin.install`, and `zsh_completion` stanzas — goreleaser auto-generates from `.goreleaser.yml` `brews` config
- [ ] Verify `brew install` installs the binary and activates zsh completions without any user action — blocked on first tagged release
- [x] Update issue #16 checklist item for Homebrew once formula is live
