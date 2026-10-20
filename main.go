package main

import (
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-git/go-git/v5"
)

func main() {
	path := "."
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	commits, err := LoadCommits(path)
	if errors.Is(err, git.ErrRepositoryNotExists) {
		fail("%s is not a git repository", path)
	}
	if err != nil {
		fail("%v", err)
	}
	if len(commits) == 0 {
		fail("%s has no commits yet", path)
	}

	if _, err := tea.NewProgram(newModel(path, commits), tea.WithAltScreen()).Run(); err != nil {
		fail("%v", err)
	}
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "git-tree: "+format+"\n", args...)
	os.Exit(1)
}
