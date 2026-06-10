# ADR-001: Matching Engine Core Architecture
 
**Status:** Accepted
**Date:** 2026-06-10
**Phase:** 0 — Matching engine core
**Deciders:** Jay
 
---
 
## Context
 
The product's core technical problem is linking payment transactions (transaction-level data only: merchant descriptor, MCC, amount, timestamp) to itemised receipts acquired from separate channels (email parsing, OCR, optionally POS). Card rails do not carry line items, so a matching engine must associate the two sides reliably.
 
Constraints inherited from the roadmap (v2):
 
- **C1 — No merchant-cooperation dependency.** The engine must work on whatever receipt sources exist; POS data is an optional enhancement.
- **C2 — The itemisation rate is the moat.** The engine exists to make multi-source acquisition coherent; it must be source-agnostic so channels can be added without core changes.
- **False matches are worse than no matches.** This is compliance-adjacent data (MTD digital records). A wrong match silently corrupts a tax record; an unmatched transaction is merely a prompt for user action.
Phase 0 builds and validates the engine against synthetic adversarial data, before any real data source exists.
 
## Decision
 
### D1 — Source-agnostic core behind a canonical data model
 
The matcher operates exclusively on canonical `Transaction` and `Receipt` types. All acquisition channels (email, OCR, POS, Open Banking, future card auth feed) normalise into these types upstream. The matcher has no knowledge of source-specific quirks beyond a `Source` enum used for confidence weighting (e.g. POS-sourced receipts carry an exact transaction reference and can short-circuit to an exact match).
 
*Rationale:* channels will be added and removed (constraint C1/C2); the engine must not churn when they do. This also makes the engine a pure function over typed inputs — directly testable with table, property, and fuzz strategies.
 
### D2 — Three-stage pipeline: candidate generation → scoring → resolution
 
1. **Candidate generation** — for each transaction, retrieve receipts within a configurable time window and amount tolerance band. Wide-net recall stage; cheap filters only.
2. **Scoring** — each (transaction, receipt) candidate pair receives a weighted composite confidence score in [0,1] from independent component scores:
   - *Amount score* — tolerance-band based, not equality. Must handle tips, partial captures, cashback, FX rounding. Exact match scores highest; score decays across configured tolerance bands.
   - *Temporal score* — proximity decay across a generous window. Must handle auth-vs-settlement skew (days), timezone offsets, and receipt timestamps that precede or follow the transaction event.
   - *Merchant similarity score* — normalised string similarity (token-based + Jaro-Winkler) between the transaction's merchant descriptor and the receipt's merchant name, after basic prefix stripping (`SQ *`, `PAYPAL *`, store numbers). Full canonical merchant resolution is deliberately deferred to Phase 3; Phase 0 ships a basic normaliser behind an interface so Phase 3 is a drop-in replacement.
   - MCC participates only as a weak corroborating signal, never as a hard filter.
3. **Resolution** — enforce 1:1 matching globally. Greedy assignment by descending confidence with conflict detection: when two candidates for the same transaction (or two transactions for the same receipt) score within an ambiguity margin of each other, the engine emits `conflict` rather than choosing — **unless** the top two candidates are distinguishable on a primary signal (see Amendment 2026-06-10 below). Ambiguity is surfaced, not guessed away.
### D3 — Refusal threshold as a first-class output
 
Every match outcome is one of `matched | unmatched | conflict`, with confidence and method attached. Pairs scoring below a configurable threshold are `unmatched` even if they are the best candidate. The engine is explicitly biased toward refusal: the precision target dominates the recall target (see gate criteria in the milestone document).
 
*Rationale:* the false-match rate is the trust-destroying failure mode for a compliance product. Unmatched items route to user-assisted flows (e.g. OCR capture prompt) which are acceptable UX; silent mismatches are not.
 
### D4 — Determinism and config-as-data
 
All weights, windows, tolerance bands, thresholds, and ambiguity margins live in a single `Config` value with sane defaults. The engine is deterministic for a given (inputs, config) pair — no randomness, no clock reads inside the matcher (time enters only via input data). 
 
*Rationale:* reproducible evaluation runs; tunable against the Phase 1/2 real-data baseline without code changes; property tests can assert invariants across randomly generated configs.
 
### D5 — Money as int64 minor units; time as explicit UTC with original offset retained
 
Amounts are integers in minor units (pence) — no floats anywhere in money paths. Timestamps normalise to UTC for comparison but retain the original offset for diagnostics, since timezone skew is an expected adversarial case.
 
### D6 — Go, standard library only, for the core
 
The engine is a Go module with no third-party dependencies in the core package. Similarity functions (Jaro-Winkler, token-set ratio) are implemented internally. Synthetic data generation and the evaluation harness live in separate packages (`synth`, `cmd/evaluate`) so the core stays a small, pure, fuzzable surface.
 
