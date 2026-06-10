# Milestones

Each subdirectory is one **execution unit** that triggers a development cycle.

## Start a new milestone

```bash
cp -r docs/milestones/_template docs/milestones/phase-XX-short-name
# Edit MILESTONE.md — complete Test plan before setting status to Ready
```

## Status values

| Status | Meaning |
|---|---|
| `Draft` | Scope and gates being defined |
| `Ready` | Test plan complete; ADRs accepted; dependencies met |
| `In progress` | Active implementation |
| `Gate review` | Code complete; formal gate run at protocol scale (see [TESTING.md § Formal gate protocol](../development/TESTING.md#formal-gate-protocol)) |
| `Done` | Gate passed; `gate-results.json` committed; independent re-run would agree |
| `Killed` | Kill/pivot criterion met; see roadmap CHANGELOG |

**Done requires more than green unit tests.** Phase 0 gate pass specifically needs n≥10⁴ generated pairs, five thresholds, and reproducible metadata in the committed report.

## Index

| Milestone | Status |
|---|---|
| [Track A — Customer discovery](track-a-customer-discovery/MILESTONE.md) | Draft |
| [Phase 0 — Matching engine](phase-00-matching-engine/MILESTONE.md) | Done |
| [Phase 1 — Email ingestion](phase-01-email-ingestion/MILESTONE.md) | Draft |
| [Phase 2 — Open Banking](phase-02-open-banking/MILESTONE.md) | Draft |
| [Phase 3 — Merchant normalisation](phase-03-merchant-normalisation/MILESTONE.md) | Draft |
| [Phase 4 — OCR capture](phase-04-ocr-capture/MILESTONE.md) | Draft |
| [Phase 5 — POS integration](phase-05-pos-integration/MILESTONE.md) | Draft |
| [Phase 6 — MTD output](phase-06-mtd-output/MILESTONE.md) | Draft |
