package main

import (
	"strings"
	"testing"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
)

// history builds commits from "name:parent1,parent2" specs, listed newest first.
func history(specs ...string) []*Commit {
	hash := func(name string) plumbing.Hash {
		var h plumbing.Hash
		copy(h[:], name)
		return h
	}
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	var commits []*Commit
	for i, spec := range specs {
		name, parents, _ := strings.Cut(spec, ":")
		c := &Commit{Hash: hash(name), Committed: base.Add(-time.Duration(i) * time.Hour)}
		for parents != "" {
			var p string
			p, parents, _ = strings.Cut(parents, ",")
			c.Parents = append(c.Parents, hash(p))
		}
		commits = append(commits, c)
	}
	return commits
}

func assertGraph(t *testing.T, commits []*Commit, want ...string) {
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
	assertGraph(t, history("c:b", "b:a", "a:"),
		"●",
		"●",
		"●",
	)
}

func TestLayoutBranchAndMerge(t *testing.T) {
	assertGraph(t, history("m:b,f", "f:b", "b:a", "a:"),
		"●─╮",
		"│ ●",
		"●─╯",
		"●",
	)
}

func TestLayoutTwoBranchTips(t *testing.T) {
	assertGraph(t, history("x:b", "y:b", "b:a", "a:"),
		"●",
		"│ ●",
		"●─╯",
		"●",
	)
}

func TestLayoutUnrelatedRoots(t *testing.T) {
	assertGraph(t, history("o2:o1", "m2:m1", "o1:", "m1:"),
		"●",
		"│ ●",
		"● │",
		"  ●",
	)
}

func TestLayoutMergeCrossesLane(t *testing.T) {
	assertGraph(t, history("p:a", "x:b", "m:c,a", "c:b", "b:a", "a:"),
		"●",
		"│ ●",
		"├─┼─●",
		"│ │ ●",
		"│ ●─╯",
		"●─╯",
	)
}

func TestLayoutThreeLanesJoin(t *testing.T) {
	assertGraph(t, history("m:a,f", "x:a", "f:a", "a:"),
		"●─╮",
		"│ │ ●",
		"│ ● │",
		"●─╯─╯",
	)
}

func TestSortCommitsChildrenFirst(t *testing.T) {
	commits := history("parent:", "child:parent")
	// child is dated older than its parent: clock skew must not reorder them.
	set := map[plumbing.Hash]*Commit{}
	for _, c := range commits {
		set[c.Hash] = c
	}
	sorted := sortCommits(set)
	if sorted[0] != commits[1] || sorted[1] != commits[0] {
		t.Fatalf("child must come before its parent")
	}
}
