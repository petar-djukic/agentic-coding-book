<!-- chapter: C1.4 -->

# How a Harness Touches Your Code

## Learning Objectives

After reading this chapter, the reader will be able to:

1. Trace the sequence from a request to a changed file, naming every component that touches the code.
2. Explain why a specification must be self-contained, deriving it from how the harness assembles context rather than accepting it as a rule.
3. Describe the generate-verify-feed-back cycle as the harness's heartbeat, and say what happens when the verifier is absent.
4. Extend the skeleton from the first chapter with edit application behind policy and a verification gate whose output re-enters context.

## 4.1 The Model Never Opens a File

Watch a coding agent work and the natural description is that the model edited the file. It read the code, it found the bug, it wrote the fix. That description is wrong in a way that matters, and the way it is wrong determines what a programmer can do about the output.

A language model is a function from text to text: tokens in, tokens out. Nothing in that definition involves a file handle or a process, and a model can no more open `uploader.go` than a calculator can. What the model produces, when it appears to edit a file, is text that describes an edit — structured according to a schema the harness supplied, and meaningless until something else acts on it.

The something else is the harness. It read the file. It placed the contents in the request. It received the model's text, parsed the structured call out of it, checked the call against policy, and wrote the result to disk. Every operation that touched the codebase was performed by ordinary software running on the programmer's machine or in their CI environment [@latentpatterns-harness].

Which of those two accounts a programmer believes decides where they look when an edit comes back wrong. If the model edited the file, a bad edit is a model problem and the response is a better prompt. If the harness assembled a request, sent it, and applied what came back, then a bad edit has four or five places it could have originated — and most of them are inspectable.

Figure 4.1 traces one request all the way through.

**Figure 4.1** One request, from a sentence the programmer types to a changed file on disk.

![](figures/fig-4-1-harness-sequence.png)

*The model appears once, in the middle, and touches nothing. Every arrow that reaches the file system or the compiler leaves the harness, not the model. The dashed return from the compiler is the step that closes the loop: verification output re-enters the context as the next input, and the cycle repeats until the checks pass or the harness stops it.*

The rest of this chapter walks that sequence one stage at a time. The stages are ordinary — read, request, apply, verify — and each one is a place where the harness makes a decision the model never sees.

## 4.2 Assembling the Context

The model's entire world is the request. It has no memory of the last request, no view of the repository, and no way to ask for anything it was not given. So the first thing the harness does, before any generation happens, is decide what the model will know.

That decision is mechanical and it is lossy. A few files go in; the rest of the repository does not exist for this request. Older turns get dropped to make room. The harness adds tool definitions and a system prompt, then packs everything into a fixed budget. The context window is not a view onto the codebase; it is a copy of a chosen subset of it, frozen at the moment of the request.

> **Definition: Context assembly** — the harness's selection of what enters the model's context window for a given request: which files, how much conversation history, which tool definitions, and in what order.

Two properties of that copy shape everything downstream.

**Reading is whole-file.** When a harness includes a source file, it includes the file. There is no mechanism by which the model reads the first forty lines, decides it needs a function defined at line 300, and fetches it. Selective reading exists in some harnesses — a search tool, a symbol index — but each of those is a separate round trip: the model must emit a call, the harness must execute it, and the result must come back as a new turn. Within a single request, what was included is what exists.

**Position matters.** Attention is strongest at the beginning and end of the context and weakest in the middle [@liu2023]. A constraint placed in the middle of a long request is present in a way a byte counter would confirm and absent in a way the output reveals. The next chapter examines why; here the consequence is enough, and the consequence is that inclusion is necessary but not sufficient.

Put those together and a consequence follows that has nothing to do with style. A specification that says *see the architecture document for the module boundaries* has told the model nothing. Nothing in the loop fetches that document — there is no resolver, no link following, no lazy load. The sentence arrives as text, the reference registers as a fact about the world, and the model interpolates the module boundaries from the average of every module boundary in its training data.

The specification is not the file. The specification is what landed in the window.

> **Common Error:** Writing a specification that points at another document. Cross-references work for human readers, who can open the other document; they resolve to nothing for an agent, because nothing in the harness follows them. The agent does not report the gap — it fills it, and the filled version is plausible.

This is the mechanism behind an instruction that appears throughout the rest of this book: specifications given to agents are self-contained. Part II treats that as a working requirement and shows what a self-contained specification looks like at length. The reason is here, and it is not a matter of preference. A cross-reference is an instruction to fetch, and there is no fetcher.

## 4.3 Applying the Edit

