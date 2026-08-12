# Part V — Orchestration

*Outline stage — drafting notes are in `brainstorm/`.*

The machinery that makes on-the-loop development work past a single context window. Two agent roles carry the cycle — a planner that decomposes work and a generator that implements one task per isolated worktree — coordinated through a blackboard: shared state both roles read and write. GitHub issues are one blackboard and beads is another; the part also builds one from scratch, first as a YAML file, then as a Dolt database. Around the roles sits the orchestrator, which manages the sub-agent lifecycle — spawn, monitor, collect, kill — and inherits the failure modes of any distributed system.

This is the largest part and the one with the most drafted notes — the inner/outer loop split, the plan-generate-verify-repair-improve cycle, context engineering across hundreds of tasks, and the multi-agent coordination patterns are all sketched in `brainstorm/`.

Live material: [How to Loop Engineering](https://meshintelligence.substack.com/p/how-to-loop-engineering?utm_source=github&utm_campaign=agentic-coding-book) · [How to Use GitHub as Long-Term Memory](https://meshintelligence.substack.com/p/how-to-use-github-as-long-term-memory?utm_source=github&utm_campaign=agentic-coding-book) · [How to Use Git Worktrees with Coding Agents](https://meshintelligence.substack.com/p/how-to-use-git-worktrees-with-coding?utm_source=github&utm_campaign=agentic-coding-book) · [Three Commands to a Crude Orchestrator](https://meshintelligence.substack.com/p/three-commands-to-a-crude-orchestrator?utm_source=github&utm_campaign=agentic-coding-book) · [How to Code with GLM 5.2 on OpenCode](https://meshintelligence.substack.com/p/how-to-glm-52-on-opencode?utm_source=github&utm_campaign=agentic-coding-book)
