# Phase 1 Milestones — Email Receipt Ingestion (v1.0)

**Status (post–validation v2, honest):**

| Milestone | State | Notes |
|---|---|---|
| **M1.1** | **Done** | Maildir + Gmail/Graph pilot fetch (token file + httptest); pattern scrubber + manifest; consent + recruitment docs |
| **M1.2** | **Done** | 560-message frozen corpus; per-pattern classifier table tests; hold-out negatives; ParseRFC822 fuzz; smoke in CI |
| **M1.3** | **Done** | Extractor registry + fallback chain; grade/canonicalise tests; SA category table; OCR counter; HTML/PDF fuzz |
| **M1.4** | **Done** | Priority extractors + scrubbed fixtures (`.txt` + re-scrubbed `.eml`); field-accuracy tests |
| **M1.5** | **Done** | Milestone fixture table + provenance retention (`MergedFrom`); dedup→matcher conflict loop |
| **M1.6** | **Partial** | **Pipeline + fixture-lab harness done.** Volunteer moat gate (≥8 mailboxes, paper-gap decision tree) **pending** — see [validation response](validation-response-gate-deferral.md) |

CI runs **`make evaluate-ingest-smoke`** on the synthetic corpus (expect high scores) plus hold-out fixtures and pattern-based PII audit on all committed mail under `test/testdata/email/`.

**Objective:** turn mailboxes into canonical, graded, deduplicated `Receipt` records with zero per-receipt user action. **Moat measurement** (cohort-weighted purchase coverage, kill/pivot tree) remains the M1.6 exit criterion and is **not closed** until volunteer ground truth exists or Jay records an explicit scope decision.

**Governing ADR:** [ADR-002](../../adr/0002-email-ingestion.md) (v1.0); ADR-001 (v1.2) for the canonical model and downstream contract.
**Scope boundary:** no transaction sources (Phase 2), no OCR (Phase 4 — image inputs route to `requires_ocr` and are counted), no UI beyond volunteer consent/connection flow, no production OAuth verification (pilot modes only).
**Parallel dependency:** Track A volunteers supply scrubbed fixtures and the cohort merchant-mix weight vector. Phase 2 engineering may start in parallel as an explicit risk decision; the moat gate is not waived by default.

---

## Gate thresholds (ratified 2026-06-10 — fixture-lab subset only)

**Ratification:** [validation/phase-01/2026-06-10-ratification-thresholds.md](../../validation/phase-01/2026-06-10-ratification-thresholds.md). **Moat thresholds** remain committed but **unmeasured** until M1.6 volunteer gate runs.

**Gate deferral:** [validation-response-gate-deferral.md](validation-response-gate-deferral.md) — **pending Jay Imbery decision** on whether to defer the volunteer gate or hold M1.6 open.

### Fixture-lab gate (engineering evidence — not M1.6 complete)

| Metric | Definition | Threshold | Enforced by |
|---|---|---|---|
| **Classifier precision** (receipt-bearing) | receipt-bearing verdicts that are correct | ≥ **97%** | `make evaluate-ingest-smoke` |
| **Classifier recall** (receipt-bearing) | labelled receipt-bearing messages found | ≥ **92%** | `make evaluate-ingest-smoke` |
| **Extraction field accuracy** (priority merchants) | supplier + total on labelled fixtures | ≥ **98%** | field-accuracy tests |
| **Dedup correctness** | declared outcomes in M1.5 fixture table | **zero** false merges | table + integration tests |
| **Determinism** | identical pipeline output for identical inputs | 100% | `pipeline_determinism_test` |
| **Fixture PII audit** | no pattern PII in committed mail | pass | `TestScrubber_NoPIIInCommittedFixtures` + manifest |

### Volunteer moat gate (M1.6 exit — pending)

| Metric | Threshold | Status |
|---|---|---|
| Moat combined-channel projection | ≥ **70%** | Not measured |
| Moat email-channel coverage | ≥ **80%** | Not measured |
| Paper-gap / kill-pivot tree | operational branch | Not run |

---

## M1.6 — Measurement harness & gate run

**Deliverables (done):** fixture-lab evaluator (`ingest-fixture-smoke-v2`); hold-out classifier fixtures; pattern scrubber + signed manifest for `.eml` fixtures.

**Deliverables (pending):** volunteer ground-truth labelling; formal moat gate run; decision-tree branch documented.

**Exit criteria:** volunteer gate executed **or** written scope decision by Jay; independent validation before "Phase 1 complete."

---

## Explicitly out of scope for Phase 1

Transaction feeds and end-to-end matching on real data (Phase 2); OCR/photo capture (Phase 4 — but `requires_ocr` counts are a Phase 1 deliverable); LLM extraction (deferred per ADR-002); merchant directory (Phase 3); MTD output/export (Phase 6); production OAuth verification.

## Carried-over contracts this phase must honour

ADR-001's duplicate-receipt conflict behaviour; Phase 0 cohort descriptor corpus unchanged; **no PII in committed fixtures** (pattern audit + manifest sign-off).
