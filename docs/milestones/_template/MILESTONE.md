# [Phase XX] Milestone title

**Status:** Draft  
**Owner:**  
**Roadmap ref:** [ROADMAP.md § Phase X](../../roadmap/ROADMAP.md)  
**Created:** YYYY-MM-DD  
**Target gate date:**

---

## Summary

One paragraph: what this milestone delivers and why it exists in the sequence.

## Dependencies

| Dependency | Status | Notes |
|---|---|---|
| Prior milestone / ADR / external | | |

## Design constraints

Which of C1 / C2 / C3 apply? How does this milestone respect them?

- [ ] C1 — No merchant-cooperation dependency for core value
- [ ] C2 — Effect on itemisation rate (moat metric) stated
- [ ] C3 — Prioritised for MTD cohort merchant mix

## Deliverables

| Deliverable | Package / path | Done |
|---|---|---|
| e.g. Matcher engine | `internal/match/` | [ ] |

## ADRs required

| ADR | Status | Blocks start? |
|---|---|---|
| [0002 Canonical data model](../../adr/0002-canonical-data-model.md) | accepted | yes |

## Test plan

> **Required before status → Ready.** See [TESTING.md](../../development/TESTING.md).

### Unit tests

| Behaviour | Test location | Fixture |
|---|---|---|
| | `internal/.../_test.go` | |

### Property / fuzz tests

| Invariant | Fuzz target |
|---|---|
| | `Fuzz...` |

### Integration / fixture tests

| Scenario | Fixture file |
|---|---|
| | `test/testdata/...` |

### Harness metrics

| Metric | Target (gate) | Measured |
|---|---|---|
| e.g. `FalseMatchRate` | < 1% | — |

## Acceptance criteria

Copied from roadmap gate. All must pass for **Done**.

- [ ] Criterion 1 (measurable)
- [ ] Criterion 2

## Kill / pivot criterion

From roadmap. If met, stop and update [roadmap CHANGELOG](../../roadmap/CHANGELOG.md).

## Implementation checklist

- [ ] Failing acceptance tests written
- [ ] ADRs accepted
- [ ] Core implementation
- [ ] Harness metrics wired
- [ ] `make test` / `make test-race` / `make test-fuzz` pass
- [ ] Gate review recorded below

## Gate review

_Fill when status = Gate review._

| Metric | Threshold | Result | Date |
|---|---|---|---|
| | | | |

**Decision:** Pass / Pivot / Kill

**Notes:**

## Risks

| Risk | Mitigation | Phase ref |
|---|---|---|
| | | |
