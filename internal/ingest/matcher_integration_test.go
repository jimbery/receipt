package ingest_test

import (
	"testing"
	"time"

	"github.com/jimbery/receipt/internal/dedup"
	"github.com/jimbery/receipt/internal/emailtypes"
	"github.com/jimbery/receipt/internal/extract"
	"github.com/jimbery/receipt/internal/match"
	"github.com/jimbery/receipt/internal/model"
)

// M1.5 closed loop: dedup pass-through → canonicalise → Phase 0 matcher conflict (ADR-001).
func TestDedupAmbiguousDuplicates_ProduceMatcherConflict(t *testing.T) {
	day := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	fam := "vendor|2026-03-10|999|GBP"
	resolved := dedup.ResolveFamilies([]dedup.ExtractedWithProvenance{
		{MessageID: "amb-1", Kind: emailtypes.KindPurchaseReceipt, Extracted: emailtypes.ExtractedReceipt{
			SourceMessageID: "amb-1", Supplier: "Vendor", FamilyKey: fam,
			TotalMinor: 999, IssuedAt: day, Currency: "GBP", DocumentKind: emailtypes.KindPurchaseReceipt,
		}},
		{MessageID: "amb-2", Kind: emailtypes.KindPurchaseReceipt, Extracted: emailtypes.ExtractedReceipt{
			SourceMessageID: "amb-2", Supplier: "Vendor", FamilyKey: fam,
			TotalMinor: 999, IssuedAt: day, Currency: "GBP", DocumentKind: emailtypes.KindPurchaseReceipt,
		}},
	})
	if len(resolved) != 2 {
		t.Fatalf("dedup: got %d want 2 pass-through receipts", len(resolved))
	}

	receipts := make([]model.Receipt, 0, len(resolved))
	for _, ex := range resolved {
		r, err := extract.Canonicalise(ex, "mb-1")
		if err != nil {
			t.Fatal(err)
		}
		receipts = append(receipts, r)
	}

	txn := model.Transaction{
		ID:         "txn-1",
		Merchant:   "Vendor",
		Amount:     model.NewMoney(999, "GBP"),
		OccurredAt: model.TimestampFromTime(day),
	}
	result := match.NewEngine(match.DefaultConfig()).Match([]model.Transaction{txn}, receipts)
	if len(result.Conflicts) == 0 {
		t.Fatal("expected matcher conflict for ambiguous duplicate receipts from dedup")
	}
}
