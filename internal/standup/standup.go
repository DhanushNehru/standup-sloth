// Package standup holds the shared types and the prompt-building,
// citation-checking and rendering logic for standup-sloth.
package standup

import "time"

// Item is one citable piece of work: a git commit or a merged pull request.
type Item struct {
	// ID is what the model may cite: a short commit hash ("a1b2c3d")
	// or a PR number without the '#' ("12").
	ID     string
	FullID string // full commit hash, or same as ID for PRs
	Title  string
	Author string
	When   time.Time
	URL    string // set for pull requests
	Kind   Kind
}

// Kind distinguishes commits from pull requests.
type Kind int

const (
	KindCommit Kind = iota
	KindPR
)

// Bullet is an accepted standup line.
type Bullet struct {
	Text  string   // the bullet text without the leading "- "
	Cites []string // item IDs it cited
}

// Rejected is a model-written line the citation check refused, with why.
type Rejected struct {
	Line   string
	Reason string
}
