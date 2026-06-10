# Testing strategy

Testability is a **gate requirement**, not a post-hoc quality step. Every milestone must ship tests before or alongside the code it protects.

## Test pyramid for this project

```
        ┌─────────────┐
        │  Harness    │  cohort-weighted itemisation, match rate, FMR, categorisation accuracy
        │  metrics    │
        ├─────────────┤
        │ Integration │  end-to-end: fixture transactions + receipts → matches
        ├─────────────┤
        │ Property /  │  fuzz: currency mismatch, tolerance bands, ambiguous pairs
        │ fuzz        │
        ├─────────────┤
        │ Table-driven│  known scenarios from roadmap (skew, refusal, 1:1 resolution)
        │ unit tests  │
        └─────────────┘
```

## Requirements by layer

### Unit tests (every package)

- Table-driven tests for all public behaviour
- Name tests after behaviour: `TestEngine_RefusesUncertainMatch`, not `TestEngine2`
- Use `t.Parallel()` where fixtures are independent
- No network, no filesystem (except `test/testdata`)

### Property / fuzz tests (matching engine and parsers)

Mandatory for:

- **Matching engine** — amount/temporal tolerance, currency mismatch, ambiguous duplicate pairs
- **Parsers** (Phase 1+) — malformed input must not panic; unknown merchants fall back gracefully

```go
func FuzzEngine_NoMatchOnCurrencyMismatch(f *testing.F) { ... }
```

Run locally: `make test-fuzz`  
CI runs fuzz in `-fuzztime=10s` short mode.

### Integration tests

- Live under `test/` or behind `//go:build integration`
- Use labelled fixture pairs in `test/testdata/`
- Format: JSON files with `transactions`, `receipts`, `labels`, `expected_metrics`

### Harness metrics

The harness (`internal/harness/`) is the executable form of roadmap gates:

| Metric | Roadmap gate |
|---|---|
| `ItemisationRate` | Phase 1 moat metric |
| `CohortWeightedItemisationRate` | Phase 1 v2 (weighted by merchant mix) |
| `MatchRate` | Phase 2 end-to-end |
| `FalseMatchRate` | Phase 0 precision bias |
| `Precision` / `Recall` | Phase 0, 2 |
| `CategorisationAccuracy` | Phase 1, 6 |

Harness functions must be **deterministic** and callable from tests and CLI.

## Fixtures

```
test/testdata/
  matching/
    exact_pair.json
    tip_tolerance.json
    ambiguous_duplicates.json
  receipts/          # Phase 1+ email/HTML fixtures
  cohort_weights.json  # merchant mix weights for cohort-weighted metrics
```

Fixtures are versioned and referenced by name in milestone acceptance criteria.

## Coverage expectations

- `internal/match/`, `internal/harness/`, `internal/model/`: **≥ 80%** statement coverage
- Parsers: **≥ 90%** on happy path + fallback branches
- Don't chase 100% on trivial getters; do cover all gate logic

```bash
make test-cover   # writes coverage.out + prints summary
```

## Test-first checklist (per milestone)

Copy into each milestone doc:

- [ ] Acceptance criteria translated to failing tests
- [ ] Table tests for named scenarios
- [ ] Fuzz/property tests for invariant boundaries
- [ ] Fixtures added to `test/testdata/` (if integration-level)
- [ ] Harness metrics defined and wired
- [ ] `make test` and `make test-race` pass
- [ ] Formal gate protocol satisfied ([§ below](#formal-gate-protocol)) — scale, determinism, committed `gate-results.json`
- [ ] Gate thresholds documented with actual measured values from that report

## CI commands

GitHub Actions workflow: [`.github/workflows/ci.yml`](../../.github/workflows/ci.yml)

Runs on every push to `main` and on pull requests:

- **Lint** — `golangci-lint` with strict config ([`.golangci.yml`](../../.golangci.yml))
- **Test** — `make test`, `make test-race`, `make test-fuzz`

Local equivalents:

```bash
make lint        # golangci-lint (strict)
make lint-fix    # golangci-lint --fix
make test        # go test ./... -count=1
make test-race   # go test ./... -race -count=1
make test-fuzz   # fuzz targets, short duration
make test-cover  # coverage report
```

## Matcher-specific invariants (Phase 0)

These must **always** hold and are enforced by unit, property, and fuzz tests:

1. Currency mismatch → no match
2. Confidence below `MinConfidence` → no match emitted (except `exact_ref` short-circuit)
3. Each transaction and receipt used at most once in output
4. Indistinguishable equal-confidence candidates → `conflict`, never a silent match
5. Duplicate receipts for one transaction → `conflict` (engine must not guess which copy is real)
6. Engine output is **byte-for-byte stable** across repeated runs (matches, conflicts, unmatched IDs — including order)

---

## Formal gate protocol

> **Lesson (Phase 0 validation):** A green `make test` and a committed `gate-results.json` are not sufficient. The first Phase 0 pass failed independent review because the gate was run at ~10² pairs, determinism was not fully asserted, and local behaviour drifted from the spec to make fixtures pass.

Use this section whenever a milestone claims **Gate review → Done**.

### Before the first gate run

1. **Define all thresholds in writing first** — in the milestone doc, before any evaluation. Never add or relax a threshold after seeing results.
2. **Translate every acceptance criterion into a failing test** — not just a CLI command. If the spec says “equal-confidence candidates never silently resolve”, there must be a `Test…` that fails when that breaks.
3. **Do not change spec behaviour to green tests** — if a fixture expects `conflict` and the engine matches, fix the engine or escalate to ADR/spec review. Do not downgrade the fixture.
4. **Commit a hashable config artifact** — default config JSON + SHA-256 (`config/default-config.json`). Tunings go through config, not hidden code paths.

### Minimum gate run (Phase 0 matching engine)

| Requirement | Value | Why |
|---|---|---|
| Generated pairs | **≥ 10⁴** (`make evaluate-gate`) | FMR ≤ 0.5% is meaningless at n ≪ 10⁴ |
| Seed | Fixed, recorded in report | Reproducibility |
| Config hash | `match.DefaultConfig().Hash()` in report | Detects silent default drift |
| Engine version | String in report | Distinguishes harness from matcher changes |
| Schema version | In `gate-results.json` | Machine-readable contract |

```bash
# Required before claiming Phase 0 gate pass
make test && make test-race && make lint
make evaluate-gate    # n=10⁴ — must exit 0
```

**Re-run at multiple scales** locally before merge (catches sample-size artifacts):

```bash
go run ./cmd/evaluate -generated 200 -seed 42    # must not be the only run
go run ./cmd/evaluate -generated 2000 -seed 42
go run ./cmd/evaluate -generated 10000 -seed 42
go run ./cmd/evaluate -generated 1000 -seed 99  # alternate seed
```

If FMR passes at n=100 but fails at n≥200, the gate has **not** been met.

### What the harness must measure

Phase 0 gate checks **five** thresholds (all must pass):

| Metric | Definition |
|---|---|
| FMR | false matches ÷ matches emitted |
| Recall | correct labelled pairs ÷ labelled pairs |
| Hard-class recall | same, on named adversarial classes |
| Conflict correctness | genuinely ambiguous ÷ conflicts emitted |
| Conflict rate | conflicts emitted ÷ transactions evaluated |
| Determinism | identical full output across repeated runs |

**Anti-patterns — do not ship these:**

| Anti-pattern | Correct approach |
|---|---|
| `resultsEqual` compares only match/conflict **counts** | Compare match fields, conflict IDs, `CompetingIDs`, reasons, and **order** |
| Determinism check runs `-count=5` on hand-built data only | Include ambiguous scenarios; run ≥200 iterations in a dedicated test |
| One generated scenario class (`"generated"`) | Per-class tagging (`generated_tip`, `generated_settlement`, …) for per-class reporting |
| O(T×R) candidate generation | Indexed retrieval (amount buckets, temporal window) — gate must not OOM at 10⁴ |
| Overall gate pass from hand-built fixtures alone | Hand-built + generated suite at protocol scale |

### Synthetic data generator obligations (M0.4)

The generator is part of the gate. A matcher-only test suite cannot catch:

- **Periodic temporal aliasing** (e.g. 1h txn spacing + 48h settlement lag → collisions at scale). Use non-resonant spacing (e.g. prime-hour steps).
- **Relative amount spacing shrinking** as n grows. Steps must stay outside tolerance bands at max scale.
- **Missing populations** — refunds, ambiguous clusters, near-duplicates must exist in generated data, not only hand-built fixtures.

Every generated record must carry a **scenario class** label. `VerifyNoiseBounds()` must be property-tested.

### Spec fidelity checklist

Before marking gate review, confirm:

- [ ] Every fixture row in the milestone outcome table has a scenario + `Expectations` in code
- [ ] Signal distinguishability thresholds in `Config` (ADR-001 amendment 2026-06-10) — not hard-coded
- [ ] `density_stress` scenario passes (zero FMR, all conflict) at n=10/50/200
- [ ] Generator adversarial populations scale with n (ambiguous/near-duplicate ~7.5% each)
- [ ] Duplicate receipts, split-tender, refund refusal, fuel pre-auth covered (Phase 0 table)
- [ ] `MerchantResolver` has table tests **and** a fuzz target
- [ ] Amount bands (tip, partial capture, fuel pre-auth, FX, cashback) have known-value tests
- [ ] `internal/harness/gate.go` thresholds match milestone doc (including `MaxConflictRate`)

### Committed artifacts

A gate pass requires **all** of these in the PR:

1. `docs/milestones/<phase>/gate-results.json` — from `make evaluate-gate`, not hand-edited
2. `docs/milestones/<phase>/MILESTONE.md` — Gate review table filled from that JSON
3. `config/default-config.json` — hash matches report `config_hash`

If an independent reviewer cold-clones the branch and re-runs `make evaluate-gate`, they must get the same pass/fail decision.

### When the gate fails

1. Tune `match.Config` first (ADR D4 — config is data). Record rationale in [roadmap CHANGELOG](../roadmap/CHANGELOG.md) **before** re-running.
2. If structural (resolution, indexing, conflict semantics) — fix engine; do not shrink the dataset.
3. File the full validator report in [docs/validation/](../validation/README.md) (verbatim; versioned filename). Add a ratification response if decisions require sign-off.

---

## Matcher-specific invariants (Phase 0) — enforcement map

| Invariant | Enforced by |
|---|---|
| Currency mismatch | `FuzzEngine_NoFalsePositiveOnCurrencyMismatch` |
| MinConfidence | `TestMatchingScenarios_FromFixtures`, engine table tests |
| 1:1 usage | Resolver + integration scenarios |
| Equal-confidence → conflict | `TestResolver_EqualConfidenceNeverSilentlyResolves` |
| Duplicate receipt → conflict | `TestAllExtended_OutcomeClasses/duplicate_receipt` |
| Determinism (200×) | `TestResolver_Determinism200Iterations`, `TestEngine_Determinism_Ambiguous200Iterations` |
| Noise bounds | `FuzzGenerator_NoiseBounds` |
| Gate thresholds at 10⁴ | CI `make evaluate-gate` |
| Jaro-Winkler / token-set ∈ [0,1] | `FuzzJaroWinkler_Bounded`, `FuzzTokenSetRatio_Bounded` |
