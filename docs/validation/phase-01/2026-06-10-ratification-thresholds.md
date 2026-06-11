# Phase 1 gate thresholds — ratification

**Date:** 2026-06-10  
**Ratified by:** Jay Imbery (decider, ADR-002) — **fixture-lab subset only**  
**Status:** Accepted for CI fixture-lab thresholds. **Moat gate deferral pending** — see [validation-response-gate-deferral.md](../../milestones/phase-01-email-injestion/validation-response-gate-deferral.md).

## Context

Phase 1 gate numbers were proposed in [MILESTONE.md](../../milestones/phase-01-email-injestion/MILESTONE.md) with **`[SET: Jay]`** markers. Jay ratified the **numeric thresholds** below on 2026-06-10.

**Not ratified in this document:** redefining M1.6 to exclude the volunteer moat gate. That scope change requires an explicit entry in [validation-response-gate-deferral.md](../../milestones/phase-01-email-injestion/validation-response-gate-deferral.md).

## Ratified thresholds (fixture-lab — enforced in CI)

| Metric | Threshold | Enforced by |
|---|---|---|
| **Classifier precision** (receipt-bearing) | ≥ **97%** | `make evaluate-ingest-smoke` |
| **Classifier recall** (receipt-bearing) | ≥ **92%** | `make evaluate-ingest-smoke` |
| **Extraction field accuracy** — total / date / supplier | ≥ **98%** | field-accuracy tests |
| **Dedup false merges** | **0** | M1.5 fixture table |
| **Determinism** | **100%** | `pipeline_determinism_test` |
| **Fixture PII audit** | pass | pattern audit + manifest |

## Committed thresholds (volunteer moat — not yet measured)

| Metric | Threshold |
|---|---|
| **Moat: combined-channel projection** | ≥ **70%** |
| **Moat: email-channel coverage** | ≥ **80%** |
| **Extraction field accuracy** — line items / VAT | ≥ **95%** |
| **Categorisation accuracy** | ≥ **90%** |
| **Dedup missed merges** | ≤ **5%** |

These remain the M1.6 volunteer gate bar per ADR-002 until a written deferral is recorded.

## Cohort weight vector

Provisional MTD merchant-mix weights in `test/testdata/cohort_weights.json`.

## What this unblocks

- CI fixture-lab gate and M1.1–M1.5 engineering claims.
- **Does not** unblock "Phase 1 complete" or moat/kill-pivot claims without M1.6 volunteer gate or formal deferral.
