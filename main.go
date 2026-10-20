// standup-sloth drafts your daily standup from git history using a small
// local model, and checks every line against the commits it read.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/DhanushNehru/standup-sloth/internal/ghprs"
	"github.com/DhanushNehru/standup-sloth/internal/gitlog"
	"github.com/DhanushNehru/standup-sloth/internal/ollama"
	"github.com/DhanushNehru/standup-sloth/internal/standup"
)

func main() {
	var (
		since   = flag.Duration("since", 24*time.Hour, "how far back to read git history (e.g. 24h, 48h, 168h for a week)")
		author  = flag.String("author", "", "only use commits whose author matches this substring (e.g. your name or email)")
		model   = flag.String("model", "qwen2.5:0.5b", "Ollama model tag; any Apache-2.0 small instruct model works")
		baseURL = flag.String("ollama", "http://127.0.0.1:11434", "Ollama server URL")
		outPath = flag.String("out", "-", "write the markdown standup here, or - for stdout")
		max     = flag.Int("max", 50, "maximum commits to hand to the model")
		prs     = flag.Bool("prs", false, "also include merged PRs via the gh CLI (needs gh installed and authenticated)")
		verbose = flag.Bool("v", false, "print the raw model output to stderr")
	)
	flag.Parse()
	dir := flag.Arg(0)
	if dir == "" {
		dir = "."
	}

	ctx := context.Background()

	if !gitlog.IsRepo(ctx, dir) {
		fmt.Fprintf(os.Stderr, "standup-sloth: %s is not a git repository\n", dir)
		os.Exit(1)
	}

	commits, err := gitlog.Collect(ctx, dir, *since, *author, *max)
	if err != nil {
		fmt.Fprintf(os.Stderr, "standup-sloth: %v\n", err)
		os.Exit(1)
	}
	items := commits

	if *prs {
		if !ghprs.Available() {
			fmt.Fprintln(os.Stderr, "standup-sloth: --prs given but the gh CLI is not installed; continuing with commits only")
		} else if merged, err := ghprs.Collect(ctx, dir, *since); err != nil {
			fmt.Fprintf(os.Stderr, "standup-sloth: could not list merged PRs (%v); continuing with commits only\n", err)
		} else {
			items = append(items, merged...)
		}
	}

	if len(items) == 0 {
		fmt.Fprintf(os.Stderr, "standup-sloth: no commits in %s in the last %s. Widen --since or drop --author.\n", dir, *since)
		os.Exit(1)
	}

	client := ollama.New(*baseURL)
	if err := client.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "standup-sloth: cannot reach Ollama at %s (%v).\nStart it with `ollama serve` and pull a model, e.g. `ollama pull %s`.\n", *baseURL, err, *model)
		os.Exit(1)
	}

	raw, err := client.Chat(ctx, *model, standup.SystemPrompt, standup.BuildPrompt(items, *since))
	if err != nil {
		fmt.Fprintf(os.Stderr, "standup-sloth: model call failed: %v\n", err)
		fmt.Fprintln(os.Stderr, "Is the model pulled? Try: ollama pull "+*model)
		os.Exit(1)
	}
	if *verbose {
		fmt.Fprintf(os.Stderr, "--- raw model output ---\n%s\n------------------------\n", raw)
	}

	accepted, rejected := standup.Check(raw, items)
	res := standup.Result{
		Repo:     gitlog.RepoName(ctx, dir),
		Window:   *since,
		Model:    *model,
		Date:     time.Now(),
		Accepted: accepted,
		Rejected: rejected,
		Items:    items,
	}

	out := os.Stdout
	if *outPath != "-" {
		f, err := os.Create(*outPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "standup-sloth: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		out = f
	}
	standup.Render(out, res)
}
