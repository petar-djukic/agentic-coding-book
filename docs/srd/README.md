# Section Requirements Documents (SRDs)

One SRD per chapter, written before the chapter is drafted. The SRD is the
drafting contract: a chapter issue executes its SRD without re-deriving the
plan, and `mage audit` checks the draft against the parts of it a machine
can check. It is a writing
spec, not a systems-engineering ceremony.

Files are named `srd-<part>.<chapter>-<slug>.yaml`, matching the chapter ids
in [../ARCHITECTURE.yaml](../ARCHITECTURE.yaml).

## How a chapter binds to its SRD

Two hops, both explicit, neither inferred:

1. The chapter file names itself on its first line, as an HTML comment:
   `<!-- chapter: C1.6 -->`.
2. `ARCHITECTURE.yaml` maps that chapter id to its `srd:` path.

So the binding survives both a retitle and a renumber, and `mage audit` reports
an unmarked or unknown chapter rather than quietly checking nothing.

The marker is a comment rather than YAML front matter on purpose: pandoc merges
metadata across every concatenated input file, so a `chapter:` key would share a
namespace with the book's real metadata, while a comment renders to nothing in
every output format (GH-25).

With the binding in place, `mage audit` checks the mechanical half of the
contract: every `apparatus.key_terms` entry appears as a row in the chapter's
Key Terms table, and every citation with `role: anchor` is actually cited.
Terms and sources the SRD merely lists as `context` or `evidence` are not
required to appear -- an anchor is the source the SRD says carries the chapter,
so its absence is a defect rather than a choice.

Matching is against the table's term column, not the section text. A term named
inside another row's definition is not a term the table defines, and the looser
reading hid three real gaps until GH-50. A trailing gloss is ignored, so
`**Large language model (LLM)**` satisfies `large_language_model`.

`mage audit` also checks the reverse: every term a chapter's Key Terms table
defines must exist in `definitions.yaml`, whether or not the chapter has an SRD
yet. Terms get coined while drafting -- the four harness responsibilities and
the packaging/capability distinction both were -- and the glossary is where
they have to land (GH-52). Separator style does not matter: the glossary key
`file_system_access` and the table row **File-system access** are one term.

**What belongs in `apparatus.key_terms`.** Only terms the chapter itself
introduces, which `definitions.yaml` records per term in its `introduced`
field. A chapter's table may repeat a term defined elsewhere for the reader's
convenience, but the SRD must not require it. **`mage audit` enforces this
since GH-88**, where the rule had gone unchecked long enough for seven Part I
entries to break it -- `srd-1.4` alone required four, including one a Part II
chapter introduces, which inverted the book's dependency direction. Listing a
term owned by another chapter is also how five of the six gaps in GH-50 arose -- `srd-1.1` and `srd-1.3`
demanded `agentic_loop`, which C1.4 introduces, and `srd-1.6` demanded
`knowledge_manager` and `constitution`, owned by P5 and C2.5.

| Field | Content |
|-------|---------|
| `meta` | chapter id, title, parent part |
| `section_goal` | one phrase, rendered as "The goal of this chapter is to ..." |
| `goals` | list of `{id, goal}`: the subgoals that lead to `section_goal`. Each `id` is `G<n>.<m>`, a subgoal of book goal `G<n>` from [../VISION.yaml](../VISION.yaml) |
| `chain` | the derivation-chain links this chapter owns, from [../constitutions/argument.yaml](../constitutions/argument.yaml). Every chapter owns at least one; a chapter owning none is not carrying argument. `mage audit` enforces this in both directions since GH-86 — a link may not name a chapter ARCHITECTURE does not define, and a chapter may not go unowned. Link ids are permanent and never renumbered, because SRDs cite them |
| `constitutions` | the rule sets this chapter is checked against, as `{file, rules}`. Required — a chapter with no rule set has nothing to fail |
| `objective` | one sentence: how the goals are achieved |
| `prior_material` | list of `{path, offers}`: material to quarry. The corpus-to-part map lives in ARCHITECTURE `material_sources` |
| `citations` | list of `{id, role, note}`. `role` is `anchor` (carries a theme), `survey` (breadth pointer), `evidence` (supports one claim), `context` (frames the field), or `counterpoint` (complicates the claim). Every `id` resolves in [../../references.yaml](../../references.yaml) |
| `content` | ordered list of `{say, cites}`: what the chapter will say, and which citations carry it. This is the spine the drafter follows |
| `apparatus` | the chapter furniture the voice constitution requires: `objectives` (draft learning objectives), `sidebars` (planned, by type), `key_terms` (ids from [../definitions.yaml](../definitions.yaml)) |
| `figures` | optional list of `{shows, status: ready\|planned}`; `shows` is a full sentence describing what the figure shows |
| `links` | `requires`: chapter ids this one builds on, which must precede it in reading order; `supports`: chapter ids that consume its output |
| `gaps` | corpus holes worth filling before or during drafting |
| `acceptance` | drafting-readiness checks |

## The drafting contract

1. **The SRD comes first.** A chapter is not drafted before its SRD exists
   ([../constitutions/process.yaml](../constitutions/process.yaml):
   `drafting_pipeline`). Where drafted prose already exists and the SRD is
   written after it, the SRD says so and the draft is corrected against it.

2. **Definitions are not re-coined.** Terms come from
   [../definitions.yaml](../definitions.yaml). A term the chapter needs and
   the glossary lacks is added there first, in the same unit.

3. **Every chapter answers a chain link.** `chain` names the links from
   `argument.yaml` the chapter owns. A prescription that cannot name the Part
   I fact it rests on is a defect (`argument.yaml`: A-C4).

4. **Citations carry claims, not the argument.** Each `content` entry names
   the citation ids that support it. A `say` with no `cites` and no
   `mechanism` evidence is a claim the drafter must ground or drop
   (`argument.yaml`: claims_register).

5. **SRDs never restate the constitutions.** They reference rule sets by
   file and id. Register, claim discipline, and venue rules live in
   [../constitutions/](../constitutions/) and are read from there.

## Coverage

Parts I through V have an SRD per chapter. Part I was written at GH-17 against
the arc EPIC #4 established; Part II at GH-66, where C2.1 and C2.2 were written
against prose that already existed and the other five before drafting. Part VI
is the remaining follow-on work:

| Part | Chapters | SRD status |
|---|---|---|
| P1 — Agents and Harnesses | C1.1–C1.6 | written |
| P2 — Construction and Requirements | C2.1–C2.7 | written (GH-66); C2.1 and C2.2 were written against existing prose, and their `gaps` fields carry the draft-versus-contract disagreements |
| P3 — Testing | C3.1–C3.5 | written (GH-67) |
| P4 — Correctness | C4.1–C4.5 | written (GH-68) |
| P5 — Agent Orchestration | C5.1–C5.9 | written (GH-69); C5.4 and C5.5 carry background dependencies in their `gaps` fields, and `road-map.yaml` gates `rel05.0` on them |
| P6 — Instrumentation | C6.1–C6.7 | to write |

The Parts II–VI SRDs wait on those parts stabilizing rather than on effort:
writing a contract against a chapter list that is still moving produces a
contract that has to be rewritten. `road-map.yaml` carries the drafting order
that governs when each becomes worth writing.
