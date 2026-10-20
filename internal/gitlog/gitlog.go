// Package gitlog collects recent commits from a git repository.
package gitlog

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/DhanushNehru/standup-sloth/internal/standup"
)

const fieldSep = "\x1f"

// IsRepo reports whether dir is inside a git work tree.
func IsRepo(ctx context.Context, dir string) bool {
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "rev-parse", "--is-inside-work-tree")
	return cmd.Run() == nil
}

// RepoName returns the repository directory name for display.
func RepoName(ctx context.Context, dir string) string {
	cmd := exec.CommandContext(ctx, "git", "-C", dir, "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return dir
	}
	p := strings.TrimSpace(string(out))
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}

// Collect returns commits in dir since now-window, optionally filtered to an
// author substring. Results are oldest-first so the standup reads in order.
func Collect(ctx context.Context, dir string, window time.Duration, author string, max int) ([]standup.Item, error) {
	since := time.Now().Add(-window).Format(time.RFC3339)
	args := []string{"-C", dir, "log",
		"--since=" + since,
		"--format=%H" + fieldSep + "%h" + fieldSep + "%an" + fieldSep + "%at" + fieldSep + "%s",
	}
	if author != "" {
		args = append(args, "--author="+author)
	}
	if max > 0 {
		args = append(args, "--max-count="+strconv.Itoa(max))
	}
	cmd := exec.CommandContext(ctx, "git", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git log: %v: %s", err, strings.TrimSpace(stderr.String()))
	}
	var items []standup.Item
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		f := strings.Split(line, fieldSep)
		if len(f) != 5 {
			continue
		}
		ts, _ := strconv.ParseInt(f[3], 10, 64)
		items = append(items, standup.Item{
			ID:     strings.ToLower(f[1]),
			FullID: strings.ToLower(f[0]),
			Title:  f[4],
			Author: f[2],
			When:   time.Unix(ts, 0),
			Kind:   standup.KindCommit,
		})
	}
	// git log is newest-first; flip so the prompt and output read in order.
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
	return items, nil
}
