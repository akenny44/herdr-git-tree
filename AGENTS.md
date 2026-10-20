# AGENTS.md

Go TUI that draws a repository's commit graph. Single `main` package, built with
`go build -o git-tree .`.

## Layout

| File | Role |
|---|---|
| `main.go` | Argument (repo path, default `.`), error reporting, starts the Bubble Tea program |
| `repo.go` | `LoadCommits`: reads refs and commits with go-git, labels refs, sorts children-before-parents by committer date (`sortCommits`) |
| `graph.go` | `Layout`: assigns lanes and builds one `Row` of `Cell`s (glyph + fill + colors) per commit |
| `ui.go` | Bubble Tea model: scrolling list, cursor, detail panel, styling with lipgloss |
| `graph_test.go` | Layout and ordering tests, written as expected graph strings via `Row.String()` |

## Constraints

- `Layout` relies on children being listed before their parents; keep `sortCommits` as the only producer of that order.
- One row per commit: merges, branch-outs and joins are drawn on the commit's own row (`╮╭╯╰┤├┼─`), never on extra connector rows.
- The repository is read-only: no go-git call that writes.
- Dependencies: go-git v5, bubbletea v1, lipgloss v1, x/ansi. `x/ansi` must stay compatible with the `x/cellbuf` version in `go.mod`; `go vet` fails inside the module cache when they drift.

## Commands

```sh
go test ./...   # graph layout tests
go vet ./...
gofmt -l .
```
