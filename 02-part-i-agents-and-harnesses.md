# Part I — Agents and Harnesses

The model never touches your code. The harness does.

That sentence is why this part comes first. A coding agent is a state
machine paired with a set of tools, and the software wrapped around the
model — tool execution, file-system access, loop control, policy — is
what actually reads a file, applies an edit, and runs the compiler. The
model proposes text. Everything else is ordinary software someone wrote,
and it is the part you can change.

This part is mechanism, not practice. It establishes what an agent is,
what a harness is, how a harness manipulates a codebase, where the model
inside it fails, and where state lives between sessions. Nothing here
tells you how to work.

That comes later, and every piece of it rests on a fact established
here. Specifications must be self-contained because the harness reads
whole files into context and cannot follow a cross-reference.
Verification must be mechanical because the model cannot grade its own
output. Intent must be written down because memory has to live somewhere
the harness can reach.

A prescription in the rest of this book that cannot name the fact behind
it is a prescription the book has not earned.
