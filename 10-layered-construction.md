<!-- chapter: C2.1 -->

# Layered Construction

## Learning Objectives

After reading this chapter, the reader will be able to:

1. Distinguish the inner loop (per-task: specify, generate, verify) from the outer loop (per-project: decompose, order, validate layers).
2. Explain why agents automate the inner loop but not the outer loop.
3. Given a system to build, identify tracer bullets — thin end-to-end slices — and define the increment order.
4. Identify when an increment is sound enough to build the next increment on top of it.
5. Recognize the failure mode of skipping the outer loop and diagnose it in an ongoing project.

## 7.1 The Two Loops

Software has never been built in one pass. A programmer does not sit down with a requirements document and type a finished system from top to bottom. Development is iterative — a spiral where each pass refines the previous one. This has been understood since Boehm formalized the spiral model in 1986 [@boehm1986], and the industry uses variants of it under different names: iterative development, the DevOps inner/outer loop, agile sprints.

The spiral has two concentric cycles.

> **Definition: Inner loop** — the tight cycle executed per task: specify a small piece, write it, test it, fix it, move to the next piece. In manual development, a programmer runs this loop many times per hour.

> **Definition: Outer loop** — the broader cycle that governs the project: gather requirements, make architectural decisions, decompose the system into layers, run the inner loop on each layer, validate that the layer is sound, then proceed to the next layer.

The inner loop is the one people think about. The outer loop is the one that determines whether the project succeeds.

**Figure 7.1** The inner and outer loops of software construction.

![](figures/fig-7-1-inner-outer-loops.png)

*The outer loop governs the project: requirements, architecture, decomposition into increments, and verification gates between them. The programmer drives this loop. Within each increment, the inner loop runs one or more tasks: specify, generate, verify. The agent drives this loop. Each increment must pass its verification gate before the next increment builds on top of it. The "fail" path in the inner loop feeds errors back to the agent for correction. The "pass gate" transitions between increments are decisions the programmer makes — or, at L4, decisions the orchestrator proposes and the programmer approves.*

## 7.2 What Agents Change

When an agent writes the code, the loops split apart — just as behavioral and constructional intent split apart (Chapter 5). The inner loop can be automated: give the agent a specification, let it generate code, run the tests, feed errors back. This is what most people mean by "using a coding agent." But the outer loop remains the programmer's responsibility. The agent does not know that the data model must be solid before the business logic layer can be trusted. It does not know that the API contract must be stable before integration tests are meaningful. It does not know the construction order.

The inner loop, automated, looks like this:

1. The programmer (or orchestrator) provides a specification for a single task.
2. The agent generates code from that specification.
3. Automated verification runs: compiler, linter, tests.
4. If verification fails, the agent receives the errors and tries again.
5. If verification passes, the task is complete.

This is what the agent harness manages — prompt construction, tool execution, policy checks, and loop control for one task at a time [@latentpatterns-harness].

The outer loop, which remains the programmer's job, looks like this:

1. **Plan**: Gather requirements. Make architectural decisions. Identify the next increment — the thinnest slice of functionality that works end-to-end.
2. **Build**: Run the inner loop on that increment. The agent generates code from the specification.
3. **Verify**: Run the full verification pipeline — not just tests, but compilation, linting, coverage, security scanning, and specification conformance. This is more than the inner loop's quick compile-and-test cycle. It is a deliberate, thorough assessment of the increment as a whole.
4. **Repair**: If verification finds defects, fix them. Repair is not the same as building — it operates on different context (error messages, failing tests, the original spec) and has different success criteria (previously-failing tests pass, no regressions introduced).
5. **Improve**: After several increments, assess the accumulated code for structural quality. Identify inconsistencies, duplication, and drift from the architecture. Propose refactoring tasks that feed back into step 1.
6. Repeat until the system is complete.

The outer loop is where the programmer's experience, judgment, and architectural knowledge live. It is the thing the agent cannot generate for itself.

