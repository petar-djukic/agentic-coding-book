<!-- chapter: C1.6 -->

# Externalizing Memory

## Learning Objectives

After reading this chapter, the reader will be able to:

1. Explain why nothing survives a session unless it is written where the harness can read it back.
2. Identify what in their own practice is carried in their head and invisible to the harness.
3. Name the moment the human leaves the loop, and what has to be true before that is safe.
4. Give the skeleton from the earlier chapters a memory that survives the session gap: state written at exit and read back at the next start.

## 6.1 Nothing Survives the Gap

A programmer spends twenty minutes explaining why the retry logic has to be idempotent. The agent gets it, fixes the code, and the tests pass. The next morning, on a related task, it writes the same non-idempotent retry.

Nothing was forgotten, because nothing was ever stored. The model's weights did not change during that conversation. When the process ended, the context window was discarded along with it, and the second session began with whatever the harness assembled from disk. The twenty minutes existed only as tokens in a request that no longer exists.

> **Definition: Externalized memory** — state the harness can read back in a later session: specifications, constitutions, notes, issue trackers, committed artifacts. The model retains nothing between sessions, so anything that must survive one has to be written where the harness can reach it.

**Figure 6.1** What crosses the gap between two sessions.

![](figures/fig-6-1-what-crosses-the-gap.png)

*The shaded path is the only one that continues. A correction made in conversation and a constraint written into a specification look equally effective at the time; they differ entirely the following morning. Nothing in the loop reports the difference, because an absent constraint produces confident output the same way a satisfied one does.*

Figure 6.1 restates the harness chapters instead of adding a mechanism. The harness assembles context from files, sends a request, and applies what comes back [@latentpatterns-harness]. A session is that cycle repeated until the process exits. Persistence was never a property the loop had.

> **Common Error:** Assuming a correction persists because it worked. The correction did work, within the session where it was made, which is what makes the assumption reasonable and the failure quiet. An instruction that lives only in a conversation has a lifetime measured in one process.

## 6.2 The Window Is Not Memory Either

Within a session the situation is better, though less than it appears.

The context window has a size, and everything competes for it: the system prompt, the tool definitions, the files, and every prior turn. A long session does not accumulate understanding so much as accumulate tokens, and at some point the harness has to decide what to drop. Position inside the window matters as well, since attention is strongest at the beginning and end and weakest in the middle, so a constraint stated an hour ago sits in the least-used region of the request that is supposedly carrying it [@liu2023].

In-session memory is therefore bounded and uneven. Both properties are mechanical.

What harnesses do about this has become its own research area. A 2026 survey formalizes agent memory as a write–manage–read loop coupled to perception and action, and groups the mechanisms into five families: context-resident compression, retrieval-augmented stores, reflective self-improvement, hierarchical virtual context, and policy-learned management [@du2026]. The compaction step a coding agent runs when a session grows long is the first of those families. A repository the harness searches before answering is the second.

That survey is a 2026 preprint with no published version, and the field it surveys is four years old. The taxonomy is useful as vocabulary; the confidence attached to any particular family should be low.

None of the five removes the bound. Each one decides what to lose.

## 6.3 What Crosses the Gap

Everything that survives a session has the same property: it was written somewhere the harness reads.

That set is smaller than it feels. A specification in the repository crosses. An architecture note in a committed file crosses. An issue in a tracker the harness can query crosses. A decision made in conversation does not, and neither does the reasoning behind a decision that did — the code carries the what, and the why goes with the process.

This is the mechanism behind every artifact the rest of this book asks for. Specifications, constitutions, task decompositions, and instrumentation logs are usually read as process discipline, the kind of thing a team adopts when it grows. Under the loop described here they are something else: they are the only memory the system has. A specification is the input to the next session, not documentation of the work, and an uncommitted decision is not a decision the system holds.

Parts II through VI are largely concerned with what belongs in those artifacts. The fact that they have to exist at all is this chapter's, and it is mechanical rather than cultural.

> **From the Field:** The cheapest version of this lesson costs a day. A generation run produced a working component, the session ended, and the next run rebuilt the same component against a different set of assumptions — because the assumptions had been settled in conversation and the conversation was gone. The code was in the repository. What was missing was the sentence saying why it looked like that, and the second run had no way to know it was contradicting anything.

## 6.4 The Programmer Is a Memory Device

There is a reason this failure mode stays hidden for so long.

While a programmer sits in the session, they supply the missing context themselves. They notice the agent heading toward a decision that was settled last week and say so. They answer the question the specification did not. They recognize the wrong pattern and correct it. Every one of those is an act of memory, and the system's apparent continuity is largely theirs.

> **Definition: Human in the loop** — a mode of operation where the programmer instructs the agent, reviews each output, and corrects mistakes interactively. The programmer is the verification layer, and also the memory.

> **Definition: Human on the loop** — a mode of operation where the programmer supplies specifications and constraints in advance, the agent executes across many tasks, and results are verified through automated testing and instrumentation. The programmer builds the verification layer instead of serving as it.

The distinction is usually presented as a stance, a matter of how much a programmer trusts the tool. The mechanism makes it something narrower. In-the-loop and on-the-loop differ in where the missing context comes from: a person answering in real time, or an artifact written beforehand. That is a wiring decision, and it can be read off a system rather than argued about.

