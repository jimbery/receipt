package dedup_test

import (
	"testing"
	"time"

	"github.com/jimbery/receipt/internal/dedup"
	"github.com/jimbery/receipt/internal/emailtypes"
)

// M1.5 milestone fixture table (Phase 0 style — declared outcomes).
func TestResolveFamilies_MilestoneFixtureTable(t *testing.T) {
	day := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	mk := func(id, ref, fam string, kind emailtypes.DocumentKind, total int64) dedup.ExtractedWithProvenance {
		return dedup.ExtractedWithProvenance{
			MessageID: id,
			Kind:      kind,
			Extracted: emailtypes.ExtractedReceipt{
				SourceMessageID: id,
				Supplier:        "Amazon",
				OrderRef:        ref,
				FamilyKey:       fam,
				TotalMinor:      total,
				IssuedAt:        day,
				Currency:        "GBP",
				DocumentKind:    kind,
			},
		}
	}

	cases := []struct {
		name       string
		in         []dedup.ExtractedWithProvenance
		wantCount  int
		wantWinner string
		wantMerged []string
		wantKinds  map[string]emailtypes.DocumentKind
	}{
		{
			name: "confirmation + invoice, same order ref",
			in: []dedup.ExtractedWithProvenance{
				mk("confirm-1", "ORD1", "amazon|ORD1", emailtypes.KindDispatchNotice, 1000),
				mk("invoice-1", "ORD1", "amazon|ORD1", emailtypes.KindPurchaseReceipt, 1000),
			},
			wantCount:  1,
			wantWinner: "invoice-1",
			wantMerged: []string{"confirm-1"},
		},
		{
			name: "same receipt forwarded twice",
			in: []dedup.ExtractedWithProvenance{
				mk("fwd-a", "ORD2", "amazon|ORD2", emailtypes.KindPurchaseReceipt, 2000),
				mk("fwd-b", "ORD2", "amazon|ORD2", emailtypes.KindPurchaseReceipt, 2000),
			},
			wantCount:  1,
			wantWinner: "fwd-a",
			wantMerged: []string{"fwd-b"},
		},
		{
			name: "amazon order, 2 shipments",
			in: []dedup.ExtractedWithProvenance{
				mk("ship-a", "ORD3-SHIP-A", "amazon|ORD3-SHIP-A", emailtypes.KindPurchaseReceipt, 1500),
				mk("ship-b", "ORD3-SHIP-B", "amazon|ORD3-SHIP-B", emailtypes.KindPurchaseReceipt, 2500),
			},
			wantCount: 2,
		},
		{
			name: "two genuinely identical purchases, same day",
			in: []dedup.ExtractedWithProvenance{
				{MessageID: "coffee-1", Kind: emailtypes.KindPurchaseReceipt, Extracted: emailtypes.ExtractedReceipt{
					SourceMessageID: "coffee-1", Supplier: "Cafe", FamilyKey: "cafe|2026-03-10|350|GBP",
					TotalMinor: 350, IssuedAt: day, Currency: "GBP",
				}},
				{MessageID: "coffee-2", Kind: emailtypes.KindPurchaseReceipt, Extracted: emailtypes.ExtractedReceipt{
					SourceMessageID: "coffee-2", Supplier: "Cafe", FamilyKey: "cafe|2026-03-10|350|GBP",
					TotalMinor: 350, IssuedAt: day, Currency: "GBP",
				}},
			},
			wantCount: 2,
		},
		{
			name: "credit note alongside original invoice",
			in: []dedup.ExtractedWithProvenance{
				mk("inv-1", "INV9", "shop|INV9", emailtypes.KindPurchaseReceipt, 5000),
				{MessageID: "cn-1", Kind: emailtypes.KindCreditNote, Extracted: emailtypes.ExtractedReceipt{
					SourceMessageID: "cn-1", Supplier: "Shop", OrderRef: "INV9", FamilyKey: "shop|INV9",
					TotalMinor: 1000, IssuedAt: day, Currency: "GBP", DocumentKind: emailtypes.KindCreditNote,
				}},
			},
			wantCount: 2,
			wantKinds: map[string]emailtypes.DocumentKind{
				"inv-1": emailtypes.KindPurchaseReceipt,
				"cn-1":  emailtypes.KindCreditNote,
			},
		},
		{
			name: "ambiguous family (no refs, similar totals)",
			in: []dedup.ExtractedWithProvenance{
				{MessageID: "amb-1", Kind: emailtypes.KindPurchaseReceipt, Extracted: emailtypes.ExtractedReceipt{
					SourceMessageID: "amb-1", Supplier: "Vendor", FamilyKey: "vendor|2026-03-10|999|GBP",
					TotalMinor: 999, IssuedAt: day, Currency: "GBP",
				}},
				{MessageID: "amb-2", Kind: emailtypes.KindPurchaseReceipt, Extracted: emailtypes.ExtractedReceipt{
					SourceMessageID: "amb-2", Supplier: "Vendor", FamilyKey: "vendor|2026-03-10|999|GBP",
					TotalMinor: 999, IssuedAt: day, Currency: "GBP",
				}},
			},
			wantCount: 2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := dedup.ResolveFamilies(tc.in)
			if len(out) != tc.wantCount {
				t.Fatalf("got %d receipts, want %d", len(out), tc.wantCount)
			}
			if tc.wantWinner != "" {
				if out[0].SourceMessageID != tc.wantWinner {
					t.Fatalf("winner %s want %s", out[0].SourceMessageID, tc.wantWinner)
				}
				if len(out[0].MergedFrom) != len(tc.wantMerged) {
					t.Fatalf("merged_from %v want %v", out[0].MergedFrom, tc.wantMerged)
				}
				for i, id := range tc.wantMerged {
					if out[0].MergedFrom[i] != id {
						t.Fatalf("merged_from[%d]=%s want %s", i, out[0].MergedFrom[i], id)
					}
				}
			}
			if tc.wantKinds != nil {
				byID := make(map[string]emailtypes.DocumentKind, len(out))
				for _, r := range out {
					byID[r.SourceMessageID] = r.DocumentKind
				}
				for id, kind := range tc.wantKinds {
					if byID[id] != kind {
						t.Fatalf("%s: kind %s want %s", id, byID[id], kind)
					}
				}
			}
		})
	}
}
