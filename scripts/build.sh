#!/bin/sh
set -eu

# herdr runs builds with its own environment, which may miss the usual Go locations.
PATH="$PATH:/opt/homebrew/bin:/usr/local/bin:/usr/local/go/bin:$HOME/go/bin"

if ! command -v go >/dev/null 2>&1; then
	echo "herdr-git-tree: Go is required to build the plugin, see https://go.dev/dl/" >&2
	exit 1
fi

go build -o bin/git-tree .
