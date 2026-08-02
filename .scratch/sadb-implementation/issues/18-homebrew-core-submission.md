# 18 — Homebrew Core Submission

**What to build:** Submit sadb to `Homebrew/homebrew-core` so users can install via `brew install sadb` with no tap step required. Supersedes the custom tap from issue #17.

**Blocked by:** Issue #17 (stable tagged release must exist). homebrew-core also requires the project to be established enough to pass notability review.

**See also:** Issue #17 — Homebrew Distribution (custom tap, the stepping stone).

**Status:** todo

- [ ] Ensure a stable tagged release exists and the custom tap is working (#17)
- [ ] Run `brew audit --strict --new Formula/sadb.rb` locally and fix any issues
- [ ] Verify `brew test sadb` passes (the `assert_match` completion smoke test)
- [ ] Submit PR to `Homebrew/homebrew-core` with `Formula/s/sadb.rb`
- [ ] Once merged, deprecate the custom tap (`vjgarciag96/homebrew-super-adb`) with a notice pointing to homebrew-core
