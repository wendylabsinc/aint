# aint improve — mine Claude Code session history for incidents

Date: 2026-08-03

## Purpose

`aint improve` scans every Claude Code session transcript on the machine
(`~/.claude/projects/**/*.jsonl`), finds human messages that show
disagreement or frustration with what Claude just did, and turns each into
an **incident**: a short analysis of what happened and why, plus (where
applicable) a suggested `aint` check / linter rule to catch the mistake
mechanically next time, and a suggested documentation/memory change to
prevent it via upfront knowledge instead.

This closes the loop on `aint` itself: `aint check`/`hook` catch mistakes in
code and shell commands as they happen; `aint improve` looks backward at
mistakes that *weren't* caught, and proposes how to catch (or avoid) the
next one.

## Non-goals (v1)

- No auto-apply of suggestions — `aint improve` only writes a report; adding
  a new check, editing a memory file, or updating docs is a manual follow-up.
- No web/API calls of its own — reasoning is delegated to the user's
  already-installed `claude` CLI (`claude -p`), not a direct Anthropic API
  call. `aint check`/`list`/`hook`/`install` remain fully offline; `improve`
  is the one command in the binary that shells out.
- No cross-machine sync of the offset cache or report history.
- No incident deduplication across runs beyond the offset cache (a message
  processed once and judged not-an-incident is never re-asked about, but no
  attempt is made to merge similar incidents across sessions).

## Architecture

```
internal/improve/
  transcript.go   # walks ~/.claude/projects, parses session JSONL into typed turns
  heuristics.go   # regex/keyword signal detection over human user messages
  state.go        # per-session-file processed-line cache
  context.go      # builds bounded "what happened right before this" context per candidate
  analyze.go      # shells out to `claude -p`, parses structured JSON verdict
  report.go       # renders confirmed incidents to markdown, grouped by project/session
cmd/aint/improve.go   # runImprove(args): flag parsing, wires the above together
```

## Data model

**Session file schema** (from inspecting real transcripts): each line is a
JSON object with a `type` field. The two that matter:

```go
// type: "user" — a human-typed message has message.content as a plain
// string and origin.kind == "human". A type:"user" entry whose
// message.content is an array of {type:"tool_result",...} blocks instead
// is a tool result being echoed back, not something a human typed, and is
// never a heuristic candidate.
type userLine struct {
    Type      string `json:"type"` // "user"
    Message   struct {
        Role    string          `json:"role"`
        Content json.RawMessage `json:"content"` // string OR []block
    } `json:"message"`
    Timestamp string `json:"timestamp"`
    SessionID string `json:"sessionId"`
    CWD       string `json:"cwd"`
    Origin    struct {
        Kind string `json:"kind"` // "human" for real typed input
    } `json:"origin"`
}

// type: "assistant" — message.content is an array of blocks:
// {type:"text", text} or {type:"tool_use", name, input}.
type assistantLine struct {
    Type    string `json:"type"` // "assistant"
    Message struct {
        Content []json.RawMessage `json:"content"`
    } `json:"message"`
    Timestamp string `json:"timestamp"`
}
```

Other `type` values (`queue-operation`, `system`, `mode`,
`permission-mode`, `bridge-session`, `attachment`, `ai-title`,
`file-history-snapshot`, `file-history-delta`, `pr-link`, …) are skipped —
`transcript.go` only decodes `user` and `assistant` lines, and does so
generically enough that unrecognized fields are ignored rather than causing
a parse error.

This schema was reverse-engineered from real transcripts in one project
folder; `origin.kind == "human"` is the discriminator used to exclude
tool-result-shaped `user` lines. Implementation should sample a handful of
session files across different entrypoints (CLI, IDE extension, etc.)
before finalizing the filter, in case older sessions or other entrypoints
lack `origin` — falling back to "content is a plain string" alone if so.

**In-package types:**

