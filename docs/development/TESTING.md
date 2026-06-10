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
- [ ] Gate thresholds documented with actual measured values

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

These must **always** hold and are enforced by fuzz tests:

1. Currency mismatch → no match
2. Confidence below `MinConfidence` → no match emitted
3. Each transaction and receipt used at most once in output
4. Ambiguous duplicate amount/merchant pairs → prefer unmatched over false match
