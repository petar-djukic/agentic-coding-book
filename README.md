# Agentic Coding

**How to build software with coding agents — and build the agents
themselves.** A book on agentic coding, spec-driven development, and
AI-assisted software engineering: writing requirements coding agents can
execute, verifying generated code you did not write, and orchestrating
agent loops with Claude Code, OpenCode, and models from Opus to GLM.

**Work in progress, written in the open.** Every chapter runs as an
article first at [Mesh Intelligence](https://meshintelligence.substack.com?utm_source=github&utm_campaign=agentic-coding-book); the chapters here consolidate
the live versions as they stabilize. Star or watch the repo to follow the
drafts.

## Contents

| Chapter | State | Live articles |
|---|---|---|
| [Introduction](01-introduction.md) | drafted | — |
| [What Is Agentic Coding?](02-what-is-agentic-coding.md) | stub | [Autonomy levels](https://meshintelligence.substack.com/p/what-level-of-autonomy-is-your-ai?utm_source=github&utm_campaign=agentic-coding-book), [Five coding agents](https://meshintelligence.substack.com/p/what-five-coding-agents-taught-me?utm_source=github&utm_campaign=agentic-coding-book), [The loop is the easy part](https://meshintelligence.substack.com/p/the-loop-is-the-easy-part?utm_source=github&utm_campaign=agentic-coding-book) |
| [Requirements](03-requirements.md) | stub | [Dude, where's my code?](https://meshintelligence.substack.com/p/dude-wheres-my-code?utm_source=github&utm_campaign=agentic-coding-book), [Architecture-first](https://meshintelligence.substack.com/p/the-architecture-first-approach?utm_source=github&utm_campaign=agentic-coding-book) |
| [Verification](04-verification.md) | stub | [Staying on track with Opus 5](https://meshintelligence.substack.com/p/how-to-stay-on-track-with-opus-5?utm_source=github&utm_campaign=agentic-coding-book), [The drinking bird test](https://meshintelligence.substack.com/p/the-drinking-bird-test?utm_source=github&utm_campaign=agentic-coding-book) |
| [Orchestration](05-orchestration.md) | stub | [Loop engineering](https://meshintelligence.substack.com/p/how-to-loop-engineering?utm_source=github&utm_campaign=agentic-coding-book), [GitHub as long-term memory](https://meshintelligence.substack.com/p/how-to-use-github-as-long-term-memory?utm_source=github&utm_campaign=agentic-coding-book), [Git worktrees](https://meshintelligence.substack.com/p/how-to-use-git-worktrees-with-coding?utm_source=github&utm_campaign=agentic-coding-book), [Three commands](https://meshintelligence.substack.com/p/three-commands-to-a-crude-orchestrator?utm_source=github&utm_campaign=agentic-coding-book), [GLM 5.2 on OpenCode](https://meshintelligence.substack.com/p/how-to-glm-52-on-opencode?utm_source=github&utm_campaign=agentic-coding-book) |
| [Instrumentation](06-instrumentation.md) | stub | [$33 of code generation](https://meshintelligence.substack.com/p/what-does-33-of-ai-code-generation?utm_source=github&utm_campaign=agentic-coding-book), [Black box until you add logging](https://meshintelligence.substack.com/p/your-ai-will-be-a-black-box-until?utm_source=github&utm_campaign=agentic-coding-book) |

## The evidence base

The claims trace to public repositories:
[go-unix-utils](https://github.com/petar-djukic/go-unix-utils) (123 Unix
utilities generated from specification, verified against GNU binaries by
differential testing),
[cobbler-scaffold](https://github.com/petar-djukic/cobbler-scaffold)
(the specs and constitutions that govern the generation), and
[coding-skills](https://github.com/petar-djukic/coding-skills) (the
four-command GitHub loop the orchestration chapter describes).

## Building the PDF

Requires [mage](https://magefile.org/), pandoc, xelatex, and plantuml
(for figures). The PDF renders through the
[Eisvogel](https://github.com/Wandmalfarbe/pandoc-latex-template) pandoc
template, vendored in `templates/`.

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
covers building applications from agents.

## License

MIT for the build machinery. Book text © 2026 Petar Djukic; quotation
with attribution welcome.
