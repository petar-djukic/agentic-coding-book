# Introduction

Coding agents write more code in a day than most engineers write in a
month. More code is not the point. The point is that the relationship
between a programmer and their code has changed: you are no longer the
one typing, you are the one specifying, verifying, and orchestrating.

This book teaches experienced programmers how to work across the full
range of that shift — from a single interactive session to autonomous
pipelines running hundreds of tasks. The material draws on production
runs rather than demos: 44,628 lines of Go generated across 31 pipeline
runs, and 123 Unix utilities regenerated from specification and verified
against the GNU reference binaries by differential testing. Two open
repositories — [cobbler-scaffold](https://github.com/petar-djukic/cobbler-scaffold)
and [go-unix-utils](https://github.com/petar-djukic/go-unix-utils) —
serve as both the subject and the laboratory.

The argument underneath it: coding agents generate code, including the
code that makes them obsolete. The programmer who builds, instruments,
and improves their own generation pipeline compounds an advantage over
the one who waits for a vendor to ship one.

## What this book means by agentic coding

The vocabulary arrived faster than the practice, and four names now circulate
for things that differ in one respect: who checks the work.

**Vibe coding** came first. Andrej Karpathy coined it in February 2025 for a way
of working where you "fully give in to the vibes, embrace exponentials, and
forget that the code even exists" [@karpathy2025]. He meant it for throwaway
projects and later called the post a shower of thoughts; the phrase escaped
anyway and reached Merriam-Webster within weeks. Its defining property is that
nobody reads the output. That is a reasonable trade for a weekend prototype and
an unreasonable one for anything that has to keep working.

A 2025 review separates that from what it calls agentic coding: vibe coding as
"intuitive, human-in-the-loop interaction through prompt-based, conversational
workflows", against "autonomous software development through goal-driven agents
capable of planning, executing, testing, and iterating tasks with minimal human
intervention" [@sapkota2025]. The mechanical difference is the loop. An agent
runs its own verification and acts on what comes back, rather than emitting code
and stopping.

> **Definition: Agentic coding** — externalizing both behavioral and
> constructional intent into machine-readable artifacts, then generating code
> from those artifacts and verifying the output through automated gates instead of
> by manual review.

That definition leads with the artifacts and the gates, not with
autonomy, which makes it narrower than the term's common usage. The narrowing is
intentional. Autonomy is a consequence of having written enough down and built
enough checking; it is not a starting condition, and a system granted autonomy
without either produces confident output nobody has verified.

Read that way, the practice this book teaches is close to what the industry
calls **spec-driven development**: treating the specification as the contract an
agent generates from, not as documentation written afterwards
[@github-speckit]. The difference is emphasis, not substance. Spec-driven
development names the artifact; this book spends as much time on the gates,
because a specification with nothing checking the output against it is a wish.

At the far end sits the **dark factory**, borrowed from lights-out
manufacturing, where the plant runs with nobody on the floor. Applied to
software it means the whole cycle — plan, generate, test, review, ship — runs
with no human approving anything. The term circulates through vendor material
rather than literature, so it is better treated as a direction than as a
defined state.

These four are positions on the autonomy scale Part I ends with, not competing
schools. Vibe coding sits low on it, the dark factory at the top, and the
practice in this book occupies the middle, where the human has stopped reviewing
each change and has not stopped being accountable for what ships.

## Who this is for

Programmers who can already write production code. The book assumes you
can read the generated output and judge it — the agent's mistakes are
invisible to someone who cannot. It does not assume any prior experience
with agent orchestration.

## How the book is arranged

Part I is mechanism; the rest is practice derived from it. Each of the
later parts covers one thing the mechanism makes necessary.

| Part | Subject |
|---|---|
| I | What an agent is, what a harness is, how a harness touches your code, how the model fails, and where memory lives |
| II | Construction and requirements: layering construction, choosing the language, and externalizing intent so agents can execute it |
| III | Testing: verifying code no human wrote |
| IV | Correctness: what "correct" means when you did not write the code |
| V | Orchestration: the machinery that runs the loop across hundreds of tasks |
| VI | Instrumentation: observing what the agents actually did |

**Work in progress, written in the open.** Part I is drafted; the remaining
parts are outlines and notes. Chapters
also run as articles at
[Mesh Intelligence](https://meshintelligence.substack.com?utm_source=github&utm_campaign=agentic-coding-book). Expect stubs and seams.
