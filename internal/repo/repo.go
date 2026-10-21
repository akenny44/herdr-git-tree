package repo

import (
	"container/heap"
	"context"
	"errors"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/utils/merkletrie"
)

type Commit struct {
	Hash      plumbing.Hash
	Parents   []plumbing.Hash
	Author    string
	Email     string
	When      time.Time
	Committed time.Time
	Message   string
	Refs      []string
}

// LoadCommits returns every commit reachable from HEAD, branches, remotes and
// tags, children always before their parents, newest first otherwise.
func LoadCommits(path string) (*git.Repository, []*Commit, error) {
	repo, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return nil, nil, err
	}

	labels := map[plumbing.Hash][]string{}
	var tips []plumbing.Hash

	head, err := repo.Head()
	headBranch := ""
	if err == nil {
		tips = append(tips, head.Hash())
		if head.Name().IsBranch() {
			headBranch = head.Name().Short()
			labels[head.Hash()] = append(labels[head.Hash()], "HEAD -> "+headBranch)
		} else {
			labels[head.Hash()] = append(labels[head.Hash()], "HEAD")
		}
	}

	refs, err := repo.References()
	if err != nil {
		return nil, nil, err
	}
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Type() != plumbing.HashReference {
			return nil
		}
		name := ref.Name()
		hash := ref.Hash()
		switch {
		case name.IsBranch():
			if name.Short() == headBranch {
				return nil
			}
		case name.IsRemote():
		case name.IsTag():
			if tag, err := repo.TagObject(hash); err == nil {
				c, err := tag.Commit()
				if err != nil {
					return nil
				}
				hash = c.Hash
			}
			labels[hash] = append(labels[hash], "tag: "+name.Short())
			tips = append(tips, hash)
			return nil
		default:
			return nil
		}
		labels[hash] = append(labels[hash], name.Short())
		tips = append(tips, hash)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	commits := map[plumbing.Hash]*Commit{}
	queue := tips
	for len(queue) > 0 {
		h := queue[0]
		queue = queue[1:]
		if _, seen := commits[h]; seen {
			continue
		}
		c, err := repo.CommitObject(h)
		if errors.Is(err, plumbing.ErrObjectNotFound) {
			continue // shallow clone: history stops here
		}
		if err != nil {
			return nil, nil, err
		}
		commits[h] = newCommit(c, labels[h])
		queue = append(queue, c.ParentHashes...)
	}

	return repo, sortCommits(commits), nil
}

// FileChange is one line of `git show --name-status`: Status is A, M, D or R.
type FileChange struct {
	Status byte
	Path   string
}

// ChangedFiles lists the files commit h changes against its first parent, so a
// merge shows what it brought into its branch. A root commit adds every file.
func ChangedFiles(repo *git.Repository, h plumbing.Hash) ([]FileChange, error) {
	c, err := repo.CommitObject(h)
	if err != nil {
		return nil, err
	}
	tree, err := c.Tree()
	if err != nil {
		return nil, err
	}
	var parentTree *object.Tree
	if c.NumParents() > 0 {
		parent, err := c.Parent(0)
		if err != nil {
			return nil, err
		}
		if parentTree, err = parent.Tree(); err != nil {
			return nil, err
		}
	}
	changes, err := object.DiffTreeWithOptions(context.Background(), parentTree, tree, object.DefaultDiffTreeOptions)
	if err != nil {
		return nil, err
	}

	files := make([]FileChange, 0, len(changes))
	for _, ch := range changes {
		action, err := ch.Action()
		if err != nil {
			return nil, err
		}
		switch {
		case action == merkletrie.Insert:
			files = append(files, FileChange{'A', ch.To.Name})
		case action == merkletrie.Delete:
			files = append(files, FileChange{'D', ch.From.Name})
		case ch.From.Name != ch.To.Name:
			files = append(files, FileChange{'R', ch.From.Name + " → " + ch.To.Name})
		default:
			files = append(files, FileChange{'M', ch.To.Name})
		}
	}
	return files, nil
}

func newCommit(c *object.Commit, refs []string) *Commit {
	return &Commit{
		Hash:      c.Hash,
		Parents:   c.ParentHashes,
		Author:    c.Author.Name,
		Email:     c.Author.Email,
		When:      c.Author.When,
		Committed: c.Committer.When,
		Message:   c.Message,
		Refs:      refs,
	}
}

// sortCommits orders like `git log --date-order`: a commit is emitted only once
// all its children are, so the graph never draws an edge upwards.
func sortCommits(commits map[plumbing.Hash]*Commit) []*Commit {
	children := map[plumbing.Hash]int{}
	for _, c := range commits {
		for _, p := range c.Parents {
			children[p]++
		}
	}

	ready := &byDate{}
	for h, c := range commits {
		if children[h] == 0 {
			heap.Push(ready, c)
		}
	}

	sorted := make([]*Commit, 0, len(commits))
	for ready.Len() > 0 {
		c := heap.Pop(ready).(*Commit)
		sorted = append(sorted, c)
		for _, p := range c.Parents {
			children[p]--
			if children[p] == 0 {
				if parent, ok := commits[p]; ok {
					heap.Push(ready, parent)
				}
			}
		}
	}
	return sorted
}

type byDate []*Commit

func (b byDate) Len() int { return len(b) }
func (b byDate) Less(i, j int) bool {
	if b[i].Committed.Equal(b[j].Committed) {
		return b[i].Hash.String() < b[j].Hash.String()
	}
	return b[i].Committed.After(b[j].Committed)
}
func (b byDate) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b *byDate) Push(x any)   { *b = append(*b, x.(*Commit)) }
func (b *byDate) Pop() any {
	old := *b
	c := old[len(old)-1]
	*b = old[:len(old)-1]
	return c
}
