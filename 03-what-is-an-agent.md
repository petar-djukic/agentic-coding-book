<!-- chapter: C1.1 -->

# What Is an Agent?

## Learning Objectives

After reading this chapter, the reader will be able to:

1. Define an agent as a state machine paired with tools, and apply that definition to a system they have used.
2. Identify the loop, the state, and the tool boundary in any system presented to them as an agent.
3. Explain why the model alone is not an agent.
4. Implement the skeleton in Go: the states, the transitions, and a tool table.

## 1.1 A Word That Stopped Selecting

Autocomplete is sold as an agent. So is a chat window with a web-search button. So is a system that works through a backlog overnight and opens pull requests in the morning.

A word that covers all three lets a reader predict nothing. Told that a new tool is an agent, they cannot say whether it will run for a second or an hour, whether it can write to disk, or whether anyone will see what it did before it does it. The word has stopped narrowing the possibilities, which is the only work a category term does.

This is older than the coding tools. Wooldridge and Jennings surveyed the term in 1995 and found it already contested, dividing the usage they saw into a weak notion of agency — autonomy, social ability, reactivity, and pro-activeness — and a strong notion that added mentalistic attributes such as belief, intention, and knowledge [@wooldridge1995]. Thirty years later the same argument runs about the same word, now attached to different products.

Industry vocabulary has grown around the problem without solving it. Autonomy is graded on published scales, from fully manual operation up to systems that set their own goals [@tmforum-ig1218]. Those levels are useful for describing how much rope an operator has given a system. They do not say what the system is, so two products at the same level can share no structure at all.

A definition that answers the structural question turns out to be short.

> **Definition: Agent** — a state machine paired with a set of tools. The state machine decides what to do next; the tools are how it acts on anything outside itself. Coding agents differ from that skeleton in packaging rather than in kind.

The definition survives a change of vendor because it names parts rather than behaviors. Autonomy, initiative, and how impressive the output looks are all consequences of what states exist and which tools are wired to them.

## 1.2 The Three Parts

Applying that definition means finding three things in a running system.

> **Definition: State** — what the agent carries from one step to the next: the task it was given, what it has tried, and what came back. State is the reason step twelve differs from step one.

> **Definition: Transition** — the rule that moves the agent from one state to the next, given a signal. In an agent, most signals come from model output or from a tool's result.

> **Definition: Tool** — a capability the surrounding program exposes to the model as a callable interface: reading a file, running a command, searching a codebase. The model emits a structured call, and the program executes it and returns the result. Tools are the only way the model affects anything outside the process.

Figure 1.1 puts the three together in the smallest shape that still deserves the name.

**Figure 1.1** The agent skeleton.

![](figures/fig-1-1-agent-skeleton.png)

*Transitions are driven by model output and tool results. Executing a tool is the only state that reaches anything outside the process; every other transition is bookkeeping the loop does to itself. Verification failure returns to deciding, which is what makes the shape a loop rather than a pipeline.*

Everything in that diagram is ordinary. A queue, a switch statement, and a function that shells out would implement it. The unusual part is only that one of the inputs to the switch is a language model's output, and a model is a poor source of control flow because it is a good source of plausible text.

## 1.3 The Skeleton Is Usually Hidden

Most agents do not look like Figure 1.1, because their state machine was never written down as one.

A coding agent assembled the straightforward way is a `while` loop wrapped around a chain of conditionals: if the model asked to read, read; if it asked to write, write and then run the build; if it signalled completion, validate and either finish or continue. Those branches are states. The transitions are the conditions that select them, and the state is whatever the loop happened to keep in scope. Nothing is labelled, so recovering the machine means tracing every path by hand.

Declarative agents are the form where the skeleton stays visible. The states, the transitions, and the tool declarations live in a data file that a fixed runtime interprets, so changing the workflow is an edit to configuration rather than to code [@declarative-agents]. That property makes them useful here for a reason unrelated to whether anyone should build agents that way: the thing this chapter is trying to point at is written down in them, and can be read.

The companion volume, *Agentic Applications*, rebuilds published agent papers as declarative machines and is where the question of how to build one is treated [@agentic-applications]. This chapter borrows the lens and stops.

> **From the Field:** The skeleton became visible to me while rebuilding an agent whose behavior I could not predict. Lifting its loop into a profile — states named, transitions listed, tools declared — turned a debugging problem into a reading problem. The bug was a transition I had not known was there, sitting in an `else` branch, and it had been in the diagram I would have drawn all along, had I drawn one.

Once the shape is visible in one system, it is visible in the rest. The chapter on the agents you already use does exactly that, finding the same three parts in tools the reader has been running for months.

## 1.4 The Model Is Not the Agent

Given the definition, the model occupies one small position: it supplies the signal that selects a transition.

It does not hold the state, because it has no memory between requests. It does not execute the tools, because it has no way to reach a file. It does not run the loop, because it has no control over whether it is called again. All three of those belong to the program around it, which is ordinary software that somebody wrote and can change [@latentpatterns-harness]. Naming that program is the job of a later chapter in this part; recognizing that it exists is enough here.

> **Common Error:** Treating the model as the agent, and therefore reading every behavior as something the model chose. An agent that retries five times and gives up, that never looks at the test file, or that edits a path it should not have touched is displaying decisions made by the loop, the tool set, and the state. The model contributed one signal per step.