> **From the Field:** The first time I ran an agent on a non-trivial project, I described the whole system and let it go. The code compiled. Some tests passed. But nothing actually worked end-to-end — the data model assumed one query pattern, the API assumed another, and the two did not meet. The second time, I built one feature all the way through: schema, query, endpoint, test. It was thin and incomplete, but it worked. Every subsequent feature built on a foundation I had already verified. The difference was not the agent. The difference was building end-to-end from the start.

## 7.3 Why You Cannot Skip the Outer Loop

The failure mode is treating agent-based development as "describe the whole system, let the agent build it." This skips the outer loop. It fails for the same reason that writing an entire system in one pass fails in manual development: parts of the system depend on other parts being correct, and defects in one area are invisible until something else tries to use them.

> **Common Error:** Skipping the outer loop. The programmer describes the whole system in a single prompt and expects the agent to produce it in one generation pass. The agent produces code that compiles but has never been tested end-to-end. The data model assumes one query pattern, the API assumes another, and the two do not meet. The programmer discovers this only after the entire system has been generated — at which point every piece may need rework.

The cost of a defect increases with the amount of code built on top of it. A wrong assumption is cheap to fix when only one feature depends on it. After ten features depend on it, the fix propagates through all of them. This is true in manual development. It is worse with agents because the agent generates more code faster — which means more code built on a bad assumption before anyone notices.

### 7.3.1 The One-Shot Illusion

Agents are good enough at small tasks that the one-shot approach feels like it works. Describe a function, get a function. Describe a class, get a class. The programmer extrapolates: describe a system, get a system.

The extrapolation fails because systems have internal dependencies that functions do not. A function is self-contained. A system has components that must agree with each other: the data model constrains the queries, the queries constrain the API, the API constrains the client. These constraints cross boundaries. When the agent generates everything at once, it resolves these constraints by guessing — and the guesses are consistent with each other but may not be consistent with the programmer's intent.

The result compiles. The tests that the agent wrote for its own guesses pass. The system is internally consistent but externally wrong. The programmer discovers this only when trying to use the system for its intended purpose — at which point the fix requires reworking multiple components.

## 7.4 Tracer Bullets

The discipline is incremental construction: build one working increment at a time, verify each increment end-to-end before building the next. Hunt and Thomas called this approach **tracer bullets** [@hunt1999] — fire a round, see where it hits, adjust, fire again. Each round is a thin slice of functionality that works all the way through the system, from input to output.

> **Definition: Tracer bullet** — the thinnest slice of functionality that works end-to-end through the system. A tracer bullet touches every component it needs — data model, business logic, API, tests — but implements only enough of each to make one feature work. It proves the architecture before the architecture is complete.

The tracer bullet approach is the opposite of horizontal layering. Horizontal layering says: build the entire data model, then the entire business logic layer, then the entire API. The problem is that the first integration test does not run until the third layer is complete. If the data model and the API do not agree — and with agent-generated code, they often do not — the programmer discovers it after building three complete layers.

Tracer bullets say: build one feature through all the components. The data model has one table. The business logic has one rule. The API has one endpoint. The test exercises the feature end-to-end. It is thin and incomplete, but it works. The next increment adds the second feature, also end-to-end. Each increment builds on a foundation that has already been verified to work as a system, not just as isolated components.

### 7.4.1 Why Tracer Bullets Work Better With Agents

Agents generate code fast. This is an advantage when the code is correct and a liability when it is not. Horizontal layering maximizes the liability: the agent generates an entire data model layer — hundreds of lines — before any of it is tested against the layers that will use it. If the assumptions are wrong, hundreds of lines need rework.

Tracer bullets limit the blast radius. Each increment is small. If the agent's assumptions are wrong, the programmer discovers it immediately — on the first end-to-end test — and the rework is bounded to one thin slice. The next increment benefits from the corrected assumptions.

This also addresses the context problem. An agent working on a tracer bullet has focused context: one feature, the components it touches, the test that proves it works. An agent working on an entire horizontal layer has broad, diffuse context: every table, every column, every relationship — most of which is not relevant to any individual decision the agent is making.

### 7.4.2 Increments Are Lumpy