The model returns text. Some of it is prose the programmer sees. Some of it is a structured tool call: a name and a set of arguments, formatted according to a schema the harness put in the request.

That call is a proposal. Between the proposal and the file system sits the part of the harness that most affects what an agent can do to a codebase.

The harness parses the call, which can fail — the model produces text, and text can be malformed. It validates the arguments against the schema. It checks the call against whatever policy it carries: which paths are writable, which commands may run, whether this action needs a human to approve it first. Only then does it perform the write.

Each of those checks is a decision the model has no part in. A harness that refuses to write outside the repository is not persuading the model to stay inside; it is declining to execute a call. This is why removing a tool is different in kind from instructing the model not to use it. The instruction competes with everything else in the context. The removal is arithmetic.

The write itself is unremarkable — the file changes on disk, the same as any other program writing a file. The model, at this point, has finished. It produced text and its turn ended. It has no knowledge of whether the write succeeded, and it will not know until something tells it.

## 4.4 The Heartbeat

A programmer editing code compiles it. The compile is not a separate activity from the editing; it is how the editing gets corrected. Agent-based development inherits that cycle, with one difference: the correction has to be delivered as text, because text is the only thing the model consumes.

After the write, the harness runs a verifier. Which verifier depends on the project — a compiler, a linter, a test suite, a type checker, several in sequence. The verifier produces output. That output is appended to the context, and the harness issues another request.

> **Definition: Agentic loop** — the cycle by which a model takes actions: the model emits a response containing structured tool calls, the harness executes them, the results are appended to context, and the model generates the next response conditioned on the updated context. The cycle repeats until the model signals completion or the harness stops it.

The second request is not a retry. It is a different request, because its context now contains something the first one lacked: evidence. `undefined: retryWithBackoff` is a fact about the world that the model had no access to when it wrote the call, and now does. The output improves because the input did.

This is the loop's whole mechanism, and it is worth stating flatly. **The harness converts consequences into context.** Everything an agent appears to learn during a session is the harness taking the result of an action and putting it where the model can read it.

A coding agent session is that cycle repeated: read a file, propose an edit, write it, compile, read the errors, propose a fix, write it, run the tests, read the failures, propose another fix. Dozens of iterations for a small change; hundreds for a large one. In the go-unix-utils generation runs, the loop count per task is the most informative number in the logs — a task that converges in six turns and a task that grinds through forty are not different in difficulty so much as in how well the verifier's output told the model what was wrong.

> **Performance Observation:** The verifier sits inside the loop, so its latency multiplies. A test suite that takes two minutes and a task that needs twenty iterations spend forty minutes in verification alone, before any model latency or token cost. Projects that move a slow end-to-end suite out of the inner loop and leave a fast compile-and-unit-test gate in it are not lowering their standards; they are changing which check runs at loop frequency and which runs at the end.

## 4.5 What the Loop Cannot Notice

Take the verifier away and the loop still runs.

The harness reads the file, assembles context, sends the request, receives a tool call, applies it, and reports completion. Every stage executes. Nothing errors. The agent produces a diff, a summary of what it did, and a confident account of why the change is correct.

What is missing is the only step that could have contradicted it. Generated code fails in ways that reproduce reliably: an API used with the wrong signature, a data model that does not match the schema, error handling for a condition that cannot occur alongside no handling for one that can [@wang2024]. None of those announce themselves in the diff. They look like code, because they are code, produced by a process whose output is fluent whether or not the interpolation behind it was sound.

The verifier is what turns a proposal into evidence. Without it the loop is a text generator with write access.

This is the fact that the verification half of this book rests on. Parts III and IV are long because this step is load-bearing and because building it well is harder than it looks — a test suite generated from the same misunderstanding as the code will pass. The mechanism is here: the loop has exactly one channel through which reality reaches the model, and if nothing is connected to it, the loop cannot tell a working change from a broken one.

## 4.6 Build: The Write and the Gate

The runtime built in the first chapter executes whatever the model proposes and checks nothing afterwards — its `verifying` state emits `pass` unconditionally. Those are the two gaps this chapter traced, and this section closes both: a write that sits behind policy, and a gate whose verdict re-enters the transcript and is routed by the profile. The runtime gains one field for the work — the repository root it is given at construction, shared with the tools it builds from the profile's declarations — and the profile gains two lines: a tool declaration, and a transition for a verdict the gate has yet to produce:

```yaml
transitions:
  # ... the four from Listing 1.1, plus:
  - {from: verifying, on: fail, to: deciding}
tools:
  - name: read_file
    root: "."
  - name: write_file
    root: "."
```

Listing 4.1 writes out the stage section 4.3 described.

