package main

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const detailHeight = 8

var (
	laneColors = []lipgloss.Color{"1", "2", "3", "4", "5", "6", "9", "10", "11", "12", "13", "14"}

	hashStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	refStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	tagStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)
	headStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Bold(true)
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)
	boldStyle   = lipgloss.NewStyle().Bold(true)
)

type model struct {
	path    string
	commits []*Commit
	rows    []Row
	cursor  int
	offset  int
	width   int
	height  int
}

func newModel(path string, commits []*Commit) model {
	return model{path: path, commits: commits, rows: Layout(commits)}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) listHeight() int {
	return max(1, m.height-detailHeight-2)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.MouseMsg:
		switch {
		case msg.Button == tea.MouseButtonWheelUp:
			m.cursor--
		case msg.Button == tea.MouseButtonWheelDown:
			m.cursor++
		case msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress:
			// The list starts below the one-line header.
			if row := msg.Y - 1; row >= 0 && row < m.listHeight() && m.offset+row < len(m.commits) {
				m.cursor = m.offset + row
			}
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "up", "k":
			m.cursor--
		case "down", "j":
			m.cursor++
		case "pgup", "ctrl+b":
			m.cursor -= m.listHeight()
		case "pgdown", "ctrl+f", " ":
			m.cursor += m.listHeight()
		case "home", "g":
			m.cursor = 0
		case "end", "G":
			m.cursor = len(m.commits) - 1
		}
	}
	m.cursor = min(max(m.cursor, 0), len(m.commits)-1)
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.listHeight() {
		m.offset = m.cursor - m.listHeight() + 1
	}
	return m, nil
}

func (m model) View() string {
	if m.width == 0 {
		return ""
	}
	var b strings.Builder
	header := fmt.Sprintf("%s  %s", boldStyle.Render(m.path), dimStyle.Render(fmt.Sprintf("%d/%d", m.cursor+1, len(m.commits))))
	b.WriteString(ansi.Truncate(header, m.width, "…") + "\n")

	end := min(m.offset+m.listHeight(), len(m.commits))
	lanes := 0
	for _, row := range m.rows[m.offset:end] {
		lanes = max(lanes, len(row))
	}
	for i := m.offset; i < end; i++ {
		b.WriteString(ansi.Truncate(m.renderLine(i, lanes), m.width, "…") + "\n")
	}
	for i := end - m.offset; i < m.listHeight(); i++ {
		b.WriteString("\n")
	}

	b.WriteString(dimStyle.Render(strings.Repeat("─", m.width)) + "\n")
	for _, line := range m.details() {
		b.WriteString(ansi.Truncate(line, m.width, "…") + "\n")
	}
	b.WriteString(dimStyle.Render("↑/↓ j/k move · PgUp/PgDn page · g/G top/bottom · q quit"))
	return b.String()
}

// renderLine pads the graph to the given number of lanes so hashes line up.
func (m model) renderLine(i, lanes int) string {
	c := m.commits[i]
	var b strings.Builder
	if i == m.cursor {
		b.WriteString(cursorStyle.Render("▌"))
	} else {
		b.WriteString(" ")
	}
	row := m.rows[i]
	for j, cell := range row {
		b.WriteString(laneStyle(cell.Color).Render(string(cell.Glyph)))
		if j < len(row)-1 {
			b.WriteString(laneStyle(cell.FillColor).Render(string(cell.Fill)))
		}
	}
	b.WriteString(strings.Repeat("  ", lanes-len(row)))
	b.WriteString(" " + hashStyle.Render(c.Hash.String()[:7]))
	if len(c.Refs) > 0 {
		b.WriteString(" " + renderRefs(c.Refs))
	}
	subject := firstLine(c.Message)
	if i == m.cursor {
		subject = boldStyle.Render(subject)
	}
	b.WriteString(" " + subject)
	b.WriteString(dimStyle.Render(fmt.Sprintf(" · %s, %s", c.Author, relative(c.When))))
	return b.String()
}

func (m model) details() []string {
	lines := make([]string, 0, detailHeight-1)
	if len(m.commits) > 0 {
		c := m.commits[m.cursor]
		lines = append(lines, hashStyle.Render("commit "+c.Hash.String()))
		if len(c.Refs) > 0 {
			lines[0] += " " + renderRefs(c.Refs)
		}
		lines = append(lines,
			"Author: "+c.Author+" <"+c.Email+">",
			"Date:   "+c.When.Format("Mon Jan 2 15:04:05 2006 -0700"),
		)
		if len(c.Parents) > 1 {
			var parents []string
			for _, p := range c.Parents {
				parents = append(parents, p.String()[:7])
			}
			lines = append(lines, "Merge:  "+strings.Join(parents, " "))
		}
		for _, l := range strings.Split(strings.TrimSpace(c.Message), "\n") {
			lines = append(lines, "    "+l)
		}
	}
	for len(lines) < detailHeight-1 {
		lines = append(lines, "")
	}
	return lines[:detailHeight-1]
}

func renderRefs(refs []string) string {
	parts := make([]string, len(refs))
	for i, r := range refs {
		switch {
		case strings.HasPrefix(r, "HEAD"):
			parts[i] = headStyle.Render(r)
		case strings.HasPrefix(r, "tag: "):
			parts[i] = tagStyle.Render(r)
		default:
			parts[i] = refStyle.Render(r)
		}
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

func laneStyle(color int) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(laneColors[color%len(laneColors)])
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(s), "\n")
	return line
}

func relative(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return plural(int(d.Minutes()), "minute")
	case d < 24*time.Hour:
		return plural(int(d.Hours()), "hour")
	case d < 30*24*time.Hour:
		return plural(int(d.Hours()/24), "day")
	case d < 365*24*time.Hour:
		return plural(int(d.Hours()/24/30), "month")
	default:
		return plural(int(d.Hours()/24/365), "year")
	}
}

func plural(n int, unit string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s ago", unit)
	}
	return fmt.Sprintf("%d %ss ago", n, unit)
}
