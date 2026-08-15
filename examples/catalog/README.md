# catalog/

Copies of the grown version, so the book's promise can be checked.

Part I's Build section ends by telling the reader that a grown version of the
same design exists — a catalog of profiles and a runtime hardened for
production — and that checking their work against it is part of the exercise.
This directory is that comparison made available: the relevant pieces of
[declarative-agents](https://github.com/Nokia-Bell-Labs/declarative-agents),
copied at a pinned release, so the reader does not have to clone another
repository to take the book up on it.

Nothing here is compiled, imported, or quoted in the book. It is read.

## What is here

| Path | Corresponds to |
|---|---|
| `agents/executor/` | Listings 1.1 and 1.2 — the same machine, grown |
| `tools/filesystem/` | Listing 4.1 — the same write, declared |

`agents/executor/profile.yaml` names a machine and a tool list, exactly as
Listing 1.1 does. `machine.yaml` is Listing 1.1's states and transitions with
a purpose, invariants, a budget, and a lifecycle narrative attached — the same
skeleton, carrying what production needs and the teaching version omits.

`tools/filesystem/write.yaml` is the more useful comparison. Listing 4.1
writes a file and refuses a path outside the root, in fourteen lines. The
declaration for the same word states its parameters, the signals it emits on
success and failure, what it touches (`side_effects`), whether that is
reversible, and how to undo it (`file_snapshot_restore`, from the write
receipt). None of that changes what the tool does. All of it is what a machine
needs in order to reason about the tool without running it.

## Copy, not fork

Files are copied from upstream and reduced **by deletion of whole files**.
Every file retained here is byte-identical to its counterpart at the pinned
release, so a diff against a later tag shows real upstream change rather than
local drift. What each copy dropped is recorded in its `provenance.simplified`
field in [`../MANIFEST.yaml`](../MANIFEST.yaml).

An improvement worth making to a catalog member is made upstream. Nothing in
this directory is a place to fix something.

## Re-syncing against a newer release

```bash
git -C ../declarative-agents show <tag>:<provenance.path>/<file> \
  | diff - agents/executor/<file>
```

An empty diff means the copy is current. A non-empty one is upstream change
to review and then take, by re-copying the file and moving the `release:` pin
in the manifest. `mage -d examples audit` will not do this for you — it checks
that the pin names a real tag and that headers survived, not that the pin is
recent. A copy going stale is a decision to make, not a build failure.

## Not copied

`applier` was named for Chapter 4 when this directory was planned, on the
assumption that it is edit application as an agent family. It is not. At the
pinned release it is a deployment actuation boundary: it starts REST request,
control, and monitor listeners, waits for a lifecycle exit event, and carries
invariants about chart, release, namespace, and credential authority. A reader
sent here from §4.6 to compare their write tool would have found a service for
deploying Helm charts. The filesystem tool declarations are the real
counterpart.

The families with no book consumer stay upstream — `bench`, `mock`, `monitor`,
`chatbot`, `knowledge-manager`, `assembler`, `jurist`. The ones the later
parts need arrive with those parts: `critic` with Part III,
`specification-critic` and `lifecycle-exit` with Part IV, `planner` and
`collector` with Part V.

## License

BSD-3-Clause, copyright Nokia. Every file retains its header; that is what
discharges clause 1, and `mage -d examples audit` checks it. The reasoning for
why no `NOTICE` file accompanies them, and the boundary that keeps it true, is
in [`../README.md`](../README.md) under Third-party material.
