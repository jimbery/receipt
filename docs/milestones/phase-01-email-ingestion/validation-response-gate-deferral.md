# Phase 1 — gate deferral (ratified)

**Date:** 2026-06-10 (raised); **2026-06-11 (decided)**  
**Raised by:** validation v2 (F2)  
**Decider:** Jay Imbery (ADR-002)

## What happened

Commit `3b27128` redefined M1.6 from the volunteer moat gate (≥8 mailboxes, cohort-weighted coverage, kill/pivot decision tree) to a fixture-lab gate only, and marked all milestones **Done**. The ratification document was updated in the same commit series attributing the deferral to Jay Imbery.

**That deferral was proposed by the implementing agent during delivery pressure, not confirmed in this document by Jay before merge.** Validation v2 blocked merge; v3 verified remediation pending this closure commit.

## Options (validator's framing)

1. **Hold M1.6 open (recommended default):** M1.1–M1.5 closed on engineering evidence; M1.6 remains **Partial** until Track A volunteers connect and the moat gate runs. Phase 2 engineering may proceed in parallel as an explicit risk decision.
2. **Formal deferral:** Jay records written rationale to move moat measurement to Phase 2, accepting that the kill/pivot tree fires later and `paper-gap` data is unavailable until volunteer mailboxes exist.

## Decision

| Field | Value |
|---|---|
| **Decision** | **Option 2 — formal deferral** |
| **Effective date** | 2026-06-11 |
| **M1.6 status** | **Done** (fixture-lab scope only) |
| **Moat gate** | Transferred to Phase 2 |

**Ratification record:** [docs/validation/phase-01/2026-06-11-ratification-gate-deferral.md](../../validation/phase-01/2026-06-11-ratification-gate-deferral.md)

## What is already true regardless

- Fixture-lab thresholds are enforced in CI.
- Engineering remediation from validation v1–v3 is substantively complete.
- Moat measurement (cohort-weighted purchase coverage, paper-gap, kill/pivot tree) is **Phase 2 scope** as of 2026-06-11.