Which is why leaving the loop is not a decision made once. It is made per unit of context: each thing the programmer currently supplies from memory either gets written down or gets lost the first time nobody is watching. Writing things down is old advice, and it predates any of this [@hunt1999]. What is new is that the reader of the written thing is a process that cannot ask a follow-up question.

## 6.5 How Far the Human Has Left

Because that transition happens piecemeal, describing where a team currently sits needs more than two labels.

The vocabulary that exists for this comes from outside software. Autonomy is graded on published scales that run from fully manual operation to fully self-managing systems [@tmforum-ig1218].

> **Definition: Autonomy level** — a classification of how much of the plan-and-execute cycle a system handles without human direction, from L0, where the programmer types every line, to L5, where the system manages its own evolution from stated intent.

An earlier chapter set those scales aside, on the grounds that grading how much rope a system has been given says nothing about what the system is. That objection stands, and it is not the job being asked of them here. As a definition of an agent, a level is useless. As a description of how much of the cycle a team has stopped supplying by hand, it is exactly the right shape, because it measures the thing this chapter is about.

The levels are also vendor-consortium artifacts, with the incentive problems that implies, and they were written for network operations rather than for software construction. Used as a scale and not as a standard, they survive the transfer.

What the scale makes visible is that each step up has a prerequisite, and the prerequisite is always the same kind of thing. Moving a decision out of the programmer's head and into an artifact the harness reads is what makes the next level possible; a faster model is not. A team stuck at one level is usually not short of capability but short of written-down context, which is a diagnosis with an address.

That is the hinge this part has been building toward. The mechanism is established: the model retains nothing, the window is bounded and uneven, the harness reads only what was written, and the programmer has been silently making up the difference. Everything that follows is about what to write, how to verify what comes back, and how to run the loop when nobody is sitting in it.

## 6.6 Build: What the Skeleton Writes Down

The skeleton assembled across this part holds everything it knows in one slice: the transcript. The process exits, the slice is garbage-collected, and the next run starts empty — section 6.1's gap, reproduced in a single variable. Closing it takes less code than any earlier Build section, because the mechanism is nothing more than reading a file at the start and appending to it at the end.

Listing 6.1 closes the gap Figure 6.1 draws.

**Listing 6.1** Externalized memory: a notes file loaded before the task and appended at session end.

```go
const notesPath = "NOTES.md"

// Start runs before Run: externalized state enters the transcript
// first, so the notes precede the task Run appends.
func (a *Agent) Start() {
	notes, err := os.ReadFile(filepath.Join(a.root, notesPath))
	if err == nil {
		a.transcript = append(a.transcript, string(notes))
	}
}

// End writes what the next session is allowed to know.
func (a *Agent) End(decision string) error {
	f, err := os.OpenFile(filepath.Join(a.root, notesPath),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(decision + "\n")
	return err
}
```

*Everything that crosses the gap passes through `End`; everything else dies with the transcript. The `if err == nil` on the read is the first session's whole experience of the file not existing yet.*

The load-bearing element is the `decision` argument, and the listing cannot supply it. Something has to choose which sentence is worth carrying forward, and in this skeleton that something is whoever calls `End` — which makes section 6.4's wiring question visible in a call site. A programmer typing the sentence at the end of each session is in the loop, serving as the selector. A specification committed to the repository, read by `Start` the same way the notes are, is that selection done in advance, and the programmer who wrote it has left this particular loop.

The part's Build sections now add up to the machine Part I described: a tool boundary, a policy check on writes, a gate feeding consequences back, and a memory that survives the process. What the skeleton does not have is anything to remember — it has never been given a specification or a constitution. Writing those artifacts is the practice the next part opens with, and the skeleton is ready to read them.

## Summary

A session ends and the model retains nothing; the next session begins with whatever the harness assembles from disk. In-session memory is no substitute, because the context window is bounded and attention within it is uneven, and the five families of memory mechanism a 2026 survey identifies each decide what to lose rather than removing the bound. What crosses the gap is what was written where the harness reads: specifications, constitutions, trackers, committed artifacts. That is why the rest of this book asks for those artifacts — not as process discipline, but because they are the only memory the system has. The reason the gap stays hidden is that the programmer fills it, answering in real time what was never written down, which makes in-the-loop and on-the-loop a wiring question, not a matter of trust. Autonomy levels describe how far that transition has gone, and each step up requires moving another decision out of a person's head and into a file. The Build section closes the part's running skeleton with the property this chapter is about: a notes file read at session start and appended at session end, which is the only part of a session the next one sees.

## Key Terms

| Term | Definition |
|---|---|
| **Externalized memory** | State the harness can read back in a later session: specifications, constitutions, notes, trackers, committed artifacts. The only memory the system has |
| **Context window** | The maximum tokens a model can process in one request, covering input and output alike. Bounded, and unevenly attended within |
| **Human in the loop** | The programmer instructs, reviews each output, and corrects interactively, serving as both the verification layer and the memory |
| **Human on the loop** | The programmer supplies specifications and constraints in advance and verifies through automated checks, having built the verification layer rather than being it |
| **Autonomy level** | How much of the plan-and-execute cycle runs without human direction, from L0 (fully manual) to L5 (self-managing). A measure of how much context has been externalized, not a definition of an agent |
