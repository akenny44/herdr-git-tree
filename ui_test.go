package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMouseSelectsClickedCommit(t *testing.T) {
	var m tea.Model = newModel("repo", history("d:c", "c:b", "b:a", "a:"))
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
