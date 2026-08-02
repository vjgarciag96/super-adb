# Go over Kotlin for implementation language

sadb targets Android developers who already have a JVM, so distribution was not the deciding factor. Go was chosen deliberately to step outside the author's Kotlin comfort zone. It produces a single static binary with no runtime dependency, integrates naturally with Homebrew, and has a mature TUI ecosystem (Bubble Tea + Lip Gloss). The author is an experienced Kotlin developer using this project as a Go learning opportunity.

## Considered Options

- **Kotlin + Mosaic + Clikt**: familiar, Compose mental model applies, JVM present on target machines — rejected because staying in the comfort zone was the explicit reason not to use it
- **Go + Bubble Tea + Cobra**: single binary, no runtime, idiomatic CLI tooling, good TUI ecosystem
