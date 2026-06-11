# Phase 1 — gate deferral (pending decision)

**Date:** 2026-06-10  
**Raised by:** validation v2 (F2)  
**Decider:** Jay Imbery (ADR-002)

## What happened

Commit `3b27128` redefined M1.6 from the volunteer moat gate (≥8 mailboxes, cohort-weighted coverage, kill/pivot decision tree) to a fixture-lab gate only, and marked all milestones **Done**. The ratification document was updated in the same commit series attributing the deferral to Jay Imbery.

**That deferral was proposed by the implementing agent during delivery pressure, not confirmed in this document by Jay before merge.**

## Options (validator's framing)

1. **Hold M1.6 open (recommended default):** M1.1–M1.5 closed on engineering evidence; M1.6 remains **Partial** until Track A volunteers connect and the moat gate runs. Phase 2 engineering may proceed in parallel as an explicit risk decision.
2. **Formal deferral:** Jay records written rationale to move moat measurement to Phase 2, accepting that the kill/pivot tree fires later and `paper-gap` data is unavailable until volunteer mailboxes exist.

## Decision

| Field | Value |
|---|---|
| **Decision** | **PENDING** — Jay to record choice here |
| **Effective date** | — |
| **M1.6 status if (1)** | Partial until volunteer gate |
| **M1.6 status if (2)** | Done (fixture-lab only); new Phase 2 milestone owns moat |

## What is already true regardless

- Fixture-lab thresholds are enforced in CI.
- Engineering remediation from validation v1 is substantively complete.
- The volunteer moat gate protocol in ADR-002 D6/D7 remains the **intended** M1.6 exit unless option (2) is ratified here.
