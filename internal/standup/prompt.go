package standup

import (
	"fmt"
	"strings"
	"time"
)

// SystemPrompt constrains the model to rephrasing with citations.
const SystemPrompt = `You turn a developer's git history into their daily standup update.
You write short, plain bullets. You never invent work that is not in the list.
Every bullet you write ends with the citation of the item it came from.`

// BuildPrompt renders the user message: the source items and strict rules.
// Small models rephrase well and reason poorly, so the job is kept to
// rephrasing: roughly one bullet per item, each with its ID cited.
func BuildPrompt(items []Item, window time.Duration) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Here is my git activity from the last %s:\n\n", humanWindow(window))
	for _, it := range items {
		label := "commit"
		if it.Kind == KindPR {
			label = "merged PR"
		}
		fmt.Fprintf(&b, "[%s] (%s) %s\n", it.ID, label, it.Title)
	}
	fmt.Fprintf(&b, `
Write my standup as %d to %d bullet points.

Rules:
- One bullet per item above, in plain words (what changed, not the raw message).
- Start each bullet with "- ".
- End each bullet with that item's citation in brackets, exactly as shown above: [a1b2c3d] for a commit, [#12] for a merged PR.
- Use only citations from the list. Never invent one.
- No headings, no intro, no outro, no numbering. Bullets only.

`, minInt(3, len(items)), minInt(8, len(items)+2))
	return b.String()
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func humanWindow(d time.Duration) string {
	h := int(d.Hours())
	switch {
	case h >= 48 && h%24 == 0:
		return fmt.Sprintf("%d days", h/24)
	case h >= 24:
		return "1 day"
	default:
		return fmt.Sprintf("%d hours", h)
	}
}
