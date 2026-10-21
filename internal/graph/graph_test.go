package graph

import (
	"testing"

	"herdr-git-tree/internal/repo"
	"herdr-git-tree/internal/repo/repotest"
)

func assertGraph(t *testing.T, commits []*repo.Commit, want ...string) {
	t.Helper()
	rows := Layout(commits)
	if len(rows) != len(want) {
		t.Fatalf("got %d rows, want %d", len(rows), len(want))
	}
	for i, r := range rows {
		if got := r.String(); got != want[i] {
			t.Errorf("row %d: got %q, want %q", i, got, want[i])
		}
	}
}

func TestLayoutLinear(t *testing.T) {
	assertGraph(t, repotest.History("c:b", "b:a", "a:"),
		"●",
		"●",
		"●",
	)
}

func TestLayoutBranchAndMerge(t *testing.T) {
	assertGraph(t, repotest.History("m:b,f", "f:b", "b:a", "a:"),
		"●─╮",
		"│ ●",
		"●─╯",
		"●",
	)
}

func TestLayoutTwoBranchTips(t *testing.T) {
	assertGraph(t, repotest.History("x:b", "y:b", "b:a", "a:"),
		"●",
		"│ ●",
		"●─╯",
		"●",
	)
}

func TestLayoutUnrelatedRoots(t *testing.T) {
	assertGraph(t, repotest.History("o2:o1", "m2:m1", "o1:", "m1:"),
		"●",
		"│ ●",
		"● │",
		"  ●",
	)
}

func TestLayoutMergeCrossesLane(t *testing.T) {
	assertGraph(t, repotest.History("p:a", "x:b", "m:c,a", "c:b", "b:a", "a:"),
		"●",
		"│ ●",
		"├─┼─●",
		"│ │ ●",
		"│ ●─╯",
		"●─╯",
	)
}

func TestLayoutThreeLanesJoin(t *testing.T) {
	assertGraph(t, repotest.History("m:a,f", "x:a", "f:a", "a:"),
		"●─╮",
		"│ │ ●",
		"│ ● │",
		"●─╯─╯",
	)
}
