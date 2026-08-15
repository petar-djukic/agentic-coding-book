# examples/

The code the book prints, and the code it points at.

Every part of *Build a Coding Agent* closes with a Build section, and a reader
who follows them finishes with a working runtime. This directory holds that
runtime — in the state each Build section leaves it — together with the
reference profiles the book tells the reader to check their work against, and
the machinery that keeps the code here and the code on the page from drifting
apart.

Nothing here is a framework. The production evidence behind the book's claims
lives in the repositories the introduction names; this is the teaching
artifact, and it is small on purpose.

## What is where

| Path | Holds |
|---|---|
| [`MANIFEST.yaml`](MANIFEST.yaml) | the binding table: every part, snapshot, listing, and copied family |
| [`docs/VISION.yaml`](docs/VISION.yaml) | why the directory exists, and the three authority directions |
| [`docs/ARCHITECTURE.yaml`](docs/ARCHITECTURE.yaml) | the structure, the marker syntax, and the nine constraints the audit enforces |
| [`docs/road-map.yaml`](docs/road-map.yaml) | per-part status |
| [`docs/srd/`](docs/srd/) | one software requirements document per part |
| `parts/` | the runtime, one Go module per book part |
| `catalog/` | agent families and tool declarations copied from declarative-agents |
| `magefiles/` | `test`, `demo`, `audit` |

## Three directions of authority

Every question about this directory reduces to one of three, and each names
which side wins.

**Code governs prose.** A listing in a Build section is an extracted region of
part source, never retyped. `mage audit` compares the fence to its region byte
for byte, and drift is a finding. Where compiled source and printed draft
disagree, the prose is corrected — the same discipline that makes
`references.yaml` the authority for a citation, extended to code.

**Upstream governs the catalog.** `catalog/` copies canonical agent families
from declarative-agents, pinned by release tag. Copies are simplified by
deletion and never forked: a diff against the pinned tag is removals and
nothing else. An improvement worth making belongs upstream first.

**Chapter SRDs govern prose; part SRDs govern code.** The book's
[`docs/srd/`](../docs/srd/) says what a chapter must *say*. This directory's
[`docs/srd/`](docs/srd/) says what a module must *be* — its exports, what its
tests prove, what its demo exercises. A part SRD cites the chapter SRDs it
realizes and restates none of them.

## Snapshots

`parts/part-i/` is one Go module holding three packages: `c1.1/`, `c1.4/`, and
`c1.6/`. Each is the tree as that chapter's Build section leaves it.

The duplication is the point. A single evolving tree cannot produce every
listing byte for byte — Listing 1.2 prints a `verifying` case whose body is a
placeholder that Listing 4.2 replaces, and Listing 1.1 prints a profile that
does not yet declare `write_file`. At the end of Part I, neither is what the
source says. So the unit of correspondence is the Build section, which is
exactly where the runtime changes, and which is also what the reader has in
front of them at that point in the book.

What keeps three near-copies from rotting into three different programs is a
check: every file differing between consecutive snapshots is either inside a
listing region or declared in the manifest with a reason. Drift nobody
declared is a finding.

Module per part, package per chapter — so `go build ./...` and `go test ./...`
cover a whole part at once, and the parts stay independently buildable.

## How a listing resolves

A region is delimited by two comment lines, in the host language's comment
syntax:

```go
// example:begin c1.1-2
type Profile struct {
...
}
// example:end c1.1-2
```

`MANIFEST.yaml` registers each listing with its label as printed, the language
of its fence, and the regions that make it up — a listing may name more than
one, concatenated in order, for a listing the prose assembles from several
places in the source.

Regions resolve into `parts/` only, never into `catalog/`. That boundary is
load-bearing rather than tidy; see below.

## Running it

```bash
mage -d examples audit   # the constraints (the default target)
mage -d examples test    # go vet and go test ./... in every part module
mage -d examples demo    # run each snapshot end to end on its fixture
```

