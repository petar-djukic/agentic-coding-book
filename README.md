# Agentic Coding

**How to build software with coding agents — and build the agents
themselves.** A book on agentic coding, spec-driven development, and
AI-assisted software engineering: writing requirements a coding agent
can execute, verifying generated code you did not write, and
orchestrating agent loops past the limits of a single context window.

The material comes from production runs rather than demos — 44,628 lines
of Go generated across 31 pipeline runs, 123 Unix utilities regenerated
from specification and verified against the GNU reference binaries by
differential testing, and a $33.29 generation session accounted for
token by token.

**Work in progress, written in the open.** Part I is being re-cut
mechanism-first — the harness is the protagonist — so its four existing
drafts are being split and resequenced into six chapters. The remaining
parts are outlines. Chapters also run as articles at
[Mesh Intelligence](https://meshintelligence.substack.com?utm_source=github&utm_campaign=agentic-coding-book). Star or watch to follow the drafts.

## Contents

**[Introduction](01-introduction.md)**

### [Part I — Agents and Harnesses](02-part-i-what-is-clauding.md)

The LLM never touches the file system; the harness does. Part I is
mechanism, so that every prescription in Parts II–VI has a fact behind it.

| Chapter | State | Subject |
|---|---|---|
| What Is an Agent? | outline | A state machine plus tools; declarative agents as the lens where the skeleton is still visible |
| The Agents You Already Use | outline | OpenCode, Claude Code, LangGraph-flavored agents surveyed against that model |
| What Is a Harness? | outline | The software around the model: tool execution, file-system access, loop control, policy |
| How a Harness Touches Your Code | outline | The harness ↔ file system ↔ compiler sequence; generate → verify → feed-back as the heartbeat |
| [How the LLM Works, and Fails](05-how-the-machine-works.md) | partial | Next-token prediction; tokens and context budgets; the interpolation model; four failure modes |
| Externalizing Memory | outline | Carrying state across sessions, and the moment the human leaves the loop |

Chapters [What Is Clauding?](03-what-is-clauding.md),
[Layered Construction](04-layered-construction.md), and
[Language Selection](06-language-selection.md) are drafted against the
previous arc; the file moves and renames land with the resequencing.

### Parts II–VI

| Part | State | Subject |
|---|---|---|
| [II — Construction and Requirements](07-part-ii-requirements.md) | outline | Construction order and language choice, then externalizing intent so agents execute it without guessing |
| [III — Testing](08-part-iii-testing.md) | outline | Verifying code no human wrote; differential testing; the generated-test circularity problem |
| [IV — Correctness](09-part-iv-correctness.md) | outline | What "correct" means when you did not write the code |
| [V — Orchestration](10-part-v-orchestration.md) | outline | Agent roles, GitHub as coordination substrate, worktrees, failure and recovery |
| [VI — Instrumentation](11-part-vi-instrumentation.md) | outline | Logging, cost analysis, failure diagnosis |

`outline.yaml` holds the full chapter-level plan; `dictionary.yaml`
holds the term definitions used across the book.

## The evidence base

Claims trace to public repositories:
[go-unix-utils](https://github.com/petar-djukic/go-unix-utils) (123 Unix
utilities generated from specification, verified against GNU binaries),
[cobbler-scaffold](https://github.com/petar-djukic/cobbler-scaffold)
(the specifications and constitutions governing generation), and
[coding-skills](https://github.com/petar-djukic/coding-skills) (the
four-command GitHub loop Part V describes).

## Sources and research notes

Every claim in a drafted chapter cites its source; `references.yaml` is the
bibliography, and `mage all` resolves it.

The drafting notes and literature surveys behind the chapters are kept in a
private working tree rather than published here. Two reasons, both boring:
they contain third-party source texts that are not mine to redistribute, and
material from my day job that is not public. What survives review moves into
the chapters, with the source cited. If a claim here lacks a citation you can
follow, that is a defect — open an issue.

## Building the PDF

Requires [mage](https://magefile.org/), pandoc, xelatex, and plantuml
for figures. The PDF renders through the vendored
[Eisvogel](https://github.com/Wandmalfarbe/pandoc-latex-template)
template.

```bash
mage all      # figures + PDF into generated-files/
mage outline  # outline PDF from docs/srd/ into generated-files/
mage critic   # LLM critic over a drafted chapter (see below)
mage clean    # remove generated artifacts
```

`mage outline` renders the book's structure from the specification rather
than from the chapters: it reads `docs/ARCHITECTURE.yaml` for the part and
chapter order, `docs/road-map.yaml` for per-chapter status, and every
`docs/srd/*.yaml` for that chapter's goal, objective, subgoals, content
spine, figures, and gaps. Chapters with no SRD are listed with their status,
so the outline doubles as a coverage report. It needs at least one SRD and
fails with a message naming the directory otherwise.

`mage critic <chapter>` reads a drafted chapter, the binding rule sets under
`docs/constitutions/`, and the chapter's SRD when one can be paired, then asks
the model for findings: a constitution rule id, a line anchor, the text being
objected to, and what is wrong. Pass `all` to critique every drafted chapter.
It exits non-zero when any finding is blocking, and it never edits the prose —
revision is a separate step.

```bash
mage critic all                          # every drafted chapter
mage critic 05-how-the-machine-works.md  # one chapter
```

The model call is gated on `ANTHROPIC_API_KEY`. Without it the target reports
that it is skipping and exits zero, so a CI run with no credentials does not
fail the build. Chapters are paired to their SRD by title; when no SRD matches,
the chapter is checked against the constitutions alone and the mismatch is
reported. Pass `-srd` to `cmd/critic` to pair one explicitly.

## Author

Petar Djukic — Principal AI Architect, 20+ years of production systems,
69 US patents, PhD in Computer Engineering. Designed
[Declarative Agents](https://github.com/Nokia-Bell-Labs/declarative-agents),
open-sourced by Nokia Bell Labs. A companion volume,
[Agentic Applications](https://github.com/petar-djukic/agentic-applications-book),
covers building applications out of agents.

## License

MIT for the build machinery. Book text © 2026 Petar Djukic; quotation
with attribution welcome.
