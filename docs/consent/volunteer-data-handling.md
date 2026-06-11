# Volunteer Mailbox Data Handling

**Version:** 1.0  
**Status:** Required before any OAuth token is issued (ADR-002 D5)

## Purpose

Track A volunteers connect personal mailboxes for pilot receipt ingestion. This document describes what we collect, how we protect it, and how volunteers can withdraw.

## What we access

- Email messages likely to contain purchase receipts (filtered by classifier before deep parsing)
- Message metadata: sender, subject, date, MIME structure
- Attachments referenced by receipt-bearing messages (PDF/HTML)

We do **not** access unrelated mailbox folders beyond the connected account scope agreed at consent.

## What we store

- Canonical `Receipt` records derived from purchase emails
- Scrubbed fixture excerpts only (PII replaced via `internal/scrub`) — never raw volunteer mailbox dumps in the repository
- Aggregated gate metrics (precision, recall, moat coverage)

## PII handling

- Names, addresses, phone numbers, and account identifiers are scrubbed before any fixture is committed
- OAuth tokens are stored locally on the volunteer machine during pilot; not committed to source control
- Ground-truth labelling sessions use message IDs and redacted previews

## Consent checklist

Volunteers must confirm:

1. They understand which mailbox is connected and for what pilot period
2. They may disconnect at any time and request deletion of derived records
3. Data is used only for Receipt Capture product development and gate measurement
4. Fixtures exported for engineering are PII-scrubbed

## Withdrawal

On withdrawal: revoke OAuth token, delete local snapshots within 30 days, remove volunteer ID from cohort weight tables. Credit notes and receipts already merged into anonymised evaluation corpora remain in scrubbed form only.
