package herdr

import "testing"

func TestFindTreeTab(t *testing.T) {
	panes := []byte(`{"result":{"panes":[
		{"pane_id":"w:p1","tab_id":"w:t1","cwd":"/repo"},
		{"pane_id":"w:p2","tab_id":"w:t2","cwd":"/other","label":"git tree"},
		{"pane_id":"w:p3","tab_id":"w:t3","cwd":"/repo","label":"git tree"}
	]}}`)

	if got := findTreeTab(panes, "/repo"); got != "w:t3" {
		t.Errorf("got %q, want the tree tab on /repo", got)
	}
	if got := findTreeTab(panes, "/elsewhere"); got != "" {
		t.Errorf("got %q, want no tab", got)
	}
}