Calling a system an agent, then, is a claim about wiring. Whether it is a good agent depends on which states exist, what the tools can reach, and what happens on failure — all questions with inspectable answers, which is why the rest of this book can be about engineering rather than about prompting.

## 1.5 Build: The Skeleton in Go

A structural definition makes a claim that can be tested by construction: name the parts, and the thing can be built from them. This book runs that test on itself. Each part closes by building a piece of the program its chapters explained — a few dozen lines of Go at a time, starting here with the skeleton and ending, in Part V, with an orchestrator running many copies of it at once. The listings are pedagogical; the production evidence stays with the repositories the introduction names.

Listing 1.1 is Figure 1.1 as code.

**Listing 1.1** The agent skeleton: four states, a tool table, and one interface where the model plugs in.

```go
type State int

const (
	Deciding State = iota
	Executing
	Verifying
	Done
)

// Tool is the only path from the process to anything outside it.
type Tool interface {
	Name() string
	Run(args map[string]string) (string, error)
}

// Model supplies one signal per step: the next call, or done.
type Model interface {
	Decide(transcript []string) (call Call, done bool)
}

type Call struct {
	Tool string
	Args map[string]string
}

type Agent struct {
	state      State
	transcript []string
	pending    Call
	tools      map[string]Tool
	model      Model
}

func (a *Agent) Run(task string) []string {
	a.transcript = append(a.transcript, task)
	a.state = Deciding
	for a.state != Done {
		switch a.state {
		case Deciding:
			call, done := a.model.Decide(a.transcript)
			if done {
				a.state = Done
				break
			}
			a.pending = call
			a.state = Executing
		case Executing:
			a.transcript = append(a.transcript, a.execute(a.pending))
			a.state = Verifying
		case Verifying:
			// Nothing checks the work yet; a later chapter's Build
			// section wires the gate in here.
			a.state = Deciding
		}
	}
	return a.transcript
}

func (a *Agent) execute(c Call) string {
	tool, ok := a.tools[c.Tool]
	if !ok {
		return "no such tool: " + c.Tool
	}
	result, err := tool.Run(c.Args)
	if err != nil {
		return "error: " + err.Error()
	}
	return result
}
```

*The states are an enum, the transitions are the assignments to `a.state`, and the transcript is the state the machine carries between steps. The model appears once, behind an interface, supplying the signal that section 1.4 said was its whole contribution.*

Everything in the listing is ordinary Go. The loop is a `for`, the machine is a `switch`, and a failed tool lookup becomes a string in the transcript — which the model reads on its next turn, the same way it reads everything else. The one non-ordinary element is quarantined behind the `Model` interface, and the listing works identically whether the implementation calls a vendor API or a table of canned responses, which is what the definition predicted: the model is a component, not the agent.

Listing 1.2 supplies a tool and the table that holds it.

**Listing 1.2** One tool, and the table that is the tool boundary.

```go
type ReadFile struct{ root string }

func (t ReadFile) Name() string { return "read_file" }

func (t ReadFile) Run(args map[string]string) (string, error) {
	data, err := os.ReadFile(filepath.Join(t.root, args["path"]))
	return string(data), err
}

func NewAgent(m Model, root string) *Agent {
	return &Agent{
		model: m,
		tools: map[string]Tool{"read_file": ReadFile{root}},
	}
}
```

*The map in `NewAgent` is the complete answer to what this agent can touch. Deleting an entry removes the capability; nothing else has to change.*

Section 1.3's observation about declarative agents is sitting in this listing: the states, the transitions, and that table could be listed in a data file and interpreted by a fixed runtime, and the Go above is the same machine with the skeleton compiled in rather than read at startup. The skeleton cannot yet write a file, check its work, or remember having run. Each of those gaps is a later chapter's Build section.

## Summary

An agent is a state machine paired with a set of tools. The state machine decides what to do next; the tools are how it acts on anything outside itself. State is what the agent carries between steps, a transition is the rule that moves it forward given a signal, and a tool is the only path from the process to a file, a command, or a network. The term needed defining because it had stopped selecting: it was contested in the research literature in 1995, and the autonomy scales published since grade how much freedom a system is given without saying what the system is. Most agents hide the skeleton inside a loop of conditionals, where the states are implicit in the branches; declarative agents keep it written down, which is why they serve as the lens here. The model sits at one point in the machine, supplying the signal that selects a transition, and holds neither the state nor the tools nor the loop. The Build section renders the definition as a few dozen lines of Go — a switch, a map, and one interface where the model plugs in — beginning the harness this book constructs part by part.

## Key Terms

| Term | Definition |
|---|---|
| **Agent** | A state machine paired with a set of tools. The state machine decides what to do next; the tools are how it acts on anything outside itself. Coding agents differ from this skeleton in packaging rather than in kind |
| **State** | What the agent carries from one step to the next: the task, what has been tried, and what came back |
| **Transition** | The rule moving the agent from one state to the next given a signal. In an agent, signals come mostly from model output and tool results |
| **Tool** | A capability the surrounding program exposes to the model as a callable interface. The model emits a structured call; the program executes it and returns the result. Tools are the only way the model affects anything |
| **Declarative agent** | An agent whose states, transitions, and tool declarations live in a data file interpreted by a fixed runtime, so the skeleton is readable and a workflow change is a configuration edit |
