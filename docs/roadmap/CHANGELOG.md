# Roadmap changelog

Record material changes to strategy, gates, or sequencing here.

## Format

```markdown
## YYYY-MM-DD — vX.Y summary title

**Trigger:** what caused the change (data, interview, gate failure, ADR)
**Changes:**
- bullet
**Milestones affected:** phase-XX-…
```

---

## 2026-06-10 — v2 initial

**Trigger:** Post-viability analysis; MTD mandate cohort reframing.

**Changes:**
- Target customer: MTD-mandated sole traders and small landlords
- OCR promoted above POS integration (Phases 4/5 swapped)
- Itemisation rate elevated to moat metric with kill/pivot thresholds
- Customer discovery (Track A) runs parallel to Phase 0–2
- Design constraint C1: no merchant-cooperation dependency for core value

**Milestones affected:** all

---

## 2026-06-10 — Phase 0 gate tuning log (ratified post validation v2)

**Trigger:** Independent validation v2 conditional pass; Decisions 1–2 require written ratification before Phase 0 complete.

**Config changes** (`match.DefaultConfig`, before gate re-commit):

| Parameter | Prior | Committed | Rationale |
|---|---|---|---|
| `MinConfidence` | 0.75 | 0.82 | Reduce false matches at scale; precision bias |
| `AmbiguityMargin` | 0.03 | 0.02 | Tighter ambiguity detection |
| `MinMerchantForCandidate` | — | 0.55 | Filter cross-merchant false candidates in dense index |
| `DistinguishAmountEpsilon` | — | 0.01 | ADR amendment: resolve distinguishable near-duplicates |
| `DistinguishMerchantEpsilon` | — | 0.01 | ADR amendment |
| `DistinguishTemporalSecs` | — | 120 | ADR amendment: 3-min spacing still conflicts; 5-min+ resolves |

**Generator changes** (ratified):

- `TxnSpacingHours = 37` — breaks 48h settlement periodic aliasing; sole-trader year spread
- Ambiguous population scales at 2% of n (min 4); near-duplicate at 5%; refunds 2% — capped because txn+receipt conflict entries double-count toward the ≤5% overall conflict-rate gate
- Permanent `density_stress` scenario class (10 pairs in gate; tested to n=200)

**Milestones affected:** phase-00-matching-engine

---

## 2026-06-10 — Phase 0 complete (validation v3)

**Trigger:** Independent validation v3 PASS @ `2bcc071`; close residuals 1–2.

**Changes:**
- Fuzz targets on `JaroWinkler` and `TokenSetRatio` (`similarity/fuzz_test.go`); wired in `make test-fuzz` and nightly workflow
- ADR-001 generator amendment corrected: ambiguous population **2%** (matches code)
- Phase 0 milestone status → **Done**; v3 report filed in `docs/validation/`

**Milestones affected:** phase-00-matching-engine
