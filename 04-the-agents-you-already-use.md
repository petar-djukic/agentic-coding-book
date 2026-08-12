# The Agents You Already Use

## Learning Objectives

After reading this chapter, the reader will be able to:

1. Map a tool they already use onto the state-machine-plus-tools model.
2. Name what a given agent decides on their behalf: which tools exist, what is permitted, and what happens to context when it runs out.
3. Distinguish a packaging difference from a capability difference when comparing two agents.

## 2.1 You Are Already Running One

Most programmers acquired an agent before they went looking for one. It arrived as an editor feature, a terminal command a colleague recommended, or a checkbox in a code-review tool.

Adoption is broad enough to be the default condition. Stack Overflow's 2025 survey of 33,662 developers found 84% using or planning to use AI tools, up from 76% the year before, with 51% of professional developers using them daily [@stackoverflow2025]. Whatever those tools are, they are running inside the loop from the previous chapter, deciding things.

A programmer inherits each tool's decisions without being asked, which makes those more useful to examine than the question of which tool is best. Three of them account for most of the difference between one agent and another:

1. **The tool set** — which capabilities exist at all.
2. **Policy** — what is permitted, and what pauses for a human.
3. **Context handling** — what happens to state as a session grows past what fits.

The rest of this chapter takes those three in order across three shipping tools.

## 2.2 The Same Three Parts

Every observation below was made directly, on macOS, on 11 August 2026, against Claude Code 2.1.220, OpenCode 1.18.16, and Crush v0.88.1. Their documentation churns faster than a book can track, so what follows is what the tools reported about themselves when asked, and it is a snapshot with a date on it. Figure 2.1 sets the three side by side before the paragraphs take the axes one at a time.

**Figure 2.1** The same skeleton, three tools.

![](figures/fig-2-1-same-skeleton-three-tools.png)

*A shaded box is a component the tool exposes as configuration. The blank box marks a component absent from the configuration surface inspected, which is a statement about that inspection rather than about the tool's internals. Versions and date as given above.*

**The tool set.** All three assemble their capabilities from a declared list rather than a fixed build. Crush's configuration schema, which the binary prints on request, carries top-level `tools`, `mcp`, and `lsp` keys. OpenCode manages Model Context Protocol servers through a subcommand. Claude Code reads a `.claude/` directory holding `skills/` and `commands/`. The packaging differs; the arrangement does not, and in all three cases what the agent can do is a list somebody can read.

**Policy.** All three express permission as data. Claude Code's `settings.json` carries a `permissions.allow` list whose entries are patterns like `Bash(gh pr merge:*)`. Crush's schema exposes `permissions.allowed_tools`. OpenCode attaches a rule list to each named agent, and printing them shows entries of the form `{permission, action, pattern}` with actions including `allow` and `ask` — a directory outside the workspace, for instance, resolves to `ask`.

**Context handling.** Here the three diverge, and this is the decision most often invisible. Crush exposes a single option, `options.disable_auto_summarize`. OpenCode makes compaction an agent in its own right: `compaction` appears in its agent list alongside `build`, `plan`, and `explore`, which means the thing that decides what to forget is itself a configured component with a prompt. In the Claude Code settings inspected, no equivalent key appeared.

Across those three axes, little varies. Three tools from three vendors, and each one is a tool list, a permission table, and a context policy — the previous chapter's skeleton in different file formats.

| | Claude Code 2.1.220 | OpenCode 1.18.16 | Crush v0.88.1 |
|---|---|---|---|
| Tool set declared in | `.claude/skills/`, `.claude/commands/`, MCP | MCP servers via subcommand | `tools`, `mcp`, `lsp` schema keys |
| Permission expressed as | `permissions.allow` patterns | per-agent `{permission, action, pattern}` rules | `permissions.allowed_tools` |
| Named sub-agents | not observed in this inspection | `build`, `plan`, `explore`, `general`, `summary`, `title`, `compaction` | not observed in this inspection |
| Context policy | not in the surface inspected | `compaction` as a named agent | `options.disable_auto_summarize` |
| Configuration format | JSON plus directories | JSONC against a published schema | schema-validated config |

*Observed 11 August 2026. A blank cell records what the inspection found, not a capability the tool lacks.*

