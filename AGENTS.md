# AGENTS.md

Go TUI that draws a repository's commit graph. Single `main` package, built with
`go build -o git-tree .`.

## Layout

| File | Role |
|---|---|
| `main.go` | Argument (repo path, default `.`, or `--herdr-open`), error reporting, starts the Bubble Tea program |
| `herdr.go` | `--herdr-open`, the herdr plugin's `open` action: reads `HERDR_PLUGIN_CONTEXT_JSON`, takes the focused pane's live `foreground_cwd` (`herdr pane get`), runs `herdr plugin pane open` |
| `herdr-plugin.toml` | herdr plugin manifest: `tree` pane (`bin/git-tree`, new tab) and `open` action |
| `scripts/build.sh` | Plugin build step (`go build -o bin/git-tree .`), run by `herdr plugin install` only |
| `repo.go` | `LoadCommits`: reads refs and commits with go-git, labels refs, sorts children-before-parents by committer date (`sortCommits`) |
| `graph.go` | `Layout`: assigns lanes and builds one `Row` of `Cell`s (glyph + fill + colors) per commit |
| `ui.go` | Bubble Tea model: scrolling list, cursor, detail panel, styling with lipgloss |
| `graph_test.go` | Layout and ordering tests, written as expected graph strings via `Row.String()` |

## Constraints

- `Layout` relies on children being listed before their parents; keep `sortCommits` as the only producer of that order.
- One row per commit: merges, branch-outs and joins are drawn on the commit's own row (`╮╭╯╰┤├┼─`), never on extra connector rows.
- The repository is read-only: no go-git call that writes.
- herdr plugin: the id `herdr-git-tree` is hard-coded in `herdr.go` (`pluginID`) and must match the manifest. Plugins cannot declare keybindings; `prefix+t` lives in the user's `~/.config/herdr/config.toml` and the README. In a plugin pane (`HERDR_PLUGIN_ENTRYPOINT_ID` set), `fail` waits for Enter, since the tab closes with the process.
- `herdr plugin link` skips `[[build]]`: run `sh scripts/build.sh` after every change, then test with `herdr plugin action invoke herdr-git-tree.open`.
- Dependencies: go-git v5, bubbletea v1, lipgloss v1, x/ansi. `x/ansi` must stay compatible with the `x/cellbuf` version in `go.mod`; `go vet` fails inside the module cache when they drift.

## Commands

```sh
go test ./...   # graph layout tests
go vet ./...
gofmt -l .
```
