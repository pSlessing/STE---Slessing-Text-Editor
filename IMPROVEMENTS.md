# STE Code Review — Faults & Potential Improvements

Reviewed: `main.go`, `core/*.go`, `modules/gitPlugin.go`, `Makefile`, `install.sh`, `go.mod`, `readme.md`.
`go vet ./...`-equivalent passes clean and both the main binary and the git plugin build successfully; the issues below are logic, robustness, and maintainability findings found by reading the code, not compiler errors.

## Robustness / correctness edge cases
- **`getCurrentColorPos`** falls back to index `0` ("Black") when the current color isn't found in `colorNames` (`core/visuals.go:242-243`), silently misreporting the selection rather than surfacing that the stored color is out of the known palette (can happen since `tcell.Color` supports arbitrary RGB values but the settings UI only offers 21 named colors).
- **`cmdCommand`'s output/error handling is weak** (`core/coreCommands.go:12-21`) — `cmd.Output()`'s error return is discarded (`outputString, _ := cmd.Output()`), so a failing external command shows an empty status message instead of the error/stderr. `cmd.Output()` only captures stdout; combining stdout+stderr (as the git plugin correctly does with `CombinedOutput()`) would be more useful for a general "run a command" feature.
- **No confirmation/sandboxing around `cmd`/`Command`** — it runs arbitrary shell commands with the editor's own privileges. That's a reasonable "power user" feature (similar to Vim's `:!`), but combined with the no-args panic above and no display of stderr, it's currently more likely to crash or confuse than help. At minimum, guard against empty args and echo failures.


## Testing / process

- **No automated tests anywhere in the repo** (`find . -iname '*_test.go'` returns nothing). Given how much of the logic is fiddly index/offset arithmetic (cursor movement, tab expansion, scrolling, word-jump), this is exactly the kind of code that benefits most from unit tests — e.g. `bufferToVisual`, `insertRune`/`insertEnter`/`deleteAtCursor`, and `getFilesForPrefix` are all pure-enough functions to test without a real terminal.
- **No CI configuration** — `.github/` exists but only check its contents if CI is desired; as of this review there's no automated `go build`/`go vet` gate, so regressions like the panic above would only be caught manually.
- **`install.sh` has a commented-out Go-installation check** (`install.sh:20-26`) with the comment `NEEDS REFACTORING; JUST RUN GO` — if Go isn't installed, the script fails several steps later with a much less clear error (`go: command not found`) instead of the intended friendly message. Either restore the check or delete the dead comment block.
- **Plugin binaries (`*.so`) require the plugin and the host binary to be built with the exact same Go toolchain version** (a hard requirement of Go's `plugin` package) — this is inherent to the chosen plugin mechanism, not a bug, but is worth documenting in the README since `install.sh`/`Makefile` don't pin or verify toolchain versions between building `ste` and building plugins.

## Minor / cosmetic

- `readme.md`'s feature bullet formatting merges two unrelated bullets onto one line: `"Customizability" ... - Fuzzy file search within current directory` (`readme.md:12`) reads as if fuzzy search is a sub-point of customizability rather than its own feature.

- Naming inconsistency: `cmdCommand`'s registered `Name` is `"Command"` (capitalized) while every other built-in command name is lowercase (`core/editor.go:146`) — `ExecuteCommand` lowercases user input before lookup (`strings.ToLower(parts[0])`, `core/editor.go:210`), so the capitalized primary name is effectively unreachable by typing `command`; only the `cmd` alias works. Should be `"command"` for consistency and so `help` lists something typeable.