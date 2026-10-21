package graph

import (
	"strings"

	"github.com/go-git/go-git/v5/plumbing"

	"herdr-git-tree/internal/repo"
)

// Cell is one lane of a graph row: its glyph, then the character drawn
// between it and the next lane (a space, or ─ when an edge crosses).
type Cell struct {
	Glyph     rune
	Fill      rune
	Color     int
	FillColor int
}

type Row []Cell

func (r Row) String() string {
	var b strings.Builder
	for _, c := range r {
		b.WriteRune(c.Glyph)
		b.WriteRune(c.Fill)
	}
	return strings.TrimRight(b.String(), " ")
}

// Layout assigns each commit a lane and draws one graph row per commit.
// commits must list children before parents.
func Layout(commits []*repo.Commit) []Row {
	var lanes []plumbing.Hash // the commit each lane is waiting for
	var colors []int
	nextColor := 0

	allocate := func() int {
		for i, h := range lanes {
			if h.IsZero() {
				colors[i] = nextColor
				nextColor++
				return i
			}
		}
		lanes = append(lanes, plumbing.ZeroHash)
		colors = append(colors, nextColor)
		nextColor++
		return len(lanes) - 1
	}

	rows := make([]Row, len(commits))
	for i, c := range commits {
		col := -1
		var merging []int
		for j, h := range lanes {
			if h == c.Hash {
				if col < 0 {
					col = j
				} else {
					merging = append(merging, j)
				}
			}
		}
		if col < 0 {
			col = allocate()
		}

		var row Row
		cell := func(j int) *Cell {
			for len(row) <= j {
				row = append(row, Cell{Glyph: ' ', Fill: ' '})
			}
			return &row[j]
		}
		for j, h := range lanes {
			if !h.IsZero() {
				*cell(j) = Cell{Glyph: '│', Fill: ' ', Color: colors[j]}
			}
		}
		*cell(col) = Cell{Glyph: '●', Fill: ' ', Color: colors[col]}

		type edge struct{ to, color int }
		var edges []edge
		for _, j := range merging {
			*cell(j) = Cell{Glyph: pick(j > col, '╯', '╰'), Fill: ' ', Color: colors[j]}
			edges = append(edges, edge{j, colors[j]})
		}

		if len(c.Parents) == 0 {
			lanes[col] = plumbing.ZeroHash
		} else {
			lanes[col] = c.Parents[0]
		}
		for n, p := range c.Parents {
			if n == 0 || p == c.Parents[0] {
				continue
			}
			k := -1
			for j, h := range lanes {
				if h == p && j != col {
					k = j
					break
				}
			}
			if k >= 0 {
				*cell(k) = Cell{Glyph: pick(k > col, '┤', '├'), Fill: ' ', Color: colors[k]}
			} else {
				k = allocate()
				lanes[k] = p
				*cell(k) = Cell{Glyph: pick(k > col, '╮', '╭'), Fill: ' ', Color: colors[k]}
			}
			edges = append(edges, edge{k, colors[k]})
		}

		// Freed only now so a merge parent cannot reuse a lane ending on this row.
		for _, j := range merging {
			lanes[j] = plumbing.ZeroHash
		}

		for _, e := range edges {
			lo, hi := min(col, e.to), max(col, e.to)
			for m := lo; m < hi; m++ {
				row[m].Fill = '─'
				row[m].FillColor = e.color
				if m > lo {
					switch row[m].Glyph {
					case ' ':
						row[m] = Cell{Glyph: '─', Fill: '─', Color: e.color, FillColor: e.color}
					case '│':
						row[m].Glyph = '┼'
					}
				}
			}
		}

		for len(lanes) > 0 && lanes[len(lanes)-1].IsZero() {
			lanes = lanes[:len(lanes)-1]
			colors = colors[:len(colors)-1]
		}
		rows[i] = row
	}
	return rows
}

func pick(cond bool, a, b rune) rune {
	if cond {
		return a
	}
	return b
}
