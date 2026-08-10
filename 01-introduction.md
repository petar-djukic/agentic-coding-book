# Introduction

Ask a coding agent for a feature and you get code that compiles and a
structure you did not choose. The agent filled every gap in your request
with a guess, and the guesses are where projects die. This book is about
closing those gaps before generation starts: writing requirements an agent
can execute, building verification the agent cannot grade itself on, and
running the loop so the human stays in charge of what ships.

The material comes from a measured practice, not a framework pitch. The
numbers behind the chapters are public: 123 Unix utilities regenerated in
Go from specifications and verified against the GNU reference binaries by
differential testing; a $33.29 generation session accounted for token by
token; 320,000 generated lines deleted on purpose because the
specification, not the code, is the artifact worth keeping.

Each chapter runs as an article first. The live versions are at
[Mesh Intelligence](https://meshintelligence.substack.com?utm_source=github&utm_campaign=agentic-coding-book), and the chapters here consolidate them as they
stabilize. The book is written in the open; expect stubs, seams, and
revisions.

## The argument in one paragraph

Coding agents split the programmer's job. You still need to know how to
program — the agent's mistakes are invisible to someone who cannot read
the code — but the day-to-day work moves from writing code to specifying,
verifying, and orchestrating. The programmer who treats the agent as a
faster typist stays at the autonomy level of an autocomplete. The
programmer who externalizes intent into specifications and builds
verification around the agent scales past a single context window. The
levels between those two points structure this book.
