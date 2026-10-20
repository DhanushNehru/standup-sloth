package standup

import (
	"strings"
	"testing"
	"time"
)

func items() []Item {
	return []Item{
		{ID: "a1b2c3d", FullID: "a1b2c3d4e5f60718293a4b5c6d7e8f9a0b1c2d", Title: "Fix login redirect loop", Author: "D", When: time.Now()},
		{ID: "b2c3d4e", FullID: "b2c3d4e5f60718293a4b5c6d7e8f9a0b1c2d3e", Title: "Add CSV export", Author: "D", When: time.Now()},
		{ID: "12", FullID: "12", Title: "Dark mode toggle", Kind: KindPR, When: time.Now()},
	}
}

func TestAcceptsBulletWithValidShortHash(t *testing.T) {
	acc, rej := Check("- Fixed the login redirect loop [a1b2c3d]\n", items())
	if len(acc) != 1 || len(rej) != 0 {
		t.Fatalf("got %d accepted %d rejected, want 1/0", len(acc), len(rej))
	}
	if acc[0].Text != "Fixed the login redirect loop [a1b2c3d]" {
		t.Fatalf("unexpected normalized text: %q", acc[0].Text)
	}
}

func TestNormalizesEchoFormat(t *testing.T) {
	// The format small models drift into: "[id] (commit) Title - sentence"
	out := "[a1b2c3d] (commit) Fix login redirect loop - Stopped the login page looping after sign-in.\n"
	acc, rej := Check(out, items())
	if len(acc) != 1 || len(rej) != 0 {
		t.Fatalf("got %d/%d: %+v", len(acc), len(rej), rej)
	}
	want := "Stopped the login page looping after sign-in [a1b2c3d]"
	if acc[0].Text != want {
		t.Fatalf("got %q, want %q", acc[0].Text, want)
	}
}

func TestAcceptsFullHashAndParens(t *testing.T) {
	acc, _ := Check("- Exported CSVs (a1b2c3d4e5f60718293a4b5c6d7e8f9a0b1c2d)\n", items())
	if len(acc) != 1 || acc[0].Cites[0] != "a1b2c3d" || !strings.HasSuffix(acc[0].Text, "[a1b2c3d]") {
		t.Fatalf("full hash should resolve to short ID, got %+v", acc)
	}
}

func TestAcceptsPRCitation(t *testing.T) {
	for _, line := range []string{
		"- Merged the dark mode toggle [#12]",
		"- Merged the dark mode toggle (#12)",
		"- Merged the dark mode toggle #12",
	} {
		acc, rej := Check(line+"\n", items())
		if len(acc) != 1 || len(rej) != 0 {
			t.Fatalf("%q: got %d accepted %d rejected, want 1/0", line, len(acc), len(rej))
		}
		if !strings.HasSuffix(acc[0].Text, "[#12]") {
			t.Fatalf("%q: PR citation not normalized: %q", line, acc[0].Text)
		}
	}
}

func TestRejectsHallucinatedHash(t *testing.T) {
	acc, rej := Check("- Rewrote the auth layer in Rust [deadbeef]\n", items())
	if len(acc) != 0 || len(rej) != 1 {
		t.Fatalf("hallucination must be rejected, got %+v / %+v", acc, rej)
	}
	if !strings.Contains(rej[0].Reason, "not in the source") {
		t.Fatalf("unexpected reason: %s", rej[0].Reason)
	}
}

func TestRejectsBulletWithoutCitation(t *testing.T) {
	acc, rej := Check("- Had a quick sync with the team\n", items())
	if len(acc) != 0 || len(rej) != 1 || rej[0].Reason != "no citation" {
		t.Fatalf("got %+v / %+v", acc, rej)
	}
}

func TestRejectsMixedRealAndFakeCitations(t *testing.T) {
	acc, rej := Check("- Fixed login and ported it to WASM [a1b2c3d] [cafebabe]\n", items())
	if len(acc) != 0 || len(rej) != 1 || !strings.Contains(rej[0].Reason, "mixed") {
		t.Fatalf("got %+v / %+v", acc, rej)
	}
}

func TestIgnoresPureProse(t *testing.T) {
	out := "Here is your standup:\n- Did work [a1b2c3d]\nHope that helps!\n"
	acc, rej := Check(out, items())
	if len(acc) != 1 || len(rej) != 0 {
		t.Fatalf("got %+v / %+v", acc, rej)
	}
}

func TestHashCollisionAcrossCommitsPrefersFullMatch(t *testing.T) {
	acc, _ := Check("- Work [b2c3d4e5f60718293a4b5c6d7e8f9a0b1c2d3e]\n", items())
	if len(acc) != 1 || acc[0].Cites[0] != "b2c3d4e" {
		t.Fatalf("got %+v", acc)
	}
}

func TestUncovered(t *testing.T) {
	acc, _ := Check("- Did work [a1b2c3d]\n", items())
	u := Uncovered(acc, items())
	if len(u) != 2 || u[0].ID != "b2c3d4e" || u[1].ID != "12" {
		t.Fatalf("got %+v", u)
	}
}

func TestNormalizesEchoWithDash(t *testing.T) {
	acc, rej := Check("[a1b2c3d] - Fixed the redirect loop\n", items())
	if len(acc) != 1 || len(rej) != 0 {
		t.Fatalf("got %d/%d", len(acc), len(rej))
	}
	if acc[0].Text != "Fixed the redirect loop [a1b2c3d]" {
		t.Fatalf("got %q", acc[0].Text)
	}
}