The audit enforces eight of the nine constraints in
[`docs/ARCHITECTURE.yaml`](docs/ARCHITECTURE.yaml): chapter ids resolve
against the book's architecture (E-C6) and a drafted Build section has a
snapshot behind it (E-C7); every copy is pinned to a real release and says
what it dropped (E-C4), with its upstream notice intact (E-C5); every listing
resolves to a marked, single-span region in `parts/` and never into `catalog/`
(E-C3); no part requires the upstream module (E-C1); consecutive snapshots
differ only where the manifest says they do (E-C2); and every chapter SRD a
part SRD cites resolves to exactly one file (E-C8). The ninth, demo
determinism (E-C9), is asserted by the tests rather than by the audit, since
checking it means running the demos twice.

Findings accumulate and report together — the audit fails once, with
everything it found, each finding naming its constraint.
[`docs/audit-baseline.yaml`](docs/audit-baseline.yaml) records accepted debt
the same way the book's does: entries are printed but do not fail, an entry
with no issue is itself a finding, and an entry matching nothing is reported
as stale, so the file cannot decay into a blanket exemption.

Both targets discover parts from `MANIFEST.yaml` rather than from the
directory tree, so a part registered before its chapters are drafted is
skipped with a printed note instead of failing the build — and a directory no
manifest entry claims is not quietly picked up. `test` runs `go vet` and
`go test ./...` in each part module; `demo` runs the `cmd/demo` entry point
each part ships.

Demos run canned. Each drives the runtime against a fixture repository with a
deterministic `Model` behind the Listing 1.2 interface — no model, no
credentials, no network — in a temporary directory the part creates and
removes, so a run that writes leaves the repository untouched and two runs
produce identical output. That is the same contract the book's root audit
holds itself to, and it is what lets anyone execute the book's own artifact.
Live-model variants are future opt-in targets and gate nothing.

Part I's demo is worth running once for its own sake. It shows the c1.4
runtime writing a wrong answer, being told so by the compiler, having a write
outside the root refused, and converging — the whole argument of §4.6 in nine
transcript lines.

The book's root `mage audit` shells in here and reports the result as one
finding. The checking code lives in `magefiles/`; the one exception is the
listing-extraction check, which must read both the chapters and the part
source, and so lives in the book's root magefiles.

## Third-party material

`catalog/` contains files copied from
[declarative-agents](https://github.com/Nokia-Bell-Labs/declarative-agents),
copyright Nokia, licensed BSD-3-Clause. This repository is MIT. The two
coexist without a `NOTICE` file, deliberately:

- **Clause 1** — retain the copyright notice in redistributed source — is
  discharged by the header staying in each copied file. `mage -d examples
  audit` checks every file under `catalog/` still carries it, so a stripped
  header is a build failure rather than a discovery.
- **Clause 2** — reproduce the notice in the documentation and other materials
  — binds binary redistribution, and nothing here ships built. `parts/`
  compiles standalone from original source; `catalog/` files are YAML
  programs.

A hand-maintained notice file would add something that drifts without adding
compliance, and the notices are already in the files, where they cannot fall
out of sync with what they cover.

**The boundary that keeps this true:** no book listing quotes catalog source.
If BSD-3-covered material were reproduced verbatim in the built PDF, clause 2
would reach the book itself and a notice would be required. The audit enforces
the boundary rather than trusting it.

Per-copy provenance — upstream path, pinned release tag, and what each copy
drops — is recorded in the `provenance:` block of each catalog entry in
[`MANIFEST.yaml`](MANIFEST.yaml). An unpinned copy is a finding: it cannot be
diffed against its source, so it drifts without anyone being able to tell.

Copies are reduced by deleting whole files, never by editing one. Every file
under `catalog/` is byte-identical to its counterpart at the pinned release,
so a diff against a later tag shows upstream change rather than local drift —
which is what makes re-syncing a review rather than an archaeology exercise.
[`catalog/README.md`](catalog/README.md) has the command.

## Status

`docs/road-map.yaml` carries it per part. Part I's Build sections are drafted,
so Part I is where the artifacts start; Parts II–V land as their Build
sections are written. The catalog grows the same way — `executor` and
`applier` now, the rest with the parts that consume them.
