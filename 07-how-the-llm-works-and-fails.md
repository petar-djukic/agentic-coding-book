<!-- chapter: C1.5 -->

# How the LLM Works, and Fails

## Learning Objectives

After reading this chapter, the reader will be able to:

1. Describe what an LLM does — next-token prediction — and distinguish it from common misconceptions (search engine, database, reasoning engine).
2. Explain what tokens are, how they differ from words, and why this matters for context budgets.
3. Define the context window and identify its constraints on agent-based development.
4. Explain the interpolation model of LLM behavior and use it to predict when the model will fail.
5. Identify the four failure modes of in-context generalization and explain why each makes external verification necessary.

## 5.1 What an LLM Is (and Is Not)

Programmers come to coding agents with a mental model of what the model "is." The mental model determines their expectations, and wrong expectations produce wrong results. Four common models are wrong in specific ways.

"It's a search engine." Search engines retrieve existing documents. The model does not retrieve — it generates. It produces text that did not exist before, token by token, conditioned on what is in context. A search engine returns nothing when it has no match. The model always returns something, even when it has no good answer.

"It's a database of code." Databases store and retrieve records. The model does not store code and look it up. Its weights encode patterns from training data — statistical regularities, not retrievable records. Asking the model for a specific function from a specific library is not a query. It is a generation task that may or may not produce the correct function, depending on how well the training data covered that library.

"It's like talking to a junior developer." Junior developers have goals, learn from feedback within a conversation, and know what they do not know. The model has none of these properties. It does not learn from the current conversation — its weights are fixed. It does not know what it does not know. It generates the next token regardless of whether it has sufficient information to do so correctly.

"It's autocomplete on steroids." This is the closest to correct but still misleading. Autocomplete predicts the most likely next word given a short local context. The model predicts the next token given the entire context window — which can include specifications, architectural constraints, conversation history, and tool outputs. The difference in context scope produces a qualitative difference in capability.

What the model actually is: a neural network trained on a large text corpus to predict the next token in a sequence.

> **Definition: Large language model (LLM)** — a neural network trained on a large text corpus to predict the next token in a sequence. At inference time, the model generates text by iteratively predicting one token at a time, conditioned on the full context provided.

The model was trained by processing billions of text sequences and adjusting its internal parameters (weights) to become better at predicting what token comes next. The training process ran once. The weights are now fixed. At inference time — when the model generates a response — no learning occurs. The model applies the patterns encoded in its weights to the context it receives, and generates one token at a time.

The architecture that makes this work is the transformer [@vaswani2017]. A transformer-based LLM has three parts:

1. **An embedding layer** that converts each input token into a high-dimensional vector — a list of hundreds or thousands of numbers that encode the token's meaning in relation to other tokens.

2. **A stack of transformer blocks** — dozens to hundreds of them — that process these vectors through two operations: attention (which relates each token to every other token in the sequence) and a feed-forward network (which transforms the representation at each position). Each block refines the internal representation.

3. **An output layer** that converts the final representation into a probability distribution over every token in the vocabulary — typically over 100,000 tokens. The model selects the next token from this distribution.

The model generates one token, appends it to the sequence, and runs the process again to generate the next token. This continues until the model produces a stop token or hits a length limit.

> **Performance Observation:** The generation process has two phases that programmers observe directly. The **prefill** phase processes the entire input at once — this is the initial pause before output begins. The **decode** phase generates tokens one at a time — this is the streaming text that follows. Long inputs produce long prefill pauses. Long outputs produce long decode times. These are different bottlenecks with different costs.

## 5.2 Tokens and the Context Budget

A token is not a word.

> **Definition: Token** — the atomic unit of text that a language model processes. Tokens are produced by a tokenizer — a deterministic mapping between text strings and numerical identifiers. Common words map to single tokens. Uncommon words, technical terms, and code constructs are split into multiple tokens.