**Listing 4.1** Edit application: the policy check runs before the write, and a refusal is a result like any other.

```go
type WriteFile struct{ root string }

func (t WriteFile) Name() string { return "write_file" }

func (t WriteFile) Run(args map[string]string) (string, error) {
	path := filepath.Clean(args["path"])
	if filepath.IsAbs(path) || path == ".." || strings.HasPrefix(path, "../") {
		return "", fmt.Errorf("write refused: %s is outside the root", args["path"])
	}
	full := filepath.Join(t.root, path)
	if err := os.WriteFile(full, []byte(args["content"]), 0o644); err != nil {
		return "", err
	}
	return "wrote " + path, nil
}
```

*The refusal happens in the harness's code path, not in the model's context. A call that fails the check never reaches the file system, whatever the request contained.*

The check is four lines, and its placement is the point. Section 4.3 called removing a tool arithmetic, against an instruction that competes with everything else in the context; the same arithmetic is available inside a tool. The path test does not compete with anything, because the model never sees it — the model sees the refusal string afterwards, appended to the transcript as one more consequence. Note what the profile did and did not do: declaring `write_file` granted the capability, and the check lives in the tool's Go, in the runtime's library. The read tool needs the same test; the write carries it here because the write is what this chapter traced.

Listing 4.2 asks the compiler.

**Listing 4.2** The gate: the compiler's opinion, captured as evidence and as a verdict.

```go
func (r *Runtime) verify() (string, bool) {
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = r.root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "build failed:\n" + string(out), false
	}
	return "build ok", true
}
```

*The model reads the string; the machine routes on the boolean, which is why one call returns both. The compiler's output passes through unedited, so the next request sees exactly what the build printed.*

Listing 4.3 spends both return values in the state that was falling through.

**Listing 4.3** The state that routes it: the verdict enters the transcript, and the profile decides where the machine goes next.

```go
		case "verifying":
			out, ok := r.verify()
			r.transcript = append(r.transcript, out)
			if ok {
				r.next("pass")
			} else {
				r.next("fail")
			}
```

*One append is the whole mechanism: the compiler's verdict lands in the transcript, and the next `Decide` is conditioned on it. Both verdicts route to `deciding` — the difference travels in the transcript — and rerouting `fail` somewhere stricter is an edit to the profile above rather than to this case.*

The case is the heartbeat of section 4.4, reduced to its mechanism: the harness converts consequences into context by appending them. Section 4.5's verifier-absent loop is now reachable from working code: deleting the `r.verify()` call leaves a loop that runs every stage, reports completion, and cannot tell a working change from a broken one. Part IV returns to the wiring the profile has made visible, when there is more than one gate to route. The skeleton still forgets — the transcript dies with the process. The chapter that closes this part takes that up, in prose and in its Build section.

## Summary

A language model produces text and touches nothing. Every operation that reaches a codebase — reading a file, writing an edit, running a compiler — is performed by the harness, which assembles a request, sends it, applies what comes back, and verifies the result. Context assembly is lossy and positional: the harness copies a chosen subset of the repository into a fixed window, and what is not in the window does not exist for that request, which is why a specification that cross-references another document conveys nothing an agent can act on. Between the model's proposal and the file system sits validation and policy, where a removed tool is a different kind of constraint than an instruction not to use one. After the write comes the verifier, whose output re-enters the context as the next input; that cycle is the loop's heartbeat, and it is the mechanism by which consequences become context. Remove the verifier and every stage still executes, producing confident output that nothing has checked — which is the reason verification occupies a third of this book. The Build section adds both stages to the running runtime: a write tool declared in the profile whose path check runs before the disk is touched, and a verify step whose compiler verdict re-enters the transcript as a signal the profile's own transitions route.

## Key Terms

| Term | Definition |
|---|---|
| **Agent harness** | The software wrapped around a language model: tool execution, file-system access, loop control, and policy. The harness runs the loop; the model predicts within it |
| **Agentic loop** | The cycle of model generation, tool execution by the harness, result injection into context, and next-step prediction that enables a text predictor to take actions |
| **Context assembly** | The harness's selection of what enters the context window for a given request: which files, how much history, which tool definitions, in what order |
| **Context window** | The maximum number of tokens a model can process and generate in a single request, covering input, output, and reasoning alike |
| **Tool** | A capability the harness exposes to the model as a callable interface. The model emits a structured call; the harness executes it. Tools are the only way the model affects anything |
| **Verification gate** | The set of automated checks that must pass before work proceeds. Inside the loop, the gate's output is the evidence the next request is conditioned on |
