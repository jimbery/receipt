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

_Fill when status = Gate review. Complete [TESTING.md § Formal gate protocol](../../development/TESTING.md#formal-gate-protocol) first._

### Pre-merge gate checklist

- [ ] All thresholds defined **before** first `make evaluate-gate` run (never adjusted post-hoc)
- [ ] Every acceptance criterion has a corresponding automated test (not CLI-only)
- [ ] Formal gate run at **protocol scale** (Phase 0: n≥10⁴ via `make evaluate-gate`)
- [ ] Spot-check at n=200 and n=2000 — pass must not be a sample-size artifact
- [ ] `gate-results.json` committed from that run (seed, config hash, scale, schema version populated)
- [ ] `config/default-config.json` hash matches report
- [ ] Determinism: full output equality tested (including conflict order), not counts only
- [ ] No fixture/spec downgrades to green tests without ADR

| Metric | Threshold | Result | Date |
|---|---|---|---|
| | | | |

**Decision:** Pass / Pivot / Kill

**Notes:**

**Machine-readable report:** [gate-results.json](gate-results.json)

## Risks

| Risk | Mitigation | Phase ref |
|---|---|---|
| | | |
