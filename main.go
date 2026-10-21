package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-git/go-git/v5"
)

func main() {
	path := "."
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	if path == "--herdr-open" {
		if err := openInHerdr(); err != nil {
			fail("%v", err)
		}
		return
	}
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
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
	// In a herdr pane the tab closes with the process: keep the message readable.
	if os.Getenv("HERDR_PLUGIN_ENTRYPOINT_ID") != "" {
		fmt.Fprint(os.Stderr, "Press Enter to close.")
		bufio.NewReader(os.Stdin).ReadString('\n')
	}
	os.Exit(1)
}
