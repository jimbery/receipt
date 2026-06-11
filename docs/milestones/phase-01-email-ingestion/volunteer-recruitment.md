# Volunteer recruitment (M1.1 deliverable)

**Status:** Kicked off — 2026-06-10  
**Owner:** Track A (customer discovery)  
**Feeds:** scrubbed fixtures for corpus growth (`make scrub-emls`); cohort weight refresh for Track A

## Objective

Recruit UK sole-trader / small-business volunteers who match the provisional MTD cohort merchant mix (`test/testdata/cohort_weights.json`) and will connect a personal mailbox for the Phase 1 pilot.

## Target cohort

| Segment | Target count | Mailbox |
|---|---|---|
| Trades / hardware-heavy (Screwfix, Toolstation, B&Q) | 3–4 | Gmail or Microsoft |
| E-commerce / mixed (Amazon-heavy) | 2–3 | Gmail or Microsoft |
| Fuel + utilities / subscriptions | 2–3 | Gmail or Microsoft |
| Long-tail / marketing noise (hard negatives) | 1–2 | Any |

**Total:** 8–12 volunteers when Track A scales outreach; Phase 1 gate does not require a labelled cohort.

## Process

1. **Outreach** — Track A shortlist from network + accountant referrals; no paid ads in Phase 1.
2. **Consent** — Share [volunteer-data-handling.md](../../consent/volunteer-data-handling.md); volunteer confirms checklist before any token is issued.
3. **Connection** — Volunteer runs pilot OAuth locally per [pilot-mail-oauth.md](../../development/pilot-mail-oauth.md); tokens stay on their machine (`~/receipt-pilot/`).
4. **Fixture export** — `make scrub-emls` from gitignored `emls/` into `test/testdata/email/merchants/`.
5. **Withdrawal** — Documented in consent doc; cohort table updated on exit.

## Repo artifacts (evidence)

| Artifact | Location |
|---|---|
| Data handling + consent checklist | `docs/consent/volunteer-data-handling.md` |
| Pilot OAuth setup | `docs/development/pilot-mail-oauth.md` |
| Cohort weight vector (frozen) | `test/testdata/cohort_weights.json` |
| Scrubbed fixtures from volunteer mail | `test/testdata/email/merchants/` (via `make scrub-emls`) |

## Current state

- Pilot mailbox scrubbed fixtures committed (Amazon, Screwfix, Toolstation).
- Phase 1 gate runs on frozen corpus + labelled `.txt` fixtures — no volunteer label sets required.

## Next actions

- [ ] Grow scrubbed fixture set as volunteers connect
- [ ] Refresh cohort weights from Track A interviews when available