The tokenizer is not a neural network. It is a lookup table, built before training, that defines how text is split into pieces the model can process. "The" is one token. "Transformer" might be two tokens ("Trans" + "former"). A Python function signature like `def calculate_total(items: list[dict]) -> float:` might consume 15–20 tokens. Code is less token-efficient than prose — more tokens per semantic unit.

Modern models have vocabularies of over 100,000 tokens. Every token the model has ever seen in training, and every token it will generate, is drawn from this fixed vocabulary.

Why this matters: the context window is measured in tokens, not words. Everything the model sees and generates must fit within it.

> **Definition: Context window** — the maximum number of tokens a model can process and generate in a single request. The window includes the input (system prompt, conversation history, tool definitions, tool outputs, retrieved documents) and the output (the model's response, including any reasoning).

Context windows have grown rapidly — from 8,000 tokens (early GPT-4) to 128,000 (GPT-4 Turbo) to 200,000 (Claude 3) to over 1,000,000 (current frontier models). These numbers are large but not infinite, and for agent-based development, they are consumed faster than programmers expect.

Consider what occupies context during a typical coding agent session:

| Component | Typical size |
|---|---|
| System prompt and tool definitions | 5,000–15,000 tokens |
| Architectural constraints and constitution | 2,000–10,000 tokens |
| Source files read into context | 1,000–50,000 tokens per file |
| Conversation history (prior turns) | Grows with each exchange |
| Tool outputs (compiler errors, test results) | 500–5,000 tokens per call |
| The model's own response | 500–4,000 tokens |

An orchestration pipeline that runs hundreds of tasks, reading source files and processing tool outputs at each step, can exhaust even a million-token context window. Context management — deciding what enters the window and what is evicted — is an engineering problem addressed in Part V.

Two properties of context that affect agent behavior:

**Context is consumed by both input and output.** A 200,000-token window does not mean 200,000 tokens of input. The model's response, including any internal reasoning, also occupies context. A model that "thinks" extensively before responding may consume tens of thousands of tokens on reasoning alone.

**Attention is strongest at both ends of the context and weakest in the middle.** Recall is highest for material at the beginning of the window and for material at the end; what sits between them is recalled least reliably. This "lost in the middle" effect [@liu2023] means that how context is structured matters, not just what is in it. Position is an engineering decision rather than a stylistic one — and the two strong positions are not interchangeable, because they are strong for different reasons.

> **Good Practice:** Place stable, reusable context (system prompts, constitutions, architectural constraints) at the beginning of the window, and task-specific, variable content (the current file, the current requirement) at the end. The beginning earns the stable material twice over: it is well attended, and it is the only position that can be cached — prefix caching lets the serving infrastructure reuse the computation for the unchanging head of the context, cutting latency and cost on every subsequent request. The end earns the current task because recency favors it. What lands in the middle is what you can most afford to lose.

## 5.3 The Interpolation Machine

How should a programmer think about what the model "does" when it generates code?

"All models are wrong, but some are useful" — a line usually attributed to the statistician George Box, who published it in this form with Draper [@boxdraper1987]; the 1976 paper it is often credited to argues the point without containing the sentence [@box1976]. The point is not that models are bad. The point is that the value of a model lies not in whether it is true but in whether it helps you make better decisions. The question for this section is not "what is the LLM really doing inside?" — that question is open and may remain open. The question is: which mental model of LLM behavior helps a programmer get better results from it?

The answer matters because the mental model determines what the programmer does. A programmer who thinks the model "understands" the problem will trust the output and skip verification. A programmer who thinks the model "just guesses randomly" will not bother writing specifications, since random processes do not respond to better inputs. Both models are wrong, and both lead to worse outcomes. The right mental model should predict when the model will succeed, when it will fail, and — most importantly — what the programmer can change to shift the outcome.

The model is not retrieving memorized code. It is not executing logic. It is not reasoning about the problem the way a programmer reasons. It is doing something that has no clean everyday analogy, but the closest productive metaphor is **interpolation**.

Given the context — the specification, the architectural constraints, the conversation history, the tool outputs — the model constructs a response by interpolating across the patterns it learned during training. "Interpolation" here means: the model finds the point in its learned space that is most consistent with the context it was given, and generates text from that point.

This metaphor is supported by three theoretical accounts of how in-context learning works. None is fully settled. Together, they provide a productive framework — not because they are true, but because they are useful.

**Pattern completion.** The most mechanistically concrete account comes from work on induction heads — specific attention patterns that implement a pattern-completion algorithm [@olsson2022]. If the sequence [A][B] has appeared in context, and [A] appears again, the model predicts [B]. These patterns develop during training through a detectable phase transition — a sudden jump in in-context learning ability that coincides with the formation of these attention structures. The patterns generalize beyond literal copying to abstract pattern matching, which begins to explain why a model can complete novel task patterns from a few examples.

**Implicit learning.** A second account proposes that the forward pass of a transformer implements something equivalent to gradient descent on the in-context examples [@akyurek2022; @vonoswald2023]. The model has learned, during pre-training, to run a learning algorithm inside its forward pass. This explains why more examples in context improve performance (more gradient steps) and why the model generalizes to new tasks (gradient descent is task-agnostic).

**Task inference.** A third account frames in-context learning as Bayesian inference over a latent task variable [@xie2021]. The pre-training corpus contains text from many different "tasks" (writing styles, subject domains, instruction types). At inference time, in-context examples update the model's posterior over which task is being requested, and the model generates text that is likely under the inferred task. This predicts a specific failure mode: when the task is far outside the pre-training distribution, the model has no prior to update, and in-context learning fails.

These accounts are not mutually exclusive. The productive synthesis: the model has, through training on a vast corpus, developed internal mechanisms for recognizing task structure from examples, abstracting patterns beyond literal copying, and generating responses consistent with inferred task structure.

The interpolation metaphor captures what matters for agentic coding:

- The model's behavior is a function of what is in context. The weights are fixed. The context is the programming interface. What you put in context determines what the model interpolates from.
- Interpolation works when the target is adjacent to training data. The model generates correct code for common patterns, standard APIs, well-documented libraries — because the interpolation space is dense in those regions.
- Interpolation fails when the target is novel. Proprietary APIs, unusual architectures, domain-specific conventions that do not appear in training data — these are sparse regions where the model interpolates from the wrong anchors. The output is plausible but wrong.
- The model does not know the difference. It generates with the same confidence whether the interpolation is good or bad. There is no internal uncertainty signal.

> **Common Error:** Treating the model as a reasoning engine that "understands" the problem. The model interpolates. When the interpolation is good, the output looks like understanding. When the interpolation is bad, the output looks like understanding too. The difference is not visible in the output — only in the verification results.

> **From the Field:** The interpolation model changed how I write specifications. When I described architecture in prose — "use the repository pattern with dependency injection" — the model interpolated from the average repository pattern in its training data, which was not my repository pattern. When I included a concrete example — an actual interface definition with the function signatures I wanted — the model interpolated from that example. The closer the context is to what I want, the closer the interpolation lands. Specifications are interpolation anchors.

## 5.4 What the Model Cannot Do

The interpolation model predicts specific failure modes. These are not edge cases — they are structural properties of how the model works. Understanding them before relying on agent-generated code is not optional.

### 5.4.1 Out-of-Distribution Extrapolation

The model interpolates from training data. When the target is outside the distribution — a proprietary API the model has never seen, a novel architectural pattern, a domain-specific convention that does not appear in public code — the interpolation anchors are wrong. The model generates something plausible from the nearest in-distribution examples, which may be the wrong analogy entirely.

This failure mode is predictable. If the code you need is similar to code that exists in large quantities on the internet — standard REST APIs, common design patterns, popular frameworks — the model will generate it well. If the code is specific to your organization, your architecture, or your domain — the model will guess. The guesses will compile. They may even pass tests that were generated from the same wrong assumptions. They will not match your intent.

The mitigation is context. Providing examples, interface definitions, and architectural constraints in context moves the interpolation target from "the average pattern in training data" to "the pattern adjacent to what you showed the model." This is why specifications matter more for novel code than for standard code — and why constructional intent must be explicit for anything that deviates from common patterns.

### 5.4.2 Sensitivity to Framing

The model's output is sensitive to how context is structured in ways that are not always predictable. The order of examples matters — position carries weight, in the pattern described earlier in this chapter: the ends of the window are attended to more reliably than the middle, so the same example moved from the end to the middle of a long context can stop influencing the output. The format of instructions matters — numbered lists produce different behavior than prose paragraphs. The verbosity of specifications matters — terse specifications leave more room for interpolation (more guessing), while verbose specifications constrain the output more tightly.

This sensitivity is a property of interpolation: small changes to the interpolation anchors shift the output. It is the reason prompt engineering exists as a discipline, and it is the reason prompt engineering feels fragile — because it is. A prompt that works for one task may fail for a similar task because the interpolation landed differently.

For agentic coding, the practical response is not to chase the perfect prompt. It is to build verification that catches the cases where framing shifts the output in unintended directions. The verification stack exists because no prompt is reliable enough to make verification unnecessary.

### 5.4.3 Context Poisoning

Everything in context is treated as evidence. The model does not distinguish between correct and incorrect information in its context window. Stale code from earlier in a long agent session, conflicting instructions from different sources, incorrect tool outputs from a flaky test — all of these shift the interpolation.

This failure mode is insidious because it accumulates. In a short conversation, context is fresh and consistent. In a long agent session — dozens of tool calls, hundreds of messages — the context accumulates stale information, superseded decisions, and contradictory state. The model interpolates from all of it, weighting recent context more heavily but not ignoring the rest.

Context hygiene — actively curating what is and is not in the model's window — is a design discipline, not an afterthought. The orchestration techniques in Part V include context management strategies: clearing stale state between tasks, isolating sub-agents with focused context, and compressing history to preserve relevant information while evicting noise.

### 5.4.4 Confident Miscalibration

The model does not know what it does not know. When interpolation fails — when the target is out of distribution, when context is poisoned, when framing has shifted the output — the model generates a response that is fluent, well-structured, and confident. There is no hesitation, no uncertainty signal, no indication that the interpolation was bad.

This is the most dangerous failure mode for agent-based development. A programmer reviewing generated code sees code that looks correct. The variable names are sensible. The structure is clean. The comments explain the logic coherently. Nothing in the surface appearance reveals that the underlying approach is wrong — that the data model does not match the actual schema, that the API contract contradicts the specification, that the concurrency model has a race condition the model has no mechanism to detect.

This is why verification is not optional. The verification stack — compiler, linter, tests, coverage, specification conformance, intent conformance — exists because the model's confidence is not evidence of correctness. Plausible is not the same as correct.

> **Good Practice:** Assume the model is wrong until verification proves otherwise. The interpolation machine produces plausible output by design. The verification stack exists because the model cannot verify its own output. Trust the verification results. Do not trust the code's appearance.

## 5.5 Sampling: Controlling the Interpolation

When the model generates a token, it does not produce a single answer. It produces a probability distribution over every token in its vocabulary — over 100,000 candidates. The next token is selected from this distribution, and three parameters control how that selection happens.

> **Definition: Sampling** — the process of selecting the next token from the model's predicted probability distribution. Sampling parameters control the tradeoff between predictability and variety in the output.

**Temperature** adjusts the probability distribution before selection. A lower temperature sharpens the distribution — the most likely tokens become even more likely, and unlikely tokens become even less likely. A higher temperature flattens it — more tokens become plausible candidates. Setting temperature to 0 makes selection deterministic: the model always picks the single most probable token.

**Top-k** filters the distribution after it is computed: keep only the k most likely tokens, discard the rest, and re-normalize the probabilities among the survivors. Top-k with k=1 is equivalent to temperature 0 — always select the most likely token.

**Top-p** (nucleus sampling) filters differently: keep the smallest set of tokens whose combined probability exceeds p, discard the rest, and re-normalize. This adapts to the shape of the distribution — when the model is confident (one token dominates), few tokens survive. When the model is uncertain (probability is spread), more tokens survive.

For agentic coding, the practical implication is that code generation tasks — where the specification defines what the output should be — benefit from low temperature. The specification constrains the correct output; high temperature introduces unnecessary variation. Creative tasks — brainstorming names, generating documentation, exploring alternative approaches — tolerate higher temperature because there is no single correct answer.

Most coding agent frameworks set sampling parameters automatically. The reader does not need to tune them for routine work. The concept matters because it reveals something about what the model is doing: it is not producing "the answer." It is producing a distribution and sampling from it. Every response is one draw from a space of possibilities. A different random seed would produce a different response. The verification stack must hold regardless of which draw the model produced.

## 5.6 Why This Matters for the Rest of the Book

The mental model in this chapter — the model as an interpolation machine working inside a fixed context window — connects to every subsequent part of the book.

**The model is an interpolator.** What is in context determines what it interpolates from. This means context engineering is a discipline, not a convenience. Specifications, constitutions, and architectural constraints are not documentation — they are the programming interface for the model's behavior. Part II (Construction and Requirements) addresses how to write specifications that produce reliable interpolation. Part V (Agent Orchestration) addresses how to manage context across hundreds of tasks.

**The model cannot verify its own output.** Interpolation produces plausible output regardless of correctness. External verification is the only mechanism for establishing that generated code is correct. Part III (Testing) addresses how to test code that was not written by a human. Part IV (How Do You Know Your Code Is Correct?) addresses what "correct" means when the programmer did not write the code.

**The model interpolates from training patterns.** Novel constructional decisions — the architectural choices that define how software is built — are exactly the kind of thing the model will get wrong, because they are specific to the programmer's intent and rarely appear in training data. Externalizing constructional intent and defining the construction order are not optional practices for ambitious projects — they are the mechanisms that prevent the model from filling architectural gaps with training-data averages.

**The model runs inside a loop it does not control.** The previous chapter traced that loop: the harness assembles context, applies what comes back, verifies it, and feeds the result in as the next input. Everything in this chapter is a property of the part inside that loop. Part V addresses what happens when one loop is not enough.

## Summary

A large language model is a next-token predictor — a neural network that generates text one token at a time by interpolating across patterns learned during training. The context window is the fixed-size buffer of tokens the model can see; everything the model knows about the current task must fit within it, and context is consumed by input, output, and reasoning alike. The interpolation model explains both the model's capabilities (generating correct code for common patterns) and its failure modes (out-of-distribution extrapolation, sensitivity to framing, context poisoning, and confident miscalibration). Sampling parameters control the tradeoff between predictability and variety, but do not eliminate the fundamental property that every response is one draw from a distribution. Verification is not optional because the model cannot distinguish good interpolation from bad.

## Key Terms

| Term | Definition |
|---|---|
| **Large language model (LLM)** | A neural network trained on a large text corpus to predict the next token in a sequence, generating text by iterative token prediction conditioned on context |
| **Token** | The atomic unit of text a language model processes, produced by a deterministic tokenizer that maps text strings to numerical identifiers |
| **Context window** | The maximum number of tokens a model can process and generate in a single request, including input, output, and reasoning |
| **Interpolation** | The process by which a model constructs a response by finding the point in its learned space most consistent with the provided context — more than pattern matching, less than reasoning |
| **Sampling** | The process of selecting the next token from the model's predicted probability distribution, controlled by temperature, top-k, and top-p parameters |
| **Confident miscalibration** | The failure mode where the model generates fluent, confident output regardless of whether the interpolation is correct — making surface inspection unreliable as a verification method |
