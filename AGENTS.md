# AGENTS.md

Go TUI that draws a repository's commit graph. `main` stays at the module root,
so it builds with `go build -o git-tree .`; everything else lives in
`internal/`, one package per concern (`repo` ← `graph` ← `ui`; `herdr` stands
alone).

## Layout

| File | Role |
|---|---|
| `main.go` | Argument (repo path, default `.`, or `--herdr-open`), error reporting, starts the Bubble Tea program |
| `internal/herdr/herdr.go` | `herdr.Open`, run by `--herdr-open`, the herdr plugin's `open` action: reads `HERDR_PLUGIN_CONTEXT_JSON`, takes the focused pane's live `foreground_cwd` (`herdr pane get`) up to its repo root, then focuses the workspace's tree tab on that root (`findTreeTab`) or runs `herdr plugin pane open` and renames the new tab `Git-tree` (`tabLabel`) |
| `herdr-plugin.toml` | herdr plugin manifest: `tree` pane (`bin/git-tree`, new tab) and `open` action |
| `scripts/build.sh` | Plugin build step (`go build -o bin/git-tree .`), run by `herdr plugin install` only |
| `internal/repo/repo.go` | `LoadCommits`: reads refs and commits with go-git, labels refs, sorts children-before-parents by committer date (`sortCommits`); `ChangedFiles`: a commit's files against its first parent (A/M/D/R, renames detected) |
| `internal/graph/graph.go` | `Layout`: assigns lanes and builds one `Row` of `Cell`s (glyph + fill + colors) per commit |
| `internal/ui/ui.go` | `ui.New`, the Bubble Tea model: scrolling list, cursor, detail panel split in two (commit info left, changed files right, diffed lazily on selection and cached per hash), styling with lipgloss |
| `internal/repo/repotest/repotest.go` | `History("b:a", "a:")`: commits built from name:parents specs, for tests outside `repo` |
| `internal/repo/repo_test.go` | `ChangedFiles` on a temporary repository; `sortCommits` ordering (built inline: `repotest` imports `repo`) |
| `internal/graph/graph_test.go` | Layout tests, written as expected graph strings via `Row.String()` |
| `internal/ui/ui_test.go` | Mouse selection and the detail panel's file list |
| `internal/herdr/herdr_test.go` | `findTreeTab` on a `herdr pane list` output |

## Constraints

- `Layout` relies on children being listed before their parents; keep `sortCommits` as the only producer of that order.
- One row per commit: merges, branch-outs and joins are drawn on the commit's own row (`╮╭╯╰┤├┼─`), never on extra connector rows.
- The repository is read-only: no go-git call that writes.
- herdr plugin: the id `herdr-git-tree` and the pane title `git tree` are hard-coded in `internal/herdr/herdr.go` (`pluginID`, `paneTitle`) and must match the manifest; an existing tree tab is found by that label plus its `cwd` (the repo root passed as `--cwd`). Plugins cannot declare keybindings; `prefix+t` lives in the user's `~/.config/herdr/config.toml` and the README. In a plugin pane (`HERDR_PLUGIN_ENTRYPOINT_ID` set), `fail` (`main.go`) waits for Enter, since the tab closes with the process.
- `herdr plugin link` skips `[[build]]`: run `sh scripts/build.sh` after every change, then test with `herdr plugin action invoke herdr-git-tree.open`.
- Dependencies: go-git v5, bubbletea v1, lipgloss v1, x/ansi. `x/ansi` must stay compatible with the `x/cellbuf` version in `go.mod`; `go vet` fails inside the module cache when they drift.

## Commands

```sh
go test ./...   # graph layout tests
go vet ./...
gofmt -l .
```
