package repo

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func TestChangedFiles(t *testing.T) {
	dir := t.TempDir()
	repo, err := git.PlainInit(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	wt, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	write := func(name, content string) {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	commit := func() plumbing.Hash {
		if err := wt.AddWithOptions(&git.AddOptions{All: true}); err != nil {
			t.Fatal(err)
		}
		sig := &object.Signature{Name: "t", Email: "t@t", When: time.Now()}
		h, err := wt.Commit("c", &git.CommitOptions{Author: sig})
		if err != nil {
			t.Fatal(err)
		}
		return h
	}

	write("a.txt", "a\n")
	write("b.txt", "b\n")
	root := commit()
	write("a.txt", "a2\n")
	os.Remove(filepath.Join(dir, "b.txt"))
	write("dir/c.txt", "a long enough content to be detected as a rename\n")
	second := commit()
	os.Remove(filepath.Join(dir, "dir/c.txt"))
	write("dir/d.txt", "a long enough content to be detected as a rename\n")
	third := commit()

	for _, tc := range []struct {
		hash plumbing.Hash
		want []FileChange
	}{
		{root, []FileChange{{'A', "a.txt"}, {'A', "b.txt"}}},
		{second, []FileChange{{'M', "a.txt"}, {'D', "b.txt"}, {'A', "dir/c.txt"}}},
		{third, []FileChange{{'R', "dir/c.txt → dir/d.txt"}}},
	} {
		got, err := ChangedFiles(repo, tc.hash)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("ChangedFiles(%s) = %v, want %v", tc.hash.String()[:7], got, tc.want)
		}
	}
}

func TestSortCommitsChildrenFirst(t *testing.T) {
	// child is dated older than its parent: clock skew must not reorder them.
	parent := &Commit{Hash: plumbing.Hash{1}, Committed: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	child := &Commit{Hash: plumbing.Hash{2}, Parents: []plumbing.Hash{parent.Hash}, Committed: parent.Committed.Add(-time.Hour)}
	commits := []*Commit{parent, child}
	set := map[plumbing.Hash]*Commit{}
	for _, c := range commits {
		set[c.Hash] = c
	}
	sorted := sortCommits(set)
	if sorted[0] != commits[1] || sorted[1] != commits[0] {
		t.Fatalf("child must come before its parent")
	}
}
