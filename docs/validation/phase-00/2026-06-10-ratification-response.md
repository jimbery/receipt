# Phase 0 validation v2 — ratification response

**Date:** 2026-06-10  
**Responder:** implementation team (Jay)  
**In response to:** [2026-06-10-v2-350dfb6.md](2026-06-10-v2-350dfb6.md) Decisions 1–2

## Decision 1 — Generator composition

**Ratified.**

| Item | Action |
|---|---|
| 37h txn spacing | Documented in `TxnSpacingHours`, ADR-001 amendment, [CHANGELOG](../../roadmap/CHANGELOG.md) |
| Scaled ambiguous / near-duplicate populations | Ambiguous **2%** of n (min 4); near-duplicate **5%** (min 4); refunds **2%** (min 2). Ambiguous capped at 2% because txn+receipt conflict entries double-count toward the ≤5% overall conflict-rate gate — not 7.5% as v2 first suggested |
| `density_stress` permanent class | Hand-built 10-pair scenario in `AllExtended()`; `TestDensityStress_ZeroFalseMatches` at n=10/50/200 |
| Gate re-run | `gate-results.json` @ engine `phase-0.2`, n=10⁴, seed=42 — pass |

## Decision 2 — Signal distinguishability

**Ratified.**

| Item | Action |
|---|---|
| ADR-001 amendment | [0001-record-architecture-decisions.md](../../adr/0001-record-architecture-decisions.md) § Amendment 2026-06-10 |
| Config fields | `DistinguishAmountEpsilon` 0.01, `DistinguishMerchantEpsilon` 0.01, `DistinguishTemporalSecs` 120 |
| Retroactive config tuning | Logged in [CHANGELOG](../../roadmap/CHANGELOG.md) § Phase 0 gate tuning log |

## Gate result (post-ratification)

```
OVERALL recall=0.978 fmr=0.000 conflict_rate=0.043 conflict_ok=1.000 deterministic=true passed=true
config_hash=9b36c6504d7fb524be330b0ed4a0fcbdb2e19b4af898c768b16c7cc45c21a12e
```

See [gate-results.json](../../milestones/phase-00-matching-engine/gate-results.json).
