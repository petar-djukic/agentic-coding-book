<!-- chapter: C1.3 -->

# What Is a Harness?

## Learning Objectives

After reading this chapter, the reader will be able to:

1. Define a harness and list its four responsibilities.
2. Given an agent session, identify which behavior came from the model and which came from the harness.
3. Explain why the harness, rather than the model, is the part a programmer shapes.

## 3.1 The Part You Can Change

An agent is asked to make a failing test pass, and it deletes the test.

The instinct that follows is to write a better prompt. Sterner wording, an explicit prohibition, capital letters. Sometimes that works well enough to end the incident. When it does not, the reason is usually that the prompt was never what decided the outcome. Something gave the agent a tool that could delete a file. Something let that tool run without asking anyone first. Neither of those was a choice the model made, and no amount of rewording reaches either one. Both were settings in the software that runs the agent, and that software has a name.

> **Definition: Agent harness** — the software wrapped around a language model: tool execution, file-system access, loop control, and policy. The harness runs the inner loop for a single task. The model never touches the file system; the harness does.

Coding with an agent means using a harness, whether or not the programmer thinks of it in those terms. Claude Code is a harness. So are Codex, Qwen Code, Gemini CLI, OpenCode, and OpenHands, and so is the agent mode inside an IDE. Each one wraps a model that, on its own, does nothing but turn text into more text.

The term is not this book's coinage. A 2026 study of coding agents defines the harness as "a middleware layer in between a developer and a large language model that orchestrates system prompts, tool execution, context management, and iterative reasoning loops" [@bensghaier2026]. The boundaries drawn in that sentence are drawn slightly differently from the four this chapter uses — context management appears there and policy does not — but the layer being pointed at is the same one, and the reason to name it is the same. Something sits between the programmer and the model, and it is doing most of the work.

Two facts about that layer give this book its subject. A programmer using an agent cannot change what the model would predict, because the weights are fixed and sealed behind an API. The harness is ordinary software, usually configurable and often open source. Every decision a programmer gets to make sits on the harness side.

## 3.2 Four Responsibilities

A harness is easier to reason about as four jobs than as one program, because a given behavior almost always traces back to exactly one of them. Figure 3.1 draws the four as the boundary they form around the model.

**Figure 3.1** The harness as a boundary around the model.

![](figures/fig-3-1-harness-boundary.png)

*The four responsibilities form the boundary. The model sits inside it and reaches nothing directly — the only things that leave it are text and structured tool calls. The programmer's configuration arrives from outside and never passes through the model at all.*

**Tool execution.** The harness publishes a set of callable capabilities, then runs them when the model asks. The model emits a structured call — a name and arguments, matching a schema the harness supplied — and the harness parses it, executes it, and returns the result as text. An agent that can run your test suite has a tool for it; an agent that cannot has no such tool, and no phrasing will conjure one.

**File-system access.** Reading a file into the request and writing an edit back out are both operations the harness performs. This is why an agent working in one directory has no idea what is in another. The harness read what it read.

**Loop control.** After a tool runs, the harness decides whether to send another request. It counts iterations, watches for a completion signal, notices when the agent is repeating itself, and enforces the limits that stop a session from running forever. An agent that gives up too early and one that grinds through forty turns on a small change are usually running the same model under different loop control.

**Policy.** Policy is where the harness says no. Which tools exist at all, which paths are reachable, which commands may run, which actions pause for human approval — these are decisions made in configuration, before the model sees anything [@tmforum-ig1218]. Deleting a test is a policy question, not a prompting question. A harness configured to require approval before deleting a tracked file does not persuade the model to behave; it declines to execute the call.

> **Common Error:** Reading a harness decision as a model decision, and reaching for the prompt to fix it. Practitioners do this systematically: a controlled study of harness releases found them reporting quality regressions after harness updates and attributing those regressions to the underlying model [@bensghaier2026]. The prompt is the one lever that cannot reach any of the four responsibilities, and it is the lever most often pulled.

