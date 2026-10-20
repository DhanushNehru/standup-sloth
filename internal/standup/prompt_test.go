package standup

import (
	"strings"
	"testing"
	"time"
)

func TestPromptListsEveryItemAndRules(t *testing.T) {
	p := BuildPrompt(items(), 24*time.Hour)
	for _, want := range []string{"a1b2c3d", "b2c3d4e", "12", "Never invent", "commit", "merged PR"} {
		if !strings.Contains(p, want) {
			t.Fatalf("prompt missing %q:\n%s", want, p)
		}
	}
}

func TestHumanWindow(t *testing.T) {
	cases := map[time.Duration]string{
		24 * time.Hour:  "1 day",
		48 * time.Hour:  "2 days",
		168 * time.Hour: "7 days",
		8 * time.Hour:   "8 hours",
	}
	for d, want := range cases {
		if got := humanWindow(d); got != want {
			t.Fatalf("humanWindow(%v) = %q, want %q", d, got, want)
		}
	}
}
