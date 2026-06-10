# Phase 0 — Matching Engine Core

**Status:** Gate review  
**Owner:**  
**Governing ADR:** [ADR-001 Matching engine core](../../adr/0001-record-architecture-decisions.md)  
**Roadmap ref:** [ROADMAP.md § Phase 0](../../roadmap/ROADMAP.md)  
**Created:** 2026-06-10  
**Target gate date:**

---

## Objective

A source-agnostic transaction↔receipt matching engine, validated against synthetic adversarial data, with gate thresholds defined up front.

**Scope boundary:** no real data sources, no UI, no persistence beyond evaluation. Pure engine + harness.

**Package map:** `internal/match` (matcher core), `internal/synth` (data generation), `internal/harness` (metrics + gate), `cmd/evaluate` (harness CLI), `internal/model` (canonical types).

---

## Gate thresholds (defined before first evaluation run)

These numbers are commitments — adjust only with written rationale *before* a run, never after seeing results.

| Metric | Definition | Threshold |
|---|---|---|
| **Precision / FMR** | false matches ÷ total matches emitted | FMR ≤ 0.5% |
| **Recall** | correct matches ÷ matchable pairs in ground truth | ≥ 85% overall; ≥ 70% on hardest scenario classes |
| **Conflict correctness** | of pairs flagged `conflict`, fraction genuinely ambiguous per ground truth | ≥ 90% |
| **Determinism** | identical output across repeated runs on identical (input, config) | 100%, enforced in CI |

Precision dominates: if tuning trades FMR for recall, FMR wins.

**Hardest scenario classes:** `ambiguous`, `near_duplicate`, `merchant_mangling`, `timezone_shift`, `settlement_delay`

---

## M0.1 — Canonical data model & module skeleton

| Deliverable | Package / path | Done |
|---|---|---|
| Go module skeleton | `go.mod` | [x] |
| Canonical types (`Transaction`, `Receipt`, `LineItem`, `VATLine`, `Match`, `Config`) | `internal/model/` | [x] |
| Matcher core | `internal/match/` | [x] |
| Synthetic data | `internal/synth/` | [x] |
| Evaluation CLI | `cmd/evaluate/` | [x] |
| CI: lint + test + vet + short fuzz + determinism | `.github/workflows/ci.yml` | [x] |
| Nightly long fuzz (non-blocking) | `.github/workflows/fuzz-nightly.yml` | [x] |

**Exit criteria:** module compiles; money invariants in table tests (no floats on money paths); CI green.

---

## M0.2 — Component scorers

| Deliverable | Package / path | Done |
|---|---|---|
| Amount scorer (tolerance bands) | `internal/match/scorer.go` | [x] |
| Temporal scorer (proximity decay) | `internal/match/scorer.go` | [x] |
| `MerchantResolver` interface + basic normaliser | `internal/match/merchant.go` | [x] |
| Jaro-Winkler + token-set similarity | `internal/match/similarity/` | [x] |
| Scorer table + property tests | `internal/match/scorer_test.go` | [x] |
| Fuzz: normaliser / similarity | `internal/match/similarity/` | [x] |

**Exit criteria:** per-scorer table tests; scores ∈ [0,1]; monotonic temporal decay; fuzz clean.

---

## M0.3 — Candidate generation & resolution

| Deliverable | Package / path | Done |
|---|---|---|
| Window/tolerance candidate retrieval | `internal/match/engine.go` | [x] |
| Weighted composite confidence | `internal/match/scorer.go` | [x] |
| Greedy 1:1 + ambiguity-margin conflicts | `internal/match/resolver.go` | [x] |
| `matched \| unmatched \| conflict` outcomes | `internal/model/match.go` | [x] |
| Outcome-class fixtures | `internal/synth/scenarios_extended.go`, `test/testdata/matching/` | [x] |
| Outcome assertion tests | `internal/synth/outcome_test.go` | [x] |

**Fixture coverage:** clean match, tip-adjusted, settlement-delayed, near-duplicate, duplicate receipt, unmatched txn/receipt, genuine ambiguity, merchant mangling, partial capture.

**Exit criteria:** each fixture asserts outcome *class*, not just score.

---

## M0.4 — Synthetic adversarial dataset generator

| Deliverable | Package / path | Done |
|---|---|---|
| Seeded generator with noise profiles | `internal/synth/generator.go` | [x] |
| Scenario-class tagging | `internal/synth/scenario.go`, `Class` enum | [x] |
| Noise-bound property verification | `Dataset.VerifyNoiseBounds()` | [x] |
| Fuzz: generator bounds | `FuzzGenerator_NoiseBounds` | [x] |

**Noise profiles:** timezone shift, settlement lag, tip/partial-capture, FX delta, merchant mangling, unmatched populations.

**Exit criteria:** configurable scale (10²–10⁵ pairs); property-tested noise bounds.

---

## M0.5 — Evaluation harness & gate run

| Deliverable | Package / path | Done |
|---|---|---|
| Gate thresholds + per-class metrics | `internal/harness/gate.go` | [x] |
| JSON + human summary output | `cmd/evaluate/` (`-json`, `-compare`, `-seed`, `-generated`) | [x] |
| Baseline vs candidate comparison | `harness.CompareConfigs()` | [x] |
| Formal gate results (committed) | [gate-results.json](gate-results.json) | [x] |
| Determinism enforced in CI | `TestEngine_Determinism`, CI job | [x] |

**Exit criteria — Phase 0 gate:** all four thresholds met on full adversarial set.

---

## Gate review

| Metric | Threshold | Result | Date |
|---|---|---|---|
| FMR | ≤ 0.5% | 0.000 | 2026-06-10 |
| Overall recall | ≥ 85% | 98.2% | 2026-06-10 |
| Hard-class recall | ≥ 70% | 100% (excl. ambiguous) | 2026-06-10 |
| Conflict correctness | ≥ 90% | 100% | 2026-06-10 |
| Determinism | 100% | pass | 2026-06-10 |

**Decision:** Pass

Full machine-readable report: [gate-results.json](gate-results.json)

---

## Parallel track reminder

**Track A** (customer discovery) runs alongside M0.1–M0.5, not after the gate. Target: 10–15 interviews completed by gate run so technical and WTP signals land together.

---

## Explicitly out of scope

- Real email parsing, Open Banking, OCR (Phases 1, 2, 4)
- Canonical merchant directory (Phase 3 — interface only)
- Persistence, API surface, UI
- Regulatory / licensing work
- ML-based scoring

---

## Kill / pivot criterion

**None** — Phase 0 is cheap and always informative. Gate failure concentrated in merchant-string scenarios quantifies Phase 3 lift required; not a kill signal.

---

## If the gate fails

1. Tune `match.Config` first (ADR D4 — config is data)
2. If structural: merchant similarity (expected — Phase 3) then resolution strategy (global assignment behind same interface)
3. Record rationale in [roadmap CHANGELOG](../../roadmap/CHANGELOG.md) before re-running
