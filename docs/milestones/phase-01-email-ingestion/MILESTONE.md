# Phase 1 Milestones — Email Receipt Ingestion (v1.0)

**Status (post–validation v3, ratified):**

| Milestone | State | Notes |
|---|---|---|
| **M1.1** | **Done** | Maildir + Gmail/Graph pilot fetch (token file + httptest); pattern scrubber + manifest; consent + recruitment docs |
| **M1.2** | **Done** | 560-message frozen corpus; per-pattern classifier table tests; hold-out negatives; ParseRFC822 fuzz; smoke in CI |
| **M1.3** | **Done** | Extractor registry + fallback chain; grade/canonicalise tests; SA category table; OCR counter; HTML/PDF fuzz |
| **M1.4** | **Done** | Priority extractors + scrubbed fixtures (`.txt` + re-scrubbed `.eml`); field-accuracy tests |
| **M1.5** | **Done** | Milestone fixture table + provenance retention (`MergedFrom`); dedup→matcher conflict loop |
| **M1.6** | **Done** | Fixture-lab harness + pattern scrubber + signed manifest. Moat gate **transferred to Phase 2** — see appendix |

CI runs **`make evaluate-ingest-smoke`** on the synthetic corpus (expect high scores) plus hold-out fixtures and pattern-based PII audit on all committed mail under `test/testdata/email/`.

**Objective:** turn mailboxes into canonical, graded, deduplicated `Receipt` records with zero per-receipt user action. **Moat measurement** (cohort-weighted purchase coverage, kill/pivot tree) is **Phase 2 scope** as ratified 2026-06-11.

**Governing ADR:** [ADR-002](../../adr/0002-email-ingestion.md) (v1.0); ADR-001 (v1.2) for the canonical model and downstream contract.
**Scope boundary:** no transaction sources (Phase 2), no OCR (Phase 4 — image inputs route to `requires_ocr` and are counted), no UI beyond volunteer consent/connection flow, no production OAuth verification (pilot modes only).
**Parallel dependency:** Track A volunteers supply scrubbed fixtures and the cohort merchant-mix weight vector. Phase 2 engineering proceeds with moat measurement as an explicit Phase 2 gate.

---

## Gate thresholds (ratified 2026-06-10 — fixture-lab subset)

**Ratification:** [validation/phase-01/2026-06-10-ratification-thresholds.md](../../validation/phase-01/2026-06-10-ratification-thresholds.md). **Moat thresholds** remain committed but **deferred to Phase 2 measurement** — [gate deferral ratification](../../validation/phase-01/2026-06-11-ratification-gate-deferral.md).

**Gate deferral:** [validation-response-gate-deferral.md](validation-response-gate-deferral.md) — **ratified 2026-06-11** (option 2).

### Fixture-lab gate (Phase 1 exit — enforced in CI)

| Metric | Definition | Threshold | Enforced by |
|---|---|---|---|
| **Classifier precision** (receipt-bearing) | receipt-bearing verdicts that are correct | ≥ **97%** | `make evaluate-ingest-smoke` |
| **Classifier recall** (receipt-bearing) | labelled receipt-bearing messages found | ≥ **92%** | `make evaluate-ingest-smoke` |
| **Extraction field accuracy** (priority merchants) | supplier + total on labelled fixtures | ≥ **98%** | field-accuracy tests |
| **Dedup correctness** | declared outcomes in M1.5 fixture table | **zero** false merges | table + integration tests |
| **Determinism** | identical pipeline output for identical inputs | 100% | `pipeline_determinism_test` |
| **Fixture PII audit** | no pattern PII in committed mail | pass | `TestScrubber_NoPIIInCommittedFixtures` + manifest |

---

## M1.6 — Measurement harness & gate run

**Deliverables (done):** fixture-lab evaluator (`ingest-fixture-smoke-v2`); hold-out classifier fixtures; pattern scrubber + signed manifest for `.eml` fixtures; pilot blocklist loaded from gitignored `emls/pilot-blocklist.json` at scrub time.

**Exit criteria:** fixture-lab gate passes in CI; independent validation v3 remediation verified.

---

## Appendix — Transferred to Phase 2

The following were originally M1.6 exit criteria under ADR-002 D7. Jay ratified their transfer to Phase 2 on 2026-06-11:

| Metric | Threshold | Phase 2 owner |
|---|---|---|
| Moat combined-channel projection | ≥ **70%** | Phase 2 moat gate |
| Moat email-channel coverage | ≥ **80%** | Phase 2 moat gate |
| Paper-gap / kill-pivot tree | operational branch | Phase 2 decision tree |

**Accepted cost:** kill/pivot tree fires after Open Banking ingestion begins; paper-gap data unavailable until volunteer mailboxes exist at scale.

---

## Explicitly out of scope for Phase 1

Transaction feeds and end-to-end matching on real data (Phase 2); OCR/photo capture (Phase 4 — but `requires_ocr` counts are a Phase 1 deliverable); LLM extraction (deferred per ADR-002); merchant directory (Phase 3); MTD output/export (Phase 6); production OAuth verification.

## Carried-over contracts this phase must honour

ADR-001's duplicate-receipt conflict behaviour; Phase 0 cohort descriptor corpus unchanged; **no PII in committed fixtures** (pattern audit + manifest sign-off).