Tracer bullets are not uniform slices. Real systems do not decompose into neat, equal-sized features. Some increments are heavy in one area and light in another. The first increment might require significant data model work and a trivial API endpoint. The second increment might reuse the existing data model and require significant API logic.

Some functionality exists to serve other functionality. A caching layer might be built as part of one increment but is really there because the next three increments need it. An authentication system might be built as scaffolding — enough to make the first feature work, then extended when a later feature requires finer-grained permissions.

This is normal. The point is not that every increment is the same size or touches every component equally. The point is that every increment works end-to-end. "End-to-end" means whatever it means for the system being built: for a web application, a request that hits the API and returns a correct response. For a CLI tool, a command that reads input and produces correct output. For a library, a test that imports the public API and exercises a real use case.

> **Good Practice:** Define what "end-to-end" means for the system before starting generation. Then identify the thinnest slice of functionality that exercises it. Build that first. Every subsequent increment adds functionality that also works end-to-end. Write the increment plan down — it is constructional intent, and if it remains unwritten, the agent will not decompose at all.

## 7.5 Verification Gates Between Increments

Each increment boundary is a verification gate — a set of conditions that must be true before the next increment begins.

> **Definition: Verification gate** — the set of automated checks that must pass after an increment before construction proceeds to the next. Gates prevent defects from accumulating across increments.

A verification gate is not just "the tests pass." It is a deliberate checkpoint where the programmer asks: does this increment work end-to-end?

The gate for a first increment of a web application might include:
- The schema migration runs without errors
- A request to the endpoint returns the correct response
- The end-to-end test passes against a real database
- Error cases return appropriate status codes

The gate for a later increment that adds authentication might include:
- Unauthenticated requests are rejected
- Authenticated requests to all existing endpoints still work
- The new authentication flow works end-to-end
- No regressions in previously passing tests

The specifics depend on the increment. The principle does not: every increment has a gate, and the gate confirms the system still works end-to-end.

### 7.5.1 What Happens When a Gate Fails

When a verification gate fails, the response is not to push through to the next increment. The response is to fix the current increment before proceeding. This is where the outer loop earns its value — a defect caught at a gate costs one increment of rework. A defect caught three increments later costs three increments of rework.

Repair is a distinct activity from building. The context is different: the agent receives the error message, the failing test output, the relevant code, and the original specification — not a fresh task description. The success criterion is different: repair succeeds when previously-failing tests pass AND previously-passing tests still pass. The risk is different: a fix that introduces a regression is worse than the original defect.

In practice, gate failures fall into two categories:

**Specification defects.** The increment was built correctly according to the spec, but the spec was wrong. The endpoint returns exactly what the spec said, but the spec described the wrong behavior. The fix is to update the specification and regenerate. This is a normal part of iterative development — the outer loop's purpose is to surface these defects early, when they are cheap.

**Generation defects.** The increment does not match the spec. The agent guessed wrong, misinterpreted a requirement, or introduced a structural decision that contradicts the specification. The fix is to improve the specification (make it less ambiguous) and regenerate. This is where the techniques in Part II (Requirements) apply.

### 7.5.2 Structural Quality Over Time

Hunt and Thomas describe **software entropy** — the tendency of code to decay over time as changes accumulate [@hunt1999]. In manual development, entropy is a gradual process. A programmer adds a feature, takes a shortcut, leaves a TODO. Over months and years, shortcuts accumulate and the codebase becomes harder to work with. Refactoring — restructuring code without changing behavior — is the normal countermeasure. No one expects the code they write in month one to be the right structure for month twelve. The codebase evolves, and periodic refactoring keeps entropy in check.

With agent-generated code, entropy is not gradual. It is immediate. Each increment is generated independently. The agent has no memory of the structural decisions it made in previous increments. It does not know that it already wrote a date parser in file A when it writes a different date parser in file B. It does not know that the error handling pattern in module C contradicts the pattern in module D. The result: structural decay that would take months to accumulate in manual development appears within days of agent-based generation.

