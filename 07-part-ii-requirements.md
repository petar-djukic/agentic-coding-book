# Part II — Requirements

*Outline stage — drafting notes are in `brainstorm/`.*

The spec is the source code now. Output quality is bounded by requirement quality, and the failure modes of an incomplete spec are specific and traceable: the agent fills every gap with a guess, and the guesses are consistent with each other while being inconsistent with the intent.

This part covers where requirements come from — standards, reference implementations, RFCs — how to write them so an agent executes without guessing, and what 320,000 generated-then-deleted lines settled about which artifact is worth keeping. It draws on the requirements documents used in go-unix-utils, where every utility is specified before any code is generated.

Most programmers have never written a formal spec, because they carried the requirements in their head while they coded. That does not work when something else is doing the coding.

Live material: [Dude, Where's My Code?](https://meshintelligence.substack.com/p/dude-wheres-my-code?utm_source=github&utm_campaign=agentic-coding-book) · [The Architecture-First Approach](https://meshintelligence.substack.com/p/the-architecture-first-approach?utm_source=github&utm_campaign=agentic-coding-book)
