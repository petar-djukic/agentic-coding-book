# Introduction

Coding agents write more code in a day than most engineers write in a
month. More code is not the point. The point is that the relationship
between a programmer and their code has changed: you are no longer the
one typing, you are the one specifying, verifying, and orchestrating.

This book teaches experienced programmers how to work across the full
range of that shift — from a single interactive session to autonomous
pipelines running hundreds of tasks. The material draws on production
runs rather than demos: 44,628 lines of Go generated across 31 pipeline
runs, 123 Unix utilities regenerated from specification and verified
against the GNU reference binaries by differential testing, and a
$33.29 generation session accounted for token by token. Two open
repositories — [cobbler-scaffold](https://github.com/petar-djukic/cobbler-scaffold)
and [go-unix-utils](https://github.com/petar-djukic/go-unix-utils) —
serve as both the subject and the laboratory.

The argument underneath it: coding agents generate code, including the
code that makes them obsolete. The programmer who builds, instruments,
and improves their own generation pipeline compounds an advantage over
the one who waits for a vendor to ship one.

## Who this is for

Programmers who can already write production code. The book assumes you
can read the generated output and judge it — the agent's mistakes are
invisible to someone who cannot. It does not assume any prior experience
with agent orchestration.

## How the book is arranged

Each part covers what is required to work at one level of autonomy
safely.

| Part | Subject |
|---|---|
| I | What an agent is, what a harness is, how a harness touches your code, how the model fails, and where memory lives |
| II | Construction and requirements: layering construction, choosing the language, and externalizing intent so agents can execute it |
| III | Testing: verifying code no human wrote |
| IV | Correctness: what "correct" means when you did not write the code |
| V | Orchestration: the machinery that runs the loop across hundreds of tasks |
| VI | Instrumentation: observing what the agents actually did |

**Work in progress, written in the open.** Part I is being re-cut
mechanism-first; the remaining parts are outlines and notes. Chapters
also run as articles at
[Mesh Intelligence](https://meshintelligence.substack.com?utm_source=github&utm_campaign=agentic-coding-book). Expect stubs and seams.

## A note on the term

Part I uses the term **clauding** for the practice this book teaches.
The word is a coinage, tied to the tool the author uses most, and it
predates the book's current title. Read it as a synonym for agentic
coding wherever it appears.