Verification gates catch functional defects — code that does not do what it should. They do not catch this structural decay. After several increments, the accumulated code may have problems that no individual increment introduced: duplicated helper functions written independently by different tasks, inconsistent error handling patterns, three different ways to parse the same data format, abstractions that should exist but do not.

These are not bugs. The tests pass. The code works. But the codebase is harder to extend, harder to verify, and harder to specify future work against. Each new increment is more likely to introduce further inconsistencies because the agent interpolates from an increasingly inconsistent codebase — the entropy compounds.

Refactoring is not a remedial activity. It is a normal part of software development. No one — human or agent — can foresee the final structure of a system at the beginning. The first increment reveals constraints the second increment must accommodate. The third increment introduces a pattern that, in hindsight, the first two should have used. This is not a failure of planning. It is the nature of building something complex iteratively. The structure emerges as the system grows, and refactoring is how the code catches up to the programmer's evolving understanding of what the structure should be.

The response is periodic structural review — analyzing the accumulated code for patterns, inconsistencies, and refactoring opportunities. The refactoring proposals feed back into the planning step as new tasks: consolidate the three parsers, extract the common error handling, introduce the abstraction that three modules independently need. With manual development, the programmer does this naturally — noticing duplication during code review, refactoring as they go. With agent-based development, it must be deliberate, because the agent will not notice and will not refactor on its own. Part V covers how an orchestration pipeline automates this review.

## 7.6 Incremental Construction and Autonomy Levels

This maps directly to the autonomy levels defined in Part I, in the chapter on externalizing memory.

At L2, the programmer runs both loops manually — typing code in the inner loop, applying architectural judgment in the outer loop. The programmer decides the increment order, writes each increment, tests it end-to-end, and proceeds when satisfied.

At L3, the inner loop is automated: the agent generates from specifications and tests verify the output. The programmer drives the outer loop — identifying the next tracer bullet, specifying what end-to-end means for it, deciding when the increment is sound enough to build the next one on. The programmer writes the increment specifications. The agent builds each increment. The verification gates decide whether to proceed.

At L4, the system proposes its own outer loop: it reads architectural documentation, proposes an increment plan, and the programmer reviews the plan rather than creating it. The programmer's role shifts from designing the outer loop to approving or correcting a proposed sequence of increments.

Part V of this book covers the machinery that automates the full cycle — not just the inner loop, but all five steps: planning (decomposing the project into tasks), building (running the agent on each task), verifying (the full verification pipeline at increment boundaries), repairing (fixing defects found during verification), and improving (analyzing accumulated code for structural quality). The orchestrator manages the transitions between these steps, tracks state across increments, and decides when to proceed. The programmer's job is to define the increments, set the verification criteria, and make the judgment calls the orchestrator cannot.

## Summary

Software development has always been iterative: an inner loop (specify, write, test, fix) nested inside an outer loop that governs the project. The outer loop has five steps: plan the next increment, build it, verify it thoroughly, repair what verification finds, and periodically improve the structural quality of the accumulated code. When agents automate code generation, they automate the inner loop and parts of the outer loop — but the programmer remains responsible for the planning, the verification criteria, and the judgment calls. Skipping the outer loop by asking the agent to generate an entire system in one pass produces code that is internally consistent but externally wrong, with defects that propagate through every component built on bad assumptions. The discipline of incremental construction — one verified end-to-end slice at a time, with explicit gates between increments — is what makes agent-based development reliable at scale.

## Key Terms

| Term | Definition |
|---|---|
| **Inner loop** | The per-task development cycle: specify, write, test, fix. With agents, this loop can be automated |
| **Outer loop** | The per-project development cycle: requirements, architecture, increment planning, validation. Determines construction order and increment boundaries |
| **Tracer bullet** | The thinnest slice of functionality that works end-to-end through the system, proving the architecture before it is complete |
| **Verification gate** | The set of automated checks that must pass after an increment before construction proceeds to the next |
| **Agent harness** | The orchestration layer around a language model that manages prompts, tool execution, policy checks, and loop control for one task at a time |

