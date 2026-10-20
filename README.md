# standup-sloth

Drafts your daily standup from your git history with a small local model, and checks every line against the commits it actually read.

Writing a standup is a chore. Handing it to a language model is risky: small models invent work that never happened. standup-sloth does both jobs - the model drafts, then a citation check throws out any line that cannot point at a real commit. What survives is a draft you can trust enough to edit.

Everything runs on your machine. Your git history never leaves it.

## How it works

1. **Collect** - reads the last N hours of commits from a repo (optionally merged PRs too, via the `gh` CLI).
2. **Draft** - a small instruct model running in [Ollama](https://ollama.com) turns them into bullets. The default model is `qwen2.5:0.5b` (Apache 2.0, ~400 MB, runs on a laptop CPU).
3. **Check** - every bullet must cite the commit it came from. Lines with no citation, or with an invented one, are moved to a "Rejected" section with the reason. Commits the model skipped are listed as "Not covered".
4. **Render** - prints the standup as markdown.

## Install

```sh
git clone https://github.com/DhanushNehru/standup-sloth
cd standup-sloth
go build -o standup-sloth .
```

You also need [Ollama](https://ollama.com/download) and a model:

```sh
ollama pull qwen2.5:0.5b
ollama serve
```

## Usage

```sh
./standup-sloth /path/to/your/repo
```

Writes markdown to stdout, so `./standup-sloth . > standup.md` saves it.

| Flag | Default | What it does |
| --- | --- | --- |
| `--since` | `24h` | How far back to read, e.g. `48h`, `168h` for a week |
| `--author` | (all) | Only commits whose author matches this substring |
| `--model` | `qwen2.5:0.5b` | Any Ollama instruct model tag |
| `--ollama` | `http://127.0.0.1:11434` | Ollama server URL |
| `--prs` | off | Also include merged PRs (needs the `gh` CLI, authenticated) |
| `--max` | `50` | Max commits handed to the model |
| `--out` | `-` | Output file, or `-` for stdout |
| `--v` | off | Print the raw model output to stderr |

## Real sample output

From a real run against a scratch repo with 7 commits, using `qwen2.5:1.5b` on 2 CPU cores (about 90 seconds). Note the rejected line: the model invented a `[#12]` PR that does not exist, and the check caught it.

```markdown
# Standup draft - 2026-10-20

_Repo: `demo-repo` - window: last 1 day - model: `qwen2.5:1.5b`_

## What I worked on

- Initial project scaffold with README added to master branch [733977d]
- Added math helper function "add()" and wire up CLI entry point for demo sum printing [fee7a3d]
- Created sub-function "sub()" with documentation update to README file [65f9558]
- Ignored node_modules directory and logs folder, updating .gitignore accordingly [aebe368]

## Not covered by the draft

- `f1d35af` Wire up CLI entry point printing a demo sum
- `d1039e8` Add unit test for add()
- `2faef3d` Add sub() helper

## Rejected by the citation check

- Implemented unit test for the add() method in a separate commit [#12] (cited hash/PR not in the source commits)
```

## Honest limitations

This is a small-model tool. Read these before relying on it:

- **The draft rephrases commit messages; it does not understand your work.** "Fixed the off-by-one in the retry loop" becomes whatever the model thinks that means. Edit before posting.
- **The citation check catches invented or missing citations only.** The model can still attach the wrong *real* commit to a line (you can see mild misattribution in the sample above). The check cannot catch that, which is why every line keeps its citation: you can spot a wrong one at a glance.
- **With `qwen2.5:0.5b`** the bullets often echo commit messages verbatim. `qwen2.5:1.5b` writes noticeably better prose for roughly double the time and memory (~1 GB). Bigger models work too via `--model`.
- **Slow on CPU:** tens of seconds to a couple of minutes per run on a small machine.
- **Tested on Linux only** (Go 1.22+). Not tested on macOS, Windows, or GPU runners. Pull requests welcome.

## Models and licenses

Any Ollama instruct model works with `--model`. The defaults were chosen for license and size: [Qwen2.5-0.5B-Instruct](https://huggingface.co/Qwen/Qwen2.5-0.5B-Instruct) and [Qwen2.5-1.5B-Instruct](https://huggingface.co/Qwen/Qwen2.5-1.5B-Instruct) are both Apache 2.0, so commercial use is fine. If you swap in another model, check its license yourself.

## Why a local model

Standups describe work that may be private: client repos, internal names, unreleased features. A local open-weight model means the whole pipeline - collection, drafting, checking - runs offline, with no API key, no usage bill, and no code leaving the machine.

## Contributing

Issues and PRs are welcome. Good first contributions: better prompts for tiny models, more output formats (plain text, Slack), a `--yesterday` shortcut, and tests on other platforms.

## License

[MIT](LICENSE)
