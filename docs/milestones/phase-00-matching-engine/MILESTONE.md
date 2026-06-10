# Phase 0 — Matching engine core

**Status:** Draft  
**Owner:**  
**Roadmap ref:** [ROADMAP.md § Phase 0](../../roadmap/ROADMAP.md)  
**Created:** 2026-06-10  
**Target gate date:**

---

## Summary

Build the source-agnostic matching engine: candidate generation → scoring → 1:1 resolution → confidence. Harden with table, property, and fuzz tests on synthetic adversarial data. Biased to **refuse uncertain matches**.

## Dependencies

| Dependency | Status | Notes |
|---|---|---|
| [ADR 0002 Canonical data model](../../adr/0002-canonical-data-model.md) | accepted | `internal/model/` |
| [ADR 0001 Matching engine core](../../adr/0001-record-architecture-decisions.md) | accepted | yes |
| [ADR 0003 Testability-first gates](../../adr/0003-testability-first-gates.md) | accepted | Harness + fuzz required |

## Design constraints

- **C1:** N/A — matcher has no merchant cooperation dependency
- **C2:** Matcher enables fast capture UX (snap receipt → auto-link); indirect moat enabler
- **C3:** Merchant similarity tests include Screwfix, Toolstation, B&Q, fuel descriptors

## Deliverables

| Deliverable | Package / path | Done |
|---|---|---|
| Canonical model (Transaction, Receipt, Match, Money) | `internal/model/` | [x] |
| Matching engine | `internal/match/` | [x] |
| Coverage harness (precision, recall, FMR, conflict rate) | `internal/harness/` | [x] |
| Synthetic scenario generator | `internal/synth/` | [x] |
| Evaluation command | `cmd/evaluate/` | [x] |
| Synthetic fixtures | `test/testdata/matching/` | [x] |
| Makefile test targets | `Makefile` | [x] |

## ADRs required

| ADR | Status | Blocks start? |
|---|---|---|
| [0001](../../adr/0001-record-architecture-decisions.md) | accepted | yes |
| [0002](../../adr/0002-canonical-data-model.md) | accepted | yes |
| [0003](../../adr/0003-testability-first-gates.md) | accepted | yes |

## Test plan

### Unit tests

| Behaviour | Test location | Fixture |
|---|---|---|
| Exact amount + merchant + temporal match | `internal/match/engine_test.go` | inline |
| Tip / amount tolerance | `internal/match/engine_test.go` | inline |
| Refuse wrong amount | `internal/match/engine_test.go` | inline |
| Refuse outside temporal window | `internal/match/engine_test.go` | inline |
| 1:1 resolution under duplicates | `internal/match/engine_test.go` | inline |
| Money tolerance boundaries | `internal/model/money_test.go` | inline |
| Harness precision/recall | `internal/harness/metrics_test.go` | inline |

### Property / fuzz tests

| Invariant | Fuzz target |
|---|---|
| Currency mismatch → no match | `FuzzEngine_NoFalsePositiveOnCurrencyMismatch` |
| Matches respect MinConfidence | `FuzzEngine_MatchedPairsRespectTolerance` |

### Integration / fixture tests

| Scenario | Fixture file |
|---|---|
| Exact pair | `test/testdata/matching/exact_pair.json` |
| Tip tolerance | `test/testdata/matching/tip_tolerance.json` |
| Ambiguous duplicates | `test/testdata/matching/ambiguous_duplicates.json` |

### Harness metrics

| Metric | Target (gate) | Measured |
|---|---|---|
| `Precision` | ≥ 0.99 on labelled synthetic set | — |
| `FalseMatchRate` | < 0.01 | — |
| `Recall` | ≥ 0.80 on labelled synthetic set (acceptable trade-off for precision bias) | — |

## Acceptance criteria

- [ ] High precision on labelled synthetic data (≥ 99%)
- [ ] False-match rate < 1% on adversarial ambiguous fixtures
- [ ] Matcher refuses matches below `MinConfidence` (default 0.75)
- [ ] Temporal skew tolerated: receipt up to 48h before txn, 7d after
- [ ] Amount skew tolerated: ±50p absolute or 2% relative
- [ ] `make test`, `make test-race`, `make test-fuzz` pass

## Kill / pivot criterion

**None** — this phase is cheap and always informative.

## Implementation checklist

- [x] Failing acceptance tests written
- [x] ADRs accepted
- [x] Core implementation
- [x] Harness metrics wired
- [ ] Gate review on full synthetic labelled set
- [ ] `make test` / `make test-race` / `make test-fuzz` pass (verify locally)

## Gate review

| Metric | Threshold | Result | Date |
|---|---|---|---|
| Precision | ≥ 0.99 | — | |
| FalseMatchRate | < 0.01 | — | |
| Recall | ≥ 0.80 | — | |

**Decision:** —

## Risks

| Risk | Mitigation | Phase ref |
|---|---|---|
| Temporal/amount skew breaks matching | Tolerant windows; fuzz tests | Phase 0, 2 |
| False matches erode trust | MinConfidence + precision bias | Phase 0 |
