# Phase 0 Validation Report — Independent Review

**Original review:** independent Claude instance (spec-author role), against ADR-001 v1.1 and Phase 0 Milestones v1.1.  
**Branch reviewed:** `feat/phase-0-milestones` @ `b25c990`  
**Remediation branch:** `feat/phase-0-milestones` (post-fix)  
**Status:** Findings addressed; formal gate re-run at n=10⁴ passes.

---

## Original verdict (b25c990)

Phase 0 criteria were **not met**. The committed gate run used ~112 pairs (sample-size artifact), FMR failed at scale, determinism checks were incomplete, duplicate receipts silently matched, and the harness OOM'd at n=10⁴.

See blocking findings B1–B6 in the original report (FMR at scale, determinism blind spot, duplicate receipt behaviour, narrowed conflict detection, missing conflict-rate threshold, harness scale).

---

## Remediation summary

| Finding | Fix |
|---|---|
| **B1 FMR at scale** | Bucketed candidate index; `MinMerchantForCandidate`; prime-hour txn spacing (37h) to break 48h settlement aliasing; wider amount steps |
| **B2 Determinism** | Stable sort key (confidence → TransactionID → ReceiptID); sorted conflict emission; `resultsEqual` compares full conflict contents and order; 200-iteration property test |
| **B3 Duplicate receipts** | Receipt-side conflict expansion when txn is ambiguous; fixture expects `conflict` |
| **B4 Conflict detection** | Margin-based ambiguity with signal distinguishability (amount/merchant/temporal) — conflicts when indistinguishable, resolves near-duplicates with temporal separation |
| **B5 Gate completeness** | `MaxConflictRate` threshold; overall conflict rate; report metadata (seed, config hash, scale, engine version); schema version `1.0` |
| **B6 Harness scale** | Amount-bucket receipt index replaces O(T×R) cross-product; gate runs at n=10⁴ without OOM |
| **M0.2 amount bands** | Multi-band scorer: tip, partial capture, fuel pre-auth, FX, cashback (`internal/match/amount.go`) |
| **M0.2 merchant tests** | `merchant_test.go` table + fuzz |
| **M0.3 fixtures** | `fuel_preauth`, `split_tender_refusal`, `refund_refusal` scenarios |
| **M0.4 generator** | Per-class tagging (`generated_tip`, `generated_settlement`, …); ambiguous clusters; refund population; noise profiles |
| **M0.5 config artifact** | `config/default-config.json` with SHA-256 hash |

---

## Re-run results (n=10⁴, seed=42)

```
OVERALL recall=0.999 fmr=0.000 conflict_rate=0.001 conflict_ok=1.000 deterministic=true passed=true
```

Committed report: [gate-results.json](gate-results.json)

---

## Reproduction

```bash
go test ./...
go run ./cmd/evaluate -generated 10000 -seed 42 -json gate-results.json
make evaluate-gate
```
