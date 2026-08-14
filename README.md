# Agentic Coding

**Build a coding agent — and build software with it.** A book on
agentic coding, spec-driven development, and AI-assisted software
engineering. The reader builds a coding agent across the book. It
starts as a minimal declarative runtime that interprets one YAML
profile, grows part by part into a harness composed of several
agents, and becomes an orchestrator by Part V. The practice
taught along the way — writing requirements a coding agent can
execute, verifying generated code you did not write, orchestrating
agent loops past the limits of a single context window — is what the
agent under construction demands of its builder.

The material comes from production runs rather than demos — 45,789 lines
of production Go at peak across 60 pipeline runs, and 107 Unix utilities
regenerated from specification and verified against the GNU reference
binaries by differential testing. Those figures name a snapshot of the
datasets rather than a running total; `outline.yaml` records which.

**Work in progress, written in the open.** Part I is drafted
mechanism-first — the harness is the protagonist. All six chapters are
written. Every remaining chapter is specified rather than merely outlined:
all 39 carry a drafting contract under `docs/srd/`, so what a chapter must
establish, cite, and answer is fixed before it is written. Chapters also run as articles at
[Mesh Intelligence](https://meshintelligence.substack.com?utm_source=github&utm_campaign=agentic-coding-book). Star or watch to follow the drafts.

## Contents

**[Introduction](01-introduction.md)**

### [Part I — Agents and Harnesses](02-part-i-agents-and-harnesses.md)

The LLM never touches the file system; the harness does. Part I is
mechanism, so that every prescription in Parts II–VI has a fact behind it.

| Chapter | State | Subject |
|---|---|---|
| [What Is an Agent?](03-what-is-an-agent.md) | drafted | A state machine plus tools; the declarative agent — profile as data, runtime as interpreter — as the book's working model |
| [The Agents You Already Use](04-the-agents-you-already-use.md) | drafted | Claude Code, OpenCode, and Crush inspected directly against that model; packaging vs capability |
| [What Is a Harness?](05-what-is-a-harness.md) | drafted | The software around the model: tool execution, file-system access, loop control, policy |
| [How a Harness Touches Your Code](06-how-a-harness-touches-your-code.md) | drafted | The harness ↔ file system ↔ compiler sequence; generate → verify → feed-back as the heartbeat |
| [How the LLM Works, and Fails](07-how-the-llm-works-and-fails.md) | drafted | Next-token prediction; tokens and context budgets; the interpolation model; four failure modes |
| [Externalizing Memory](08-externalizing-memory.md) | drafted | Carrying state across sessions; in-the-loop vs on-the-loop as wiring; autonomy levels |

### [Part II — Construction and Requirements](09-part-ii-construction-and-requirements.md)

| Chapter | State | Subject |
|---|---|---|
| [Layered Construction](10-layered-construction.md) | drafted | Inner and outer loops; tracer bullets; verification gates; why agents automate only one loop |
| [Language Selection](11-language-selection.md) | drafted | Language-model pairing; interpolation density; grammar complexity; the compiler in the inner loop |
| Where requirements come from · Writing specs for agents · The constitution · Spec defects · Task decomposition | specified | Externalizing intent so agents execute it without guessing |

### Parts III–VI

| Part | State | Subject |
|---|---|---|
| [III — Testing](12-part-iii-testing.md) | specified | Verifying code no human wrote; differential testing; the generated-test circularity problem |
| [IV — Correctness](13-part-iv-correctness.md) | specified | What "correct" means when you did not write the code |
| [V — Orchestration](14-part-v-orchestration.md) | specified | Planners and generators; the blackboard — GitHub, beads, or build your own; the sub-agent lifecycle; failure and recovery |
| [VI — Instrumentation](15-part-vi-instrumentation.md) | specified | Logging, cost analysis, failure diagnosis |

One drafted chapter is parked outside the sequence:
[Intent, Autonomy, and Verification](unplaced-intent-autonomy-and-verification.md)
opened the previous Part I and has no slot in the current arc. Three
concepts in it are still needed — behavioral versus constructional
intent, the autonomy levels, and the verification stack — and
`docs/road-map.yaml` tracks where they land.

`outline.yaml` holds the full chapter-level plan and the snapshot policy for
every figure the book states; `docs/definitions.yaml` holds the term
definitions, each naming the chapter that introduces it.

## The evidence base

Claims trace to public repositories:
[go-unix-utils](https://github.com/petar-djukic/go-unix-utils) (107 Unix
utilities generated from specification, verified against GNU binaries),
[cobbler-scaffold](https://github.com/petar-djukic/cobbler-scaffold)
(the specifications and constitutions governing generation), and
[coding-skills](https://github.com/petar-djukic/coding-skills) (the
four-command GitHub loop Part V describes). A fourth repository plays a
different role:
[declarative-agents](https://github.com/Nokia-Bell-Labs/declarative-agents)
is the reference implementation for the build thread — the
profile-driven runtime, the agent catalog, and a planner–executor–critic
coding agent composed from it. The reader builds their own minimal
runtime rather than installing it; the framework is where they check
their work against a grown version of the same design.

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
for figures. Diagrams are PlantUML sources under `figures/`; `mage figures`
renders them to PNGs, which are build artifacts rather than committed files. The PDF renders through the vendored
[Eisvogel](https://github.com/Wandmalfarbe/pandoc-latex-template)
template.

```bash
mage all      # figures + PDF into generated-files/
mage outline  # outline PDF from docs/srd/ into generated-files/
mage audit    # specification consistency, prose conformance, build
mage clean    # remove generated artifacts
```

`mage outline` renders the book's structure from the specification rather
than from the chapters: it reads `docs/ARCHITECTURE.yaml` for the part and
chapter order, `docs/road-map.yaml` for per-chapter status, and every
`docs/srd/*.yaml` for that chapter's goal, objective, subgoals, content
spine, figures, and gaps. It also renders what the reader builds in each part
and collects every contract's gaps into one background-needed appendix,
marking those that hold a release. Chapters with no SRD are listed with their
status, so the outline doubles as a coverage report — there are none left. It
needs at least one SRD and fails with a message naming the directory
otherwise.

`mage audit` checks three things and reports every finding together, each
naming the rule or document field it comes from. **Specification consistency**:
every constitution rule an SRD cites exists, every derivation-chain link it
claims is owned by that chapter, every subgoal hangs off a book goal, every
citation and key term resolves, and chapter ids agree across `outline.yaml`,
`docs/ARCHITECTURE.yaml`, and `docs/road-map.yaml`. **Prose conformance**: the
mechanically checkable subset of `docs/constitutions/voice.yaml` — forbidden
terms, the chapter apparatus, sidebar labels, citations that resolve, figures
referenced from the prose. **Build integrity**: every referenced figure exists
and the book still compiles.

It runs no model and needs no credentials. Judgment the checks cannot make —
clarity, honesty, pedagogy, whether the argument holds — is the job of the
`review-chapter` skill, which the coding agent runs against a draft.

`docs/audit-baseline.yaml` records findings that are known and not yet fixed,
each naming the issue that clears it. Those are printed but do not fail the
audit; anything else does. An entry with no issue, or one that no longer
matches a finding, is itself a finding — so the file cannot decay into a
blanket exemption.

## Author

Petar Djukic — Principal AI Architect, 20+ years of production systems,
69 US patents, PhD in Computer Engineering. Designed
[Declarative Agents](https://github.com/Nokia-Bell-Labs/declarative-agents),
open-sourced by Nokia Bell Labs and the reference implementation for
this book's build thread. A companion volume,
[Agentic Applications](https://github.com/petar-djukic/agentic-applications-book),
covers building applications out of agents; this book owns the coding
agent, and the two cross-reference rather than repeat each other.

## License

MIT for the build machinery. Book text © 2026 Petar Djukic; quotation
with attribution welcome.