*Rationale:* maximises testability with Go's native fuzzing; avoids dependency churn in the one component everything else relies on; matches the existing toolchain and testing discipline (table/property/fuzz, as in `bt`).
 
### D7 — Evaluation harness is part of Phase 0, not an afterthought
 
The deliverable includes a labelled synthetic dataset generator producing adversarial cases (timezone-shifted timestamps, tip-adjusted amounts, near-duplicate purchases, forwarded-twice receipts, settlement-delayed transactions) and an evaluation command reporting precision, recall, false-match rate, and conflict rate per scenario class. Gate thresholds are defined in the milestone document *before* the first evaluation run.
 
## Consequences
 
**Positive**
- New acquisition channels (Phases 1, 4, 5) and transaction sources (Phase 2, later card auth) plug in without core changes.
- The core is a pure, deterministic function — cheap to test exhaustively, safe to refactor against a fixed evaluation baseline.
- Refusal bias protects the compliance use case and user trust from day one.
- Phase 3 (merchant normalisation) slots in behind an existing interface with a measurable before/after lift.
**Negative / accepted costs**
- Refusal bias depresses recall initially; more items route to manual/OCR-assisted flows until Phase 3 lifts merchant matching quality. Accepted: recall is recoverable, trust is not.
- Greedy 1:1 resolution is suboptimal versus global assignment (e.g. Hungarian algorithm) in dense ambiguity clusters. Accepted for Phase 0: ambiguity clusters are surfaced as conflicts anyway, and optimal assignment can replace greedy behind the same interface if evaluation shows it matters.
- Internal similarity implementations carry a small correctness burden versus a library. Mitigated by property/fuzz tests against known-value tables.
## Alternatives considered
 
- **ML-based matching from the outset** — rejected for Phase 0. No real training data exists yet; a transparent scored heuristic gives explainable confidence values, and the evaluation harness built here is the prerequisite for any later learned scorer anyway.
- **Hard merchant-string equality after normalisation** — rejected; descriptor mangling makes equality brittle, and Phase 3 exists precisely because this is the hard sub-problem.
- **Allowing 1:N matches (one receipt, multiple transactions)** — deferred. Split-tender payments are real but rare in the target cohort; modelled as a future extension flag rather than core complexity now.
- **Third-party fuzzy-matching / record-linkage libraries** — rejected for the core per D6; revisit only if internal implementations prove inadequate in evaluation.

---

## Amendment 2026-06-10 — Signal distinguishability (ratified)

**Context:** Independent Phase 0 validation (v2) identified that within-margin candidate pairs with clearly different amount, merchant, or temporal evidence (e.g. near-duplicate purchases five minutes apart) should resolve to the better candidate rather than emit `conflict`, while indistinguishable dense clusters (identical merchant, amount, and temporal delta) must still conflict. This refines D2 resolution semantics without weakening the precision bias.

**Decision:** When the top two candidates for a transaction or receipt are within `AmbiguityMargin` on composite confidence, emit `conflict` only if they are **not** distinguishable: amount signal differs by ≤ `DistinguishAmountEpsilon`, merchant by ≤ `DistinguishMerchantEpsilon`, and temporal delta (seconds) by ≤ `DistinguishTemporalSecs`. All three thresholds live in `Config` (defaults 0.01, 0.01, 120).

**Consequences:** Near-duplicate and timezone-shift scenarios can match correctly; density-stress populations (same merchant, amount, minutes apart) route entirely to `conflict` with zero false matches. Tuning these thresholds follows the config-as-data discipline (D4) and must be logged before gate re-runs.

## Amendment 2026-06-10 — Synthetic generator composition (ratified)

**Context:** Validation v1 exposed periodic aliasing when transaction spacing (1h) resonated with 48h settlement lag at scale. Remediation introduced 37h prime spacing between generated transactions — realistic for a sole-trader year spread, but changes what n=10⁴ measures.

**Decision:** Generated transactions use `TxnSpacingHours = 37` (documented in `internal/synth/generator.go`). Adversarial populations scale with n: ambiguous **2%** (minimum 4, all `conflict`); near-duplicate **5%** (minimum 4); refunds **2%** (minimum 2). Ambiguous is capped at 2% because txn+receipt conflict entries double-count toward the ≤5% overall conflict-rate gate. A permanent hand-built `density_stress` scenario class (identical Screwfix pairs, single timestamp) exercises the dense regime the well-separated gate population no longer samples.

**Consequences:** Gate FMR at n=10⁴ predominantly measures well-separated pairs; density safety is asserted by the `density_stress` class and `TestDensityStress_ZeroFalseMatches`. Future generator composition changes require CHANGELOG entry before gate re-commit.
