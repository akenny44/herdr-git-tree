package herdr

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/go-git/go-git/v5"
)

const (
	pluginID  = "herdr-git-tree"
	paneTitle = "git tree" // the manifest pane's title, which herdr uses as its label
	tabLabel  = "Git-tree"
)

// Open is the plugin's `open` action: it opens the tree pane in the
// repository the focused pane is in now, not the one it was launched from, or
// switches to the tab already showing that repository.
func Open() error {
	herdr := os.Getenv("HERDR_BIN_PATH")
	if herdr == "" {
		herdr = "herdr"
	}

	var ctx struct {
		WorkspaceID    string `json:"workspace_id"`
		FocusedPaneID  string `json:"focused_pane_id"`
		FocusedPaneCwd string `json:"focused_pane_cwd"`
	}
	if err := json.Unmarshal([]byte(os.Getenv("HERDR_PLUGIN_CONTEXT_JSON")), &ctx); err != nil {
		return fmt.Errorf("reading HERDR_PLUGIN_CONTEXT_JSON: %w", err)
	}

	cwd := ctx.FocusedPaneCwd
	if ctx.FocusedPaneID != "" {
		if live := foregroundCwd(herdr, ctx.FocusedPaneID); live != "" {
			cwd = live
		}
	}

	cwd = repoRoot(cwd)

	if ctx.WorkspaceID != "" && cwd != "" {
		out, err := exec.Command(herdr, "pane", "list", "--workspace", ctx.WorkspaceID).Output()
		if err == nil {
			if tab := findTreeTab(out, cwd); tab != "" {
				return exec.Command(herdr, "tab", "focus", tab).Run()
			}
		}
	}

	args := []string{"plugin", "pane", "open", "--plugin", pluginID, "--entrypoint", "tree", "--focus"}
	if ctx.WorkspaceID != "" {
		args = append(args, "--workspace", ctx.WorkspaceID)
	}
	if cwd != "" {
		args = append(args, "--cwd", cwd)
	}
	cmd := exec.Command(herdr, args...)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return err
	}

	var opened struct {
		Result struct {
			PluginPane struct {
				Pane struct {
					TabID string `json:"tab_id"`
				} `json:"pane"`
			} `json:"plugin_pane"`
		} `json:"result"`
	}
	if err := json.Unmarshal(out, &opened); err != nil || opened.Result.PluginPane.Pane.TabID == "" {
		return fmt.Errorf("reading the opened pane's tab: %s", out)
	}
	return exec.Command(herdr, "tab", "rename", opened.Result.PluginPane.Pane.TabID, tabLabel).Run()
}

func foregroundCwd(herdr, paneID string) string {
	out, err := exec.Command(herdr, "pane", "get", paneID).Output()
	if err != nil {
		return ""
	}
	var resp struct {
		Result struct {
			Pane struct {
				ForegroundCwd string `json:"foreground_cwd"`
			} `json:"pane"`
		} `json:"result"`
	}
	if json.Unmarshal(out, &resp) != nil {
		return ""
	}
	return resp.Result.Pane.ForegroundCwd
}

// repoRoot returns the top of the git worktree containing dir, or dir itself
// when it is not in one (the pane then reports the error).
func repoRoot(dir string) string {
	if dir == "" {
		return ""
	}
	repo, err := git.PlainOpenWithOptions(dir, &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return dir
	}
	wt, err := repo.Worktree()
	if err != nil {
		return dir
	}
	return wt.Filesystem.Root()
}

// findTreeTab returns the tab of a tree pane opened on root, given the JSON
// output of `herdr pane list`.
func findTreeTab(paneList []byte, root string) string {
	var resp struct {
		Result struct {
			Panes []struct {
				Label string `json:"label"`
				Cwd   string `json:"cwd"`
				TabID string `json:"tab_id"`
			} `json:"panes"`
		} `json:"result"`
	}
	if json.Unmarshal(paneList, &resp) != nil {
		return ""
	}
	for _, p := range resp.Result.Panes {
		if p.Label == paneTitle && p.Cwd == root {
			return p.TabID
		}
	}
	return ""
}
