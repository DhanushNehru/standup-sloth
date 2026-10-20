package gitlog

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func makeRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test Dev", "GIT_AUTHOR_EMAIL=dev@example.com",
			"GIT_COMMITTER_NAME=Test Dev", "GIT_COMMITTER_EMAIL=dev@example.com",
			"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-b", "main")
	for i, msg := range []string{"first commit", "second commit", "third commit"} {
		if err := os.WriteFile(filepath.Join(dir, "f"+msg[:1]+msg[len(msg)-1:]+".txt"), []byte(msg), 0o644); err != nil {
			t.Fatal(err)
		}
		run("add", "-A")
		run("commit", "-m", msg)
		_ = i
	}
	return dir
}

func TestCollectReturnsCommitsOldestFirst(t *testing.T) {
	dir := makeRepo(t)
	items, err := Collect(context.Background(), dir, time.Hour, "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("want 3 commits, got %d", len(items))
	}
	if items[0].Title != "first commit" || items[2].Title != "third commit" {
		t.Fatalf("wrong order: %q ... %q", items[0].Title, items[2].Title)
	}
	if len(items[0].ID) != 7 || len(items[0].FullID) != 40 {
		t.Fatalf("bad IDs: %q / %q", items[0].ID, items[0].FullID)
	}
	if items[0].Author != "Test Dev" {
		t.Fatalf("bad author: %q", items[0].Author)
	}
}

func TestCollectAuthorFilter(t *testing.T) {
	dir := makeRepo(t)
	items, err := Collect(context.Background(), dir, time.Hour, "Nobody Else", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("author filter should exclude all, got %d", len(items))
	}
}

func TestCollectRespectsWindow(t *testing.T) {
	// A commit dated 2020 followed by three fresh ones: the 1h window must
	// drop only the ancient one.
	dir := t.TempDir()
	run := func(env []string, args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), env...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	base := []string{"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null",
		"GIT_AUTHOR_NAME=Test Dev", "GIT_AUTHOR_EMAIL=dev@example.com",
		"GIT_COMMITTER_NAME=Test Dev", "GIT_COMMITTER_EMAIL=dev@example.com"}
	run(base, "init", "-b", "main")
	commit := func(msg string, extraEnv []string) {
		if err := os.WriteFile(filepath.Join(dir, msg+".txt"), []byte(msg), 0o644); err != nil {
			t.Fatal(err)
		}
		run(base, "add", "-A")
		run(append(base, extraEnv...), "commit", "-m", msg)
	}
	commit("ancient", []string{"GIT_AUTHOR_DATE=2020-01-01T00:00:00Z", "GIT_COMMITTER_DATE=2020-01-01T00:00:00Z"})
	commit("fresh one", nil)
	commit("fresh two", nil)
	commit("fresh three", nil)

	items, err := Collect(context.Background(), dir, time.Hour, "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("want the 3 fresh commits only, got %d", len(items))
	}
	all, err := Collect(context.Background(), dir, 10*365*24*time.Hour, "", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 4 {
		t.Fatalf("wide window should find all 4, got %d", len(all))
	}
}

func TestIsRepo(t *testing.T) {
	if IsRepo(context.Background(), t.TempDir()) {
		t.Fatal("empty dir is not a repo")
	}
	if !IsRepo(context.Background(), makeRepo(t)) {
		t.Fatal("makeRepo should be a repo")
	}
}
