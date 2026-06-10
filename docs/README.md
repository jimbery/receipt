# Receipt Capture — Documentation

This directory is the **source of truth for what gets built**. Code follows docs, not the other way around.

## How development is triggered

```
ROADMAP (strategy)
    ↓ defines phases & gates
MILESTONE (execution unit)
    ↓ defines deliverables, tests, acceptance criteria
ADR (decisions)
    ↓ records non-obvious choices that constrain implementation
CODE + HARNESS
    ↓ proves the gate
MILESTONE status → done
```

| When you want to… | Start here |
|---|---|
| Change strategy or sequencing | [`roadmap/ROADMAP.md`](roadmap/ROADMAP.md) |
| Kick off or track a build phase | [`milestones/`](milestones/) — copy `_template/MILESTONE.md` |
| Record a technical decision | [`adr/template.md`](adr/template.md) — next number in sequence |
| Understand testing expectations | [`development/TESTING.md`](development/TESTING.md) |
| See the end-to-end workflow | [`development/WORKFLOW.md`](development/WORKFLOW.md) |
| Read independent gate validation reports | [`validation/`](validation/README.md) |

## Directory layout

```
docs/
├── roadmap/          # Product & technical strategy (versioned)
├── milestones/       # Phase execution docs — each triggers a build cycle
│   └── _template/    # Copy this to start a new milestone
├── adr/              # Architecture Decision Records
├── validation/       # Independent validator reports (verbatim, versioned)
└── development/      # Engineering conventions (testing, workflow)
```

## Milestone lifecycle

1. **Draft** — milestone doc created from template; gate criteria and test plan defined *before* coding.
2. **Ready** — dependencies met, ADRs for open questions resolved, acceptance tests sketched.
3. **In progress** — implementation; every deliverable maps to a package and test file.
4. **Gate review** — harness metrics run against criteria in the milestone doc.
5. **Done** — gate passed (or kill/pivot decision recorded in roadmap changelog).

A milestone does not move to **In progress** until its **Test plan** section is complete.

## ADR lifecycle

ADRs are numbered sequentially (`0001-…`, `0002-…`). Status: `proposed` → `accepted` → `superseded` | `rejected`.

Create an ADR when a decision:
- is hard to reverse,
- affects multiple packages, or
- needs to be understood by someone who wasn't in the room.

## Active work

| Milestone | Status | Gate metric |
|---|---|---|
| [Phase 0 — Matching engine](milestones/phase-00-matching-engine/MILESTONE.md) | Done | Precision / FMR on synthetic data |
| [Track A — Customer discovery](milestones/track-a-customer-discovery/MILESTONE.md) | Draft | Willingness-to-pay evidence |
