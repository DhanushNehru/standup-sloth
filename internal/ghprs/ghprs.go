// Package ghprs collects recently merged pull requests via the GitHub CLI.
// It is optional: standup-sloth works fine without gh installed.
package ghprs

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"time"

	"github.com/DhanushNehru/standup-sloth/internal/standup"
)

// Available reports whether the gh CLI is on PATH.
func Available() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

type prJSON struct {
	Number   int       `json:"number"`
	Title    string    `json:"title"`
	MergedAt time.Time `json:"mergedAt"`
	Author   struct {
		Login string `json:"login"`
	} `json:"author"`
	URL string `json:"url"`
}

// Collect returns PRs merged in dir's repo since now-window. It shells out
// to `gh pr list`, which needs gh to be installed and authenticated.
func Collect(ctx context.Context, dir string, window time.Duration) ([]standup.Item, error) {
	since := time.Now().Add(-window).Format("2006-01-02")
	cmd := exec.CommandContext(ctx, "gh", "pr", "list",
		"--state", "merged",
		"--search", "merged:>="+since,
		"--json", "number,title,mergedAt,url,author",
		"--limit", "50",
	)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("gh pr list: %s", string(ee.Stderr))
		}
		return nil, fmt.Errorf("gh pr list: %v", err)
	}
	var prs []prJSON
	if err := json.Unmarshal(out, &prs); err != nil {
		return nil, fmt.Errorf("parsing gh output: %v", err)
	}
	cutoff := time.Now().Add(-window)
	var items []standup.Item
	for _, p := range prs {
		if p.MergedAt.Before(cutoff) {
			continue // --search is day-granular; tighten to the real window
		}
		items = append(items, standup.Item{
			ID:     fmt.Sprintf("%d", p.Number),
			FullID: fmt.Sprintf("%d", p.Number),
			Title:  p.Title,
			Author: p.Author.Login,
			When:   p.MergedAt,
			URL:    p.URL,
			Kind:   standup.KindPR,
		})
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].When.Before(items[j].When) })
	return items, nil
}
