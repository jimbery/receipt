# Phase 1 validation v3 — gate deferral ratification

**Date:** 2026-06-11  
**Decider:** Jay Imbery (ADR-002)  
**In response to:** validation v3 closure item C3; [validation-response-gate-deferral.md](../../milestones/phase-01-email-ingestion/validation-response-gate-deferral.md)

## Decision

**Ratified: formal deferral (option 2).**

Jay records written rationale to **remove the ≥8-volunteer moat gate as a Phase 1 exit criterion** and **relocate moat measurement to Phase 2**, alongside Open Banking ingestion.

| Field | Value |
|---|---|
| **Decision** | Formal deferral — moat gate transferred to Phase 2 |
| **Effective date** | 2026-06-11 |
| **M1.6 status** | **Done** (fixture-lab engineering deliverables complete) |
| **Moat ownership** | Phase 2 — cohort-weighted purchase coverage, paper-gap quantification, kill/pivot decision tree |

## Rationale

1. **Fixture-lab evidence is sufficient for Phase 1 engineering closure.** CI enforces classifier precision/recall, field accuracy, dedup correctness, determinism, and pattern-based PII audit on committed mail fixtures.
2. **Volunteer cohort is not yet available at gate scale.** Track A recruitment continues in parallel; blocking Phase 1 merge on ≥8 connected mailboxes delays Open Banking work without improving fixture-lab quality.
3. **Accepted cost.** The kill/pivot decision tree fires later; paper-gap quantification is unavailable until volunteer mailboxes exist. Phase 2 engineering proceeds as an explicit risk decision with moat measurement restored as a first-class gate there.

## §4 amendments applied

| Document | Amendment |
|---|---|
| [ADR-002 D7](../../adr/0002-email-ingestion.md) | Moat metric + decision tree **transferred to Phase 2**; Phase 1 D6 documents fixture-lab gate only |
| [M1.6](../../milestones/phase-01-email-ingestion/MILESTONE.md) | Closed **Done** (redefined scope) with "Transferred to Phase 2" appendix |
| [ROADMAP / CHANGELOG](../../roadmap/CHANGELOG.md) | One-paragraph record of move + accepted cost |
| [validation-response-gate-deferral.md](../../milestones/phase-01-email-ingestion/validation-response-gate-deferral.md) | Decision table updated from PENDING → ratified |

## What remains true

- Fixture-lab thresholds ratified 2026-06-10 remain enforced in CI.
- Pattern scrubber + signed manifest govern committed `.eml` fixtures.
- Independent validation v3 (`0a338d1`) verified PII remediation; closure commit addresses C1–C4.
- **External action (Jay):** GitHub Support purge of pre-rebuild SHAs still visible on PR #3 timeline.
