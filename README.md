# herdr-git-tree

Interactive terminal viewer for a git repository's commit graph: every commit
reachable from HEAD, local branches, remote branches and tags, with colored
lanes, branch/tag labels and a detail panel for the selected commit: its
metadata and message on the left, the files it changes on the right.

```
 ●     911ab2d (HEAD -> main) after merge · Alice, 2 hours ago
 │ ●   c93fc58 (hotfix) hotfix urgent · Bob, 3 hours ago
 ●─┼─╮ d8c7098 (tag: v1.0) Merge branch feature · Alice, 1 day ago
 ●─╯ │ 4629b14 fix typo · Alice, 2 days ago
 │   ● d3dec44 (feature) feature finish · Bob, 2 days ago
 ●───╯ c7d66bf add readme · Alice, 3 days ago
```

## herdr plugin

In [herdr](https://herdr.dev), `prefix+t` opens the graph of the focused pane's
repository in a new `Git-tree` tab; `q` closes it.

```sh
herdr plugin install akenny44/herdr-git-tree
```

Installing builds the binary, so Go must be on the machine. herdr plugins
cannot declare keybindings, so add the default one to
`~/.config/herdr/config.toml`, then run `herdr server reload-config`:

```toml
[[keys.command]]
key = "prefix+t"
type = "plugin_action"
command = "herdr-git-tree.open"
description = "open git tree in a new tab"
```

The tab shows the repository the focused pane is in now (its foreground
process's cwd), not the one it was started in. If the workspace already has a
`Git-tree` tab on that repository, `prefix+t` switches to it instead of opening
another one.

To work on the plugin from a local checkout:

```sh
sh scripts/build.sh                # herdr plugin link does not build
herdr plugin link "$PWD"
```

## Standalone install

Requires Go 1.27+.

```sh
go build -o git-tree .
```

`go install .` also works and installs the binary as `herdr-git-tree` in
`$GOBIN`.

The result is a single binary; it reads the repository with
[go-git](https://github.com/go-git/go-git) and does not need the `git` command.

## Usage

```sh
git-tree          # repository containing the current directory
git-tree <path>   # any other repository
```

| Key | Action |
|---|---|
| `↑` / `↓`, `k` / `j` | Move the selection |
| `PgUp` / `PgDn`, `Ctrl+B` / `Ctrl+F`, `Space` | Page up / down |
| `g` / `G`, `Home` / `End` | First / last commit |
| `q`, `Esc`, `Ctrl+C` | Quit |
| Click a commit | Select it and show it in the detail panel |
| Mouse wheel | Move the selection |

The repository is only read, never modified. Commits are ordered like
`git log --all --date-order`. In a shallow clone, the graph stops where the
history does.

## Development

```sh
go test ./...
go vet ./...
gofmt -l .
```

See [AGENTS.md](AGENTS.md) for the code layout and its constraints.