## 3.3 What Changes When Only the Harness Changes

Attributing an agent's behavior to its harness is a claim that can be tested, by holding the model still and letting the harness move.

That experiment has been run. One controlled study fixed the underlying model and varied nothing but the harness, evaluating 35 sequential releases of a single coding CLI against 50 stratified SWE-bench Verified tasks. Quality moved across those releases, and the movement traced to specific pull requests in the harness [@bensghaier2026]. The same work surveyed five major open-source harnesses and found release velocities above two per day. The software around the model is changing faster than the model is, and it changes what the agent can do.

How far the effect reaches is less settled. A second study instrumented a planning agent layer by layer and found the declarative planning component carried the largest single gain — 24.1 percentage points of win rate, achieved with zero calls to the model — while the model-backed revision step fired on 4.3% of turns [@jung2026]. That agent plays Collaborative Battleship rather than writing software, so the numbers do not transfer to a coding task. What transfers is the method and its result in the abstract: once the layers of a harness are measured separately, the model's contribution can turn out to be the smaller one.

Both of those are 2026 preprints. The harness is a young enough object of study that the evidence for the claim this book is built on has not finished peer review, and a reader should weigh it accordingly. The mechanism is not in doubt; the next chapter traces the sequence and finds the harness performing every operation that reaches a codebase. What rests on early measurements is the size of the effect.

> **Good Practice:** Programmers who come up to speed on an unfamiliar harness tend to open its configuration before its documentation. A tool list and a permissions file are a precise inventory of what the agent can do in this repository, and they are usually shorter than the guide.

## 3.4 The Protagonist

This book is about the harness.

That is a claim about where a programmer's decisions have effect, and it follows from the split in §3.1. The model is the same model for everyone. Two teams pointing the same model at the same repository get different results, and the difference lives in the tools they exposed, the files they let it read, the loop they let it run, and the approvals they required.

Each of the four responsibilities is the subject of later parts. Specifications and decomposition shape what enters the context, which is file-system access and the assembly built on it. Testing and verification shape loop control, because the verifier's output is what the next iteration is conditioned on. Orchestration is loop control at a scale above the single task. Instrumentation is how any of it becomes visible. Each of those parts is a question about configuration, addressed at the length the answer takes.

The model is a component in that layer. It is an unusual component — the only one that cannot be inspected or patched — but the system around it is ordinary software, and it responds to engineering.

## Summary

A harness is the software wrapped around a language model: tool execution, file-system access, loop control, and policy. Coding with an agent means using one, and every agent product is a harness with a particular set of choices already made. Those four responsibilities account for most of what a programmer experiences as the agent's behavior, and each is decided in configuration rather than in the model. A controlled study that held one model fixed across 35 releases of a single coding CLI found quality moving with the harness alone, and practitioners in that study consistently mislaid the cause in the model. The model's weights are sealed and the harness is open, which is why the harness is this book's subject: it is the half that answers to engineering.

## Key Terms

| Term | Definition |
|---|---|
| **Agent harness** | The software wrapped around a language model: tool execution, file-system access, loop control, and policy. The harness runs the inner loop for a single task; the model never touches the file system |
| **Tool execution** | The harness's publication of callable capabilities and its running of them on request. The model emits a structured call; the harness parses, executes, and returns the result |
| **File-system access** | The harness's reading of files into the request and writing of edits back out. What the harness did not read does not exist for that request |
| **Loop control** | The harness's decision about whether to issue another request: iteration counts, completion signals, and the limits that end a session |
| **Policy** | The configured constraints the harness enforces before executing anything — which tools exist, which paths are reachable, which actions require approval |
| **Orchestrator** | The system managing the outer loop across many tasks: decomposition, ordering, state, and verification gates. Part V treats it; a harness runs one task, an orchestrator sequences many |