```go
type Candidate struct {
    SessionFile string    // absolute path to the .jsonl
    Line        int       // 1-based line number within the file
    Timestamp   time.Time
    Project     string    // cwd from the session line
    SessionID   string
    Text        string    // the flagged human message
    Signals     []string  // which heuristic categories matched, e.g. ["correction", "frustration-language"]
    Context     string    // preceding turn(s), truncated — built by context.go
}

type Incident struct {
    Candidate
    Summary             string
    RootCause           string
    AintRuleSuggestion  string // "" if not applicable
    LintRuleSuggestion  string // "" if not applicable
    DocMemorySuggestion string // "" if not applicable
    AnalysisFailed      string // non-empty if claude invocation failed; other fields empty in that case
}
```

## Flow

1. Load the offset cache from `--state` (default `~/.aint/improve-state.json`):
   `map[sessionFilePath]lastProcessedLine`. `--full` skips this load
   entirely and scans every file from line 0, as if the cache were empty.
2. Walk `--dir` (default `~/.claude/projects`) for `*.jsonl` files.
3. For each file, read line by line, skipping up to the cached offset
   (new files start at 0). Decode each remaining line; keep only `user`
   lines with string `message.content` and `origin.kind == "human"`.
4. Run every heuristic signal (see below) against the message text;
   any match produces a `Candidate` (`Signals` records which categories
   fired). Track the running line count regardless of match, so
   non-matching lines are never rescanned.
5. Build `Context` for each candidate: the assistant text + tool_use/
   tool_result turns since the *previous* human message, summarized
   (tool calls rendered as `name(input summary)`, results truncated),
   capped at a fixed character budget (~4000 chars) so the `claude -p`
   prompt stays bounded regardless of how verbose the surrounding turn was.
6. Process candidates file-by-file, in line order, up to `--limit`
   (default 50) total `claude` calls for the run. For each: call
   `analyze.go`, which prompts `claude -p` with the candidate + context and
   asks for strict JSON (schema below). On success with `is_incident: true`,
   append to the incident list. On success with `is_incident: false`, drop
   it. On any failure (binary missing, non-zero exit, timeout, unparsable
   JSON), keep it as an `Incident` with `AnalysisFailed` set and every
   other field empty, so nothing found by the heuristic is silently lost —
   it just shows up in the report flagged as unanalyzed.
7. After each file finishes (or the run hits `--limit` partway through
   it), persist that file's new offset immediately: full EOF if every
   candidate in it was processed, otherwise the line right before the
   first unprocessed candidate (so already-confirmed lines before it
   aren't rescanned, but the limited-out candidate(s) are retried next
   run). This also means a Ctrl-C mid-run only loses the file currently
   in flight, not prior progress.
8. Render all collected incidents to markdown (`report.go`), grouped by
   `Project` → `SessionID`, chronological within a session. Write to
   `--out` (default `~/.aint/improve-reports/<YYYY-MM-DD>.md` — outside any
   git repo, since incidents can quote sessions from unrelated repos). If
   no incidents were confirmed this run, skip writing and print
   `"No new incidents since last run."` to stdout instead.
9. Print a one-line summary to stdout: candidates found, incidents
   confirmed, analysis failures, report path (if written).

## Heuristics (candidate prefilter)

A small, extensible table in `heuristics.go`, each entry a compiled regex
tagged with a category name. Purely a cheap prefilter — it only decides
what's worth spending a `claude` call on; `claude` makes the real
is-this-actually-an-incident call in step 6.

| Category | Examples matched |
|---|---|
| `correction` | "that's wrong", "not what I asked/meant", "no, ...", "you misunderstood" |
| `stop-undo` | "stop doing that", "don't do that", "revert/undo that", "why did you" |
| `frustration-language` | "ugh", "seriously?", "come on", "how many times", "I already told you" |
| `shouting` | a run of ALL-CAPS words, or repeated `!!`/`??` |
| `terse-negative-reply` | short (<15 char) purely negative replies: "no.", "wrong.", "nope." |

## Claude invocation

`analyze.go` defines:

```go
type ClaudeRunner interface {
    Run(ctx context.Context, prompt string) (string, error)
}
```

The real implementation shells out to `claude -p <prompt>` (binary name
overridable via `--claude-bin`, default `"claude"`) with a bounded timeout
(60s) via `exec.CommandContext`. Tests inject a fake `ClaudeRunner`
returning canned JSON — no real subprocess or network call in the test
suite, consistent with the rest of `aint`.

