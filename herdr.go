package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

const pluginID = "herdr-git-tree"

// openInHerdr is the plugin's `open` action: it opens the tree pane in the
// directory the focused pane is in now, not the one it was launched from.
func openInHerdr() error {
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

	args := []string{"plugin", "pane", "open", "--plugin", pluginID, "--entrypoint", "tree", "--focus"}
	if ctx.WorkspaceID != "" {
		args = append(args, "--workspace", ctx.WorkspaceID)
	}
	if cwd != "" {
		args = append(args, "--cwd", cwd)
	}
	cmd := exec.Command(herdr, args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
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
