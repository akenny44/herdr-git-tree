# herdr-git-tree

Interactive terminal viewer for a git repository's commit graph: every commit
reachable from HEAD, local branches, remote branches and tags, with colored
lanes, branch/tag labels and a detail panel for the selected commit.

```
 ●     911ab2d (HEAD -> main) after merge · Alice, 2 hours ago
 │ ●   c93fc58 (hotfix) hotfix urgent · Bob, 3 hours ago
 ●─┼─╮ d8c7098 (tag: v1.0) Merge branch feature · Alice, 1 day ago
 ●─╯ │ 4629b14 fix typo · Alice, 2 days ago
 │   ● d3dec44 (feature) feature finish · Bob, 2 days ago
 ●───╯ c7d66bf add readme · Alice, 3 days ago
```

## Install

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
