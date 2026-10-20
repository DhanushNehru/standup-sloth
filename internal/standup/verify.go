package standup

import (
	"regexp"
	"strings"
)

// citationRe matches a citation token in brackets or parentheses:
// (a1b2c3d), [a1b2c3d], (#12), or the bare PR form #12.
var citationRe = regexp.MustCompile(`[\(\[]#?([0-9a-fA-F]{7,40}|[0-9]{1,7})[\)\]]|#([0-9]{1,7})\b`)

var (
	bulletRe       = regexp.MustCompile(`^\s*[-*]\s+`)
	leadCiteRe     = regexp.MustCompile(`^\[[^\]]+\]\s*`)
	kindEchoRe     = regexp.MustCompile(`(?i)^\((commit|merged PR|PR)\)\s*`)
	sepAfterEchoRe = regexp.MustCompile(`^\s*[-:–—]\s+`)
)

type lookup struct {
	byID    map[string]Item // canonical ID -> item
	commits map[string]string
	prs     map[string]bool
}

func newLookup(items []Item) lookup {
	l := lookup{byID: map[string]Item{}, commits: map[string]string{}, prs: map[string]bool{}}
	for _, it := range items {
		l.byID[it.ID] = it
		if it.Kind == KindPR {
			l.prs[it.ID] = true
			continue
		}
		l.commits[it.ID] = it.ID
		full := it.FullID
		if full == "" {
			full = it.ID
		}
		// Index every prefix of the full hash from 7 chars up, so a model
		// citing any unambiguous prefix still resolves.
		for n := 7; n <= len(full); n++ {
			l.commits[full[:n]] = it.ID
		}
	}
	return l
}

// resolve returns the canonical item ID for a citation token, or "".
func (l lookup) resolve(tok string) string {
	if len(tok) < 7 {
		if l.prs[tok] {
			return tok
		}
		return ""
	}
	return l.commits[strings.ToLower(tok)]
}

// candidate reports whether a line attempts to be a standup bullet:
// a real list bullet, or the echo format small models drift into
// ("[a1b2c3d] (commit) message - sentence").
func candidate(line string) bool {
	return bulletRe.MatchString(line) || leadCiteRe.MatchString(line)
}

// normalize strips list markers, citation tokens and echoed prefixes, then
// appends one canonical citation per cited item. The model's formatting
// choices never leak into the final standup.
func normalize(text string, cites []string, lk lookup) string {
	text = bulletRe.ReplaceAllString(text, "")
	text = leadCiteRe.ReplaceAllString(text, "")
	text = kindEchoRe.ReplaceAllString(text, "")
	text = bulletRe.ReplaceAllString(text, "") // "[id] - text" echoes leave a second dash
	// Drop an echoed source title ("Add unit test for add() - ...").
	for _, id := range cites {
		title := lk.byID[id].Title
		if title != "" && len(text) >= len(title) && strings.EqualFold(text[:len(title)], title) {
			text = text[len(title):]
			text = sepAfterEchoRe.ReplaceAllString(text, "")
			break
		}
	}
	text = citationRe.ReplaceAllString(text, "")
	text = strings.TrimSpace(text)
	text = strings.TrimRight(text, " .")
	for _, id := range cites {
		if lk.byID[id].Kind == KindPR {
			text += " [#" + id + "]"
		} else {
			text += " [" + id + "]"
		}
	}
	return text
}

// Check splits model output into candidate bullets and verifies that each one
// cites at least one item actually in the source list. Anything else is
// rejected with a human-readable reason, so a confident invention never
// lands silently in the standup.
func Check(output string, items []Item) (accepted []Bullet, rejected []Rejected) {
	lk := newLookup(items)
	for _, raw := range strings.Split(output, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || !candidate(line) {
			continue // headings and prose are ignored, not checked
		}
		matches := citationRe.FindAllStringSubmatch(line, -1)
		if len(matches) == 0 {
			rejected = append(rejected, Rejected{Line: line, Reason: "no citation"})
			continue
		}
		var cites []string
		hadBad := false
		for _, m := range matches {
			tok := m[1]
			if tok == "" {
				tok = m[2] // bare #12 form
			}
			id := lk.resolve(tok)
			if id == "" {
				hadBad = true
				continue
			}
			dup := false
			for _, c := range cites {
				if c == id {
					dup = true
				}
			}
			if !dup {
				cites = append(cites, id)
			}
		}
		switch {
		case len(cites) == 0:
			rejected = append(rejected, Rejected{Line: line, Reason: "cited hash/PR not in the source commits"})
		case hadBad:
			rejected = append(rejected, Rejected{Line: line, Reason: "mixed real and unknown citations"})
		default:
			accepted = append(accepted, Bullet{Text: normalize(line, cites, lk), Cites: cites})
		}
	}
	return accepted, rejected
}

// Uncovered returns source items no accepted bullet cited.
func Uncovered(accepted []Bullet, items []Item) []Item {
	cited := map[string]bool{}
	for _, b := range accepted {
		for _, c := range b.Cites {
			cited[c] = true
		}
	}
	var out []Item
	for _, it := range items {
		if !cited[it.ID] {
			out = append(out, it)
		}
	}
	return out
}
