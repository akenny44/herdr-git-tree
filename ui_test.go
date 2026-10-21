package main

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func TestMouseSelectsClickedCommit(t *testing.T) {
	var m tea.Model = newModel("repo", nil, history("d:c", "c:b", "b:a", "a:"))
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})

	click := func(y int) {
		m, _ = m.Update(tea.MouseMsg{Y: y, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	}
	cursor := func() int { return m.(model).cursor }

	click(3) // header, then rows 0, 1, 2
	if cursor() != 2 {
		t.Fatalf("click on the third row: cursor %d, want 2", cursor())
	}
	click(0)  // header
	click(10) // empty space below the last commit
	if cursor() != 2 {
		t.Fatalf("click outside the commits moved the cursor to %d", cursor())
	}

	m, _ = m.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown, Action: tea.MouseActionPress})
	if cursor() != 3 {
		t.Fatalf("wheel down: cursor %d, want 3", cursor())
	}
}

func TestDetailPanelListsChangedFiles(t *testing.T) {
	m := newModel("repo", nil, history("b:a", "a:"))
	m.width, m.height = 60, 20
	files := []FileChange{{'M', "main.go"}, {'A', "some/deeply/nested/directory/file.go"}}
	for i := range 8 {
		files = append(files, FileChange{'D', fmt.Sprintf("old%d.go", i)})
	}
	m.files[m.commits[0].Hash] = changedFiles{files: files}

	lines := strings.Split(ansi.Strip(m.View()), "\n")
	panel := lines[len(lines)-1-(detailHeight-1):]
	want := []string{
		"│ 10 files changed",
		"│ M main.go",
		"│ A …/nested/directory/file.go",
		"│ D old0.go",
		"│ D old1.go",
		"│ D old2.go",
		"│ … 5 more",
	}
	for i, line := range panel[:len(want)] {
		if _, right, _ := strings.Cut(line, "│"); "│"+right != want[i] {
			t.Errorf("panel line %d = %q, want right side %q", i, line, want[i])
		}
		if col := strings.Index(line, "│"); ansi.StringWidth(line[:col]) != m.width/2 {
			t.Errorf("panel line %d: separator at column %d, want %d", i, ansi.StringWidth(line[:col]), m.width/2)
		}
	}
	if !strings.HasPrefix(panel[0], "commit ") {
		t.Errorf("left side should start with the commit hash: %q", panel[0])
	}
}
