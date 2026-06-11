# Receipt Capture — Technical Development Roadmap (v2)

> **Status:** Active  
> **Last updated:** 2026-06-10  
> **Changelog:** [CHANGELOG.md](CHANGELOG.md)

**Focus:** solving the core technical limitation — reliably deriving an *itemised receipt* from a *payment transaction*.

**v2 changes (post viability analysis):**
1. Target customer reframed: **MTD-mandated sole traders and small landlords first**, not employee expense tracking.
2. Customer discovery now runs as a **parallel track alongside Phase 1**, not a sequential pre-step.
3. The itemisation rate is elevated from a validation gate to **the moat metric**, with explicit kill/pivot thresholds.
4. **OCR promoted above POS integration** (Phases 4 and 5 swapped) — the MTD cohort lives on paper trade-counter and fuel receipts.
5. Explicit kill criteria added per gate; the Flux structural lesson (never depend on merchant cooperation for core value) is now a stated design constraint.

---

## Milestone index

Each phase links to its execution doc. **Implementation starts from the milestone, not this file.**

| Phase | Milestone doc | Gate |
|---|---|---|
| Track A | [track-a-customer-discovery](../milestones/track-a-customer-discovery/MILESTONE.md) | Willingness to pay |
| Phase 0 | [phase-00-matching-engine](../milestones/phase-00-matching-engine/MILESTONE.md) | Precision / FMR on synthetic data |
| Phase 1 | [phase-01-email-ingestion](../milestones/phase-01-email-ingestion/MILESTONE.md) | Fixture-lab ingest gate (moat → Phase 2) |
| Phase 2 | [phase-02-open-banking](../milestones/phase-02-open-banking/MILESTONE.md) | End-to-end match rate + moat measurement |
| Phase 3 | [phase-03-merchant-normalisation](../milestones/phase-03-merchant-normalisation/MILESTONE.md) | Match rate lift vs Phase 2 |
| Phase 4 | [phase-04-ocr-capture](../milestones/phase-04-ocr-capture/MILESTONE.md) | Combined itemisation moat |
| Phase 5 | [phase-05-pos-integration](../milestones/phase-05-pos-integration/MILESTONE.md) | Optional coverage lift |
| Phase 6 | [phase-06-mtd-output](../milestones/phase-06-mtd-output/MILESTONE.md) | Pilot quarterly cycle |

---

## 1. The core problem, stated precisely

Card payment rails carry **transaction-level data only**:

- A merchant descriptor string (often mangled, e.g. `TESCO STORES 3344 LONDN`)
- An MCC (merchant category code)
- Amount, currency, timestamp
- Account / transaction identifiers

They do **not** carry line items. Level 3 / itemised data exists in the card schemes but is effectively absent from general retail. So the basket contents must be acquired from a **separate channel** and then **matched back** to the transaction.

Two stacked technical problems:

1. **Acquisition** — obtain a structured, itemised receipt (supplier, line items, per-line VAT, supplier VAT number, total) from somewhere other than the card rail.
2. **Matching** — link that receipt to the correct transaction event with high confidence and a near-zero false-match rate.

The matching engine is the heart of the system. Acquisition sources are pluggable inputs, ordered by data quality, infrastructure cost, **and fit to the target cohort**.

## 2. Design constraints (new in v2)

**C1 — No merchant-cooperation dependency.** Flux shut down in 2022 with over a million users because its model required per-merchant POS integration before delivering value — a two-sided network it couldn't complete. Core value here must be deliverable to customer one with zero merchant involvement. POS integration is an *enhancement* channel only, never a dependency.

**C2 — The itemisation rate is the moat, not just a gate.** Pleo, Soldo, Dext, Hubdoc, Expensify and the accounting incumbents all capture transactions; differentiation rests entirely on automatic line-item coverage without user action. If coverage is mediocre, this is a feature, not a company. Every phase decision should be evaluated by its effect on this number for the target cohort.

**C3 — Build for the MTD cohort's actual spending.** From April 2026, sole traders and landlords over £50k must keep digital records of every transaction, with thresholds falling to £30k (2027) and £20k (2028) — roughly 864k mandated taxpayers in wave one, most not yet signed up. Their spend skews to trade suppliers (Screwfix, Toolstation, B&Q, builders' merchants), fuel, Amazon, and subscriptions — *not* general consumer retail. Parser and merchant-directory effort is prioritised against this cohort's merchant mix.

## 3. Target architecture

```
   ┌─────────────────────────────────────────────┐
   │              MATCHING ENGINE                  │
   │  candidate generation → scoring → resolution  │
   │       (confidence score per match)            │
   └───────────────▲───────────────▲──────────────┘
                   │               │
        ┌──────────┘               └──────────┐
   TRANSACTION SOURCES            RECEIPT SOURCES
   - Open Banking feed            - Email ingestion (parse)
   - (later) own card auth        - OCR / photo capture
                                  - POS integration (optional)
                   │               │
                   └──────┬────────┘
                          ▼
              CANONICAL DATA MODEL
        (normalised transactions + receipts)
                          ▼
              COVERAGE / QUALITY HARNESS
        (itemisation rate, match accuracy, FMR)
                          ▼
              MTD-SHAPED OUTPUT
   (date, amount, category per HMRC digital-record rules;
    export path into MTD-compatible software)
```

New in v2: the output of the pipeline is explicitly shaped to HMRC's digital-record requirements (date, amount, category per transaction, digitally linked from point of entry). This is what the customer is actually buying.

## 4. Canonical data model

As v1 (Transaction / Receipt / Match), with one addition:

**Receipt** gains `Category string` — an expense category mapped to Self Assessment categories, since MTD digital records follow the existing SA categorisation. Auto-categorisation (MCC + merchant + line items → SA category) becomes a first-class output, not an afterthought.

→ See `internal/model/` and [ADR 0002](../adr/0002-canonical-data-model.md).

## 5–9. Phases, harness, risks, sequence

Full phase detail lives in individual [milestone docs](../milestones/). Risk register and deferred items are copied into each relevant milestone's **Risks** section.

**Sequence (v2):**

1. Phase 0 + Track A (parallel)
2. Phase 1 → fixture-lab ingest gate
3. Phase 2 → Open Banking + moat measurement (transferred from M1.6, ratified 2026-06-11)
4. Phase 3 → merchant normalisation
5. Phase 4 → OCR
6. Phase 5 → POS (optional)
7. Phase 6 → MTD output

The front half remains buildable with no licence, no card, and no merchant deals.
