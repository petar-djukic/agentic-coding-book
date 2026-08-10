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

**Work in progress, written in the open.** Part I is drafted. The
remaining parts are outlines. Chapters also run as articles at
[Mesh Intelligence](https://meshintelligence.substack.com?utm_source=github&utm_campaign=agentic-coding-book). Star or watch to follow the drafts.

## Contents

**[Introduction](01-introduction.md)**

### [Part I — The Skill](02-part-i-what-is-clauding.md)

| Chapter | State | Subject |
|---|---|---|
| [What Is Clauding?](03-what-is-clauding.md) | drafted | Behavioral vs constructional intent; in-the-loop vs on-the-loop; the six autonomy levels; the verification stack |
| [Layered Construction](04-layered-construction.md) | drafted | Inner and outer loops; tracer bullets; verification gates; why agents automate only one loop |
| [How the Machine Works](05-how-the-machine-works.md) | drafted | Next-token prediction; tokens and context budgets; the interpolation model; the agentic loop; four failure modes |
| [Language Selection](06-language-selection.md) | drafted | Language-model pairing; interpolation density; grammar complexity; the compiler in the inner loop |

### Parts II–VI

| Part | State | Subject |
|---|---|---|
| [II — Requirements](07-part-ii-requirements.md) | outline | Externalizing intent so agents execute it without guessing |
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

## Building the PDF

Requires [mage](https://magefile.org/), pandoc, xelatex, and plantuml
for figures. The PDF renders through the vendored
[Eisvogel](https://github.com/Wandmalfarbe/pandoc-latex-template)
template.

```bash
mage all      # figures + PDF into generated-files/
mage clean    # remove generated artifacts
```

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
