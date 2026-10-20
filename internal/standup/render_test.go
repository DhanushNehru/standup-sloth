package standup

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestRenderContainsSectionsAndHonesty(t *testing.T) {
	var buf bytes.Buffer
	Render(&buf, Result{
		Repo:   "demo",
		Window: 24 * time.Hour,
		Model:  "qwen2.5:0.5b",
		Date:   time.Date(2026, 10, 20, 9, 0, 0, 0, time.Local),
		Accepted: []Bullet{
			{Text: "Fixed the login redirect loop [a1b2c3d]", Cites: []string{"a1b2c3d"}},
		},
		Rejected: []Rejected{
			{Line: "Rewrote auth in Rust [deadbeef]", Reason: "cited hash/PR not in the source commits"},
		},
		Items: items(),
	})
	out := buf.String()
	for _, want := range []string{
		"Standup draft - 2026-10-20",
		"Fixed the login redirect loop",
		"Rejected by the citation check",
		"deadbeef",
		"Source commits the model read",
		"Not covered by the draft",
		"Edit it before you post it",
		"qwen2.5:0.5b",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}