> **Common Error:** Choosing an agent on the impression its interface makes, rather than on what it decides without asking. The interface is the part that varies most between these tools and the part that matters least. The permission table is the part that determines what the agent can do to a repository, and it is three lines of configuration that nobody demonstrates.

## 2.3 Packaging Differences and Capability Differences

Telling those two apart is what the survey is for.

A packaging difference changes where a decision is written down. Whether permission rules live in a JSON file, in a schema-validated config, or attached to a named agent is packaging: the same decision is being made in all three, and moving it does not change what the agent may do.

A capability difference changes what is reachable. A tool set that includes a language-server connection can resolve a symbol across a repository; one that does not has to grep. That difference survives any amount of reformatting.

The reason to insist on the distinction is that the two are marketed identically. A configuration format is easy to demonstrate and easy to compare, so it gets the screenshots. Whether a harness can pause an unsafe write is harder to show and is the question with consequences.

Making the comparison at the level of what an agent produces is harder than it looks. A December 2025 vendor report scored 470 open-source pull requests — 320 AI-co-authored against 150 human-only — and found 10.83 issues per AI pull request against 6.45, with logic and correctness issues 75% more common [@coderabbit2025]. That contrast is AI-authored against human-authored code, not one harness against another, and it comes from a company that sells code review. It says something about the class of output; it settles nothing between the three tools above.

## 2.4 Adoption Is Not Evidence of Benefit

Daily use by half the profession is a fact about distribution, not about effect.

The most careful measurement available runs against the intuition. METR ran a randomized controlled trial with 16 experienced open-source developers across 246 tasks in their own mature repositories, randomizing whether AI tools were allowed. Tasks with AI allowed took 19% longer. The developers had forecast a 24% speedup before starting, and after finishing — having been slower — still estimated they had been sped up by 20% [@metr2025].

Two limits on that result matter. It measured early-2025 tooling, Cursor Pro with Claude 3.5 and 3.7 Sonnet, which is not the agent generation surveyed in this chapter [@metr2025]. And 16 developers on mature repositories they knew well is a narrow setting, chosen deliberately to be the hard case for a tool that has never seen the codebase.

What survives those limits is the gap between the measurement and the estimate. The same survey that reports 84% adoption also reports positive sentiment falling from over 70% to 60%, and 46% of respondents distrusting the accuracy of the output against 33% who trust it [@stackoverflow2025]. Programmers are running these tools daily while doubting them, which is a coherent position and not a comfortable one.

> **Performance Observation:** In the METR trial the developers' post-hoc estimate was wrong by about 39 percentage points, in the favorable direction, after they had personally done the work [@metr2025]. Perceived speed and measured speed came apart for people with every reason to know better. Whatever else follows, an individual impression that an agent is helping is not the same kind of claim as a measurement that it did.

The mechanism explains why the answer would be this unstable. If an agent is a loop, a tool set, and a context policy, then its effect depends on which tools were wired up, what the policy permitted, and what the loop did on failure. Two programmers running the same product, configured differently, are not running the same agent. That is the subject of the rest of this book.

## Summary

Most programmers are already running an agent, usually acquired without a decision: 84% of Stack Overflow's 2025 respondents use or plan to use AI tools, and half of professional developers use them daily. Inspected directly, three shipping tools turn out to be the same skeleton in different file formats — a declared tool set, a permission table, and a context policy — differing in where each decision is written rather than in what is being decided. Distinguishing those packaging differences from capability differences is what makes the tool market legible, and only the second kind changes what an agent can reach. Adoption says nothing about benefit: a randomized trial of experienced developers on their own repositories measured a 19% slowdown while the same developers estimated a 20% speedup, and the survey reporting the highest adoption also reports the lowest trust. An agent's effect depends on how its three parts are configured, which is why two people running the same product are not necessarily running the same agent.

## Key Terms

| Term | Definition |
|---|---|
| **Packaging difference** | A difference in where a decision is written down — file format, configuration location, interface — that leaves what the agent may do unchanged |
| **Capability difference** | A difference in what the agent can reach: a tool present in one tool set and absent from another. It survives any change of format |
| **Context policy** | The harness's rule for what happens to accumulated state when a session outgrows the context window: what is summarized, what is dropped, and what decides |
| **Agent harness** | The software wrapped around a language model: tool execution, file-system access, loop control, and policy. Named and treated in the next chapter |
