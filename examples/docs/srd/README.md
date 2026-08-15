# Software Requirements Documents (part SRDs)

One SRD per book part, written before that part's snapshots are. The SRD is
the implementation contract: a sub-issue executes it without re-deriving the
design, and `mage -d examples audit` checks the part of it a machine can
check.

Files are named `srd-part-<id>.yaml`, matching the part ids in
[../ARCHITECTURE.yaml](../ARCHITECTURE.yaml) and
[../../MANIFEST.yaml](../../MANIFEST.yaml).

## What separates these from the book's SRDs

Two subjects, not two layers of one.

[`../../../docs/srd/`](../../../docs/srd/) holds one SRD per **chapter**. Those
are writing contracts: a chapter's goal, its content spine, its citations, its
apparatus. They govern what the prose must say.

This directory holds one SRD per **part**. These are software contracts: what
the module exports, what its tests must prove, what its demo must exercise.
They govern what the code must be. None of that appears in a chapter SRD, and
none of a chapter SRD's content is repeated here.

A part SRD names the chapter SRDs it realizes and stops there. The chapter
contract is cited, never restated — the same relationship a citation has to
`references.yaml` (`../VISION.yaml`: A3).

## How realizes: resolves

`realizes:` is a list of chapter SRD ids in the form `srd-<part>.<chapter>` —
`srd-1.1`, `srd-1.4`. Each resolves to the unique file under `docs/srd/` whose
name begins with that id followed by a hyphen; `srd-1.1` resolves to
`srd-1.1-what-is-an-agent.yaml`. An id matching no file, or more than one, is
a finding under **E-C8**.

The indirection is deliberate. A chapter can be retitled without touching this
directory, and a chapter that is renumbered breaks the reference loudly
instead of pointing at the wrong contract. It is the same two-hop shape the
chapter marker already uses, for the same reason.

## Fields

| Field | Content |
|-------|---------|
| `meta` | part id, title, module path, and the manifest entry this specifies |
| `realizes` | chapter SRD ids, resolved as above. Every Build section in the part is represented |
| `goal` | one sentence: what a reader who has finished this part's Build sections possesses |
| `constraints` | the ids from [../ARCHITECTURE.yaml](../ARCHITECTURE.yaml) this part is checked against. Required — a part naming none has nothing to fail |
| `snapshots` | ordered list of `{chapter, package, leaves, adds, exports}`. One entry per Build section, in reading order. `exports` is the identifiers that snapshot introduces, which is what the listings quote and what the tests exercise |
| `dependencies` | what the module may import. Standard library only unless stated, and never the upstream module (E-C1) |
| `test_contract` | list of `{proves, snapshot}`: the behaviours the tests must establish. Written as claims about the runtime, not as test function names |
| `demo_contract` | what `mage -d examples demo` runs, what the fixture contains, and what the run must produce |
| `acceptance` | implementation-readiness checks, verifiable by running something |
| `gaps` | known holes: what the prose narrates that the source cannot yet reproduce, and what is deferred to a later unit |

## The implementation contract

1. **The SRD comes first.** A snapshot is not written before the SRD that
   specifies it. Where a Build section is already drafted and the SRD is
   written after it, the SRD says so and records the disagreements in `gaps`.

2. **Exports are what the listings quote.** An identifier a listing prints is
   an identifier the SRD names. The reverse does not hold — a snapshot may
   carry seams the prose narrates without printing, and the SRD is where those
   are written down so they do not get dropped as invisible.

3. **The code is the authority.** Where compiled source and printed draft
   disagree, the prose is corrected (`../VISION.yaml`: A1). The SRD records
   the disagreement in `gaps` so the correction is scoped rather than
   discovered.

4. **Snapshots differ only by what the prose narrates.** Every difference
   between consecutive snapshots is in a listing region or declared in the
   manifest (E-C2). The `adds` field is where the SRD states the intended
   difference in advance.

5. **No upstream dependency.** The reader builds their own runtime. A part
   that needs `declarative-agents` to compile has abandoned the premise
   (E-C1).

## Coverage

| Part | SRD | Status |
|---|---|---|
| P1 — Agents and Harnesses | `srd-part-i.yaml` | written (GH-133) |
| P2 — Construction and Requirements | — | waits on C2.4 and C2.5 being drafted |
| P3 — Testing | — | waits on Part III |
| P4 — Correctness | — | waits on Part IV |
| P5 — Agent Orchestration | — | waits on Part V |

Parts II–V wait on their Build sections existing rather than on effort.
`../road-map.yaml` carries the gate.
