// Package repotest builds commit histories for tests without a repository.
package repotest

import (
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing"

	"herdr-git-tree/internal/repo"
)

// History builds commits from "name:parent1,parent2" specs, listed newest first.
func History(specs ...string) []*repo.Commit {
	hash := func(name string) plumbing.Hash {
		var h plumbing.Hash
		copy(h[:], name)
		return h
	}
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	var commits []*repo.Commit
	for i, spec := range specs {
		name, parents, _ := strings.Cut(spec, ":")
		c := &repo.Commit{Hash: hash(name), Committed: base.Add(-time.Duration(i) * time.Hour)}
		for parents != "" {
			var p string
			p, parents, _ = strings.Cut(parents, ",")
			c.Parents = append(c.Parents, hash(p))
		}
		commits = append(commits, c)
	}
	return commits
}