Prompt asks for exactly this JSON shape back (parsed leniently — the first
`{...}` block found in stdout, in case the CLI wraps it in commentary):

```json
{
  "is_incident": true,
  "summary": "what happened, in the user's terms",
  "root_cause": "why the user reacted negatively",
  "aint_rule_suggestion": "concrete check idea (id, pattern, lang, severity) or null",
  "lint_rule_suggestion": "a language-linter rule if aint doesn't cover this domain, or null",
  "doc_memory_suggestion": "concrete CLAUDE.md/memory/lessons-learned text, or null"
}
```

If `claude` is not found on `PATH` at all (checked once via
`exec.LookPath` before the candidate loop), every candidate this run
becomes an `AnalysisFailed` entry immediately — no point attempting 50
subprocess calls that will all fail the same way — and the report's
summary line calls this out explicitly.

## Report format

```markdown
# aint improve report — 2026-08-03

3 incidents across 2 sessions, 1 analysis failure.

## /Users/joannisorlandos/git/wendy/wendyos

### Session 054b65f9 — 2026-08-01T03:52:14Z

**Signals:** correction, frustration-language
**User said:**
> "no that's not it, you're editing the wrong file again"

**What happened:** ...
**Root cause:** ...
**Suggested aint rule:** ... (or "Not applicable: ...")
**Suggested doc/memory change:** ... (or "Not applicable: ...")

---
```

Analysis-failed entries render with the quote + signals but a single
`⚠️ claude analysis unavailable: <error>` line instead of the four
suggestion fields.

## CLI

```
aint improve                        # scan ~/.claude/projects, write a report if anything new
aint improve --dir <path>           # scan a different session directory
aint improve --out <path>           # write the report somewhere specific (e.g. into a repo, deliberately)
aint improve --state <path>         # use a different offset-cache file
aint improve --claude-bin <name>    # override the claude CLI binary/path (default "claude")
aint improve --limit <n>            # max claude calls this run (default 50)
aint improve --full                 # ignore the cache, rescan everything from line 0
```

## Error handling

- `--dir` missing/unreadable → clear error to stderr, exit 1.
- Malformed JSON on a line (common on the trailing line of a session still
  being written) → skip that line, keep going.
- `claude` not on `PATH`, or a per-candidate call errors/times out/returns
  unparsable output → that candidate becomes an `AnalysisFailed` incident
  rather than aborting the run or getting silently dropped.
- State file missing → treated as an empty cache (first-ever run scans
  everything from line 0, bounded by `--limit`).
- State file corrupt (unparsable JSON) → treated as empty cache and
  overwritten on save, with a warning to stderr (rather than failing the
  run over a cache-only file).

## Testing

- **`transcript_test.go`**: fixture `.jsonl` files under `testdata/`
  covering human text lines, tool-result-shaped `user` lines (must be
  excluded), assistant text/tool_use lines, and unrelated `type`s (must be
  skipped without error) — including a truncated/malformed trailing line.
- **`heuristics_test.go`**: table test per signal category, true-positive
  and true-negative strings.
- **`state_test.go`**: load-when-missing, load-when-corrupt, save/reload
  round-trip, partial-offset-on-limit behavior.
- **`context_test.go`**: builds the expected truncated summary from a
  fixture sequence of assistant/tool turns.
- **`analyze_test.go`**: fake `ClaudeRunner` returning valid JSON,
  JSON wrapped in commentary, malformed JSON, and a `Run` error — asserts
  the right `Incident`/`AnalysisFailed` shape for each.
- **`report_test.go`**: renders a fixed `[]Incident` and compares against a
  golden markdown fixture.
- **`cmd/aint/improve_test.go`**: end-to-end `runImproveWithIO` against a
  temp `--dir`/`--state`/`--out` and an injected fake `ClaudeRunner`,
  covering: first run (cold cache), second run (nothing new), and a run
  that hits `--limit` partway through a file (offset lands before the
  first unprocessed candidate).

No real subprocess execution or network access anywhere in the test suite.
