package match_test

import (
	"testing"
	"time"

	"github.com/jimbery/receipt/internal/harness"
	"github.com/jimbery/receipt/internal/match"
	"github.com/jimbery/receipt/internal/model"
)

func TestEngine_Match_Table(t *testing.T) {
	t.Parallel()
	base := time.Date(2026, 3, 15, 14, 30, 0, 0, time.UTC)
	ts := model.NewTimestamp(base, 0)

	tests := []struct {
		name         string
		txns         []model.Transaction
		receipts     []model.Receipt
		wantMatch    bool
		wantConflict bool
		wantTxnID    string
		wantRecID    string
	}{
		{
			name: "exact match",
			txns: []model.Transaction{{
				ID: "t1", Merchant: "TESCO STORES 3344 LONDN", MCC: "5411",
				Amount: model.NewMoney(4523, "GBP"), OccurredAt: ts,
			}},
			receipts: []model.Receipt{{
				ID: "r1", Supplier: "Tesco", Total: model.NewMoney(4523, "GBP"),
				IssuedAt: model.NewTimestamp(base.Add(2*time.Hour), 0),
			}},
			wantMatch: true, wantTxnID: "t1", wantRecID: "r1",
		},
		{
			name: "refuses wrong amount",
			txns: []model.Transaction{{
				ID: "t3", Merchant: "AMAZON",
				Amount: model.NewMoney(5000, "GBP"), OccurredAt: ts,
			}},
			receipts: []model.Receipt{{
				ID: "r3", Supplier: "Amazon", Total: model.NewMoney(12000, "GBP"),
				IssuedAt: model.NewTimestamp(base.Add(time.Hour), 0),
			}},
			wantMatch: false,
		},
		{
			name: "POS exact ref short-circuit",
			txns: []model.Transaction{{
				ID: "t-pos", Merchant: "SQ *CAFE", ExternalRef: "ref-1",
				Amount: model.NewMoney(500, "GBP"), OccurredAt: ts,
			}},
			receipts: []model.Receipt{{
				ID: "r-pos", Source: model.ReceiptSourcePOS, Supplier: "Cafe",
				TransactionRef: "ref-1", Total: model.NewMoney(500, "GBP"), IssuedAt: ts,
			}},
			wantMatch: true, wantTxnID: "t-pos", wantRecID: "r-pos",
		},
		{
			name: "ambiguous same timestamp emits conflict",
			txns: []model.Transaction{
				{ID: "t-a1", Merchant: "BP", Amount: model.NewMoney(5000, "GBP"), OccurredAt: ts},
				{ID: "t-a2", Merchant: "BP", Amount: model.NewMoney(5000, "GBP"), OccurredAt: ts},
			},
			receipts: []model.Receipt{
				{ID: "r-a1", Supplier: "BP", Total: model.NewMoney(5000, "GBP"), IssuedAt: ts},
				{ID: "r-a2", Supplier: "BP", Total: model.NewMoney(5000, "GBP"), IssuedAt: ts},
			},
			wantConflict: true,
		},
	}

	engine := match.NewEngine(match.DefaultConfig())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := engine.Match(tt.txns, tt.receipts)

			if tt.wantConflict {
				if len(result.Conflicts) == 0 {
					t.Fatal("expected conflict")
				}
				return
			}

			if !tt.wantMatch {
				if len(result.Matches) != 0 {
					t.Fatalf("expected no matches, got %v", result.Matches)
				}
				return
			}
			if len(result.Matches) != 1 {
				t.Fatalf("expected 1 match, got %d", len(result.Matches))
			}
			m := result.Matches[0]
			if m.Outcome != model.OutcomeMatched {
				t.Fatalf("outcome = %s", m.Outcome)
			}
			if m.TransactionID != tt.wantTxnID || m.ReceiptID != tt.wantRecID {
				t.Fatalf("got %s↔%s want %s↔%s", m.TransactionID, m.ReceiptID, tt.wantTxnID, tt.wantRecID)
			}
			if m.Confidence < engine.Config().MinConfidence && m.Method != model.MatchMethodExactRef {
				t.Fatalf("confidence %.3f below threshold", m.Confidence)
			}
		})
	}
}

func TestEngine_NearDuplicate_NoFalseMatch(t *testing.T) {
	base := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	engine := match.NewEngine(match.DefaultConfig())

	txns := []model.Transaction{
		{
			ID:         "t1",
			Merchant:   "SHELL FUEL",
			MCC:        "5541",
			Amount:     model.NewMoney(8000, "GBP"),
			OccurredAt: model.NewTimestamp(base, 0),
		},
		{
			ID:         "t2",
			Merchant:   "SHELL FUEL",
			MCC:        "5541",
			Amount:     model.NewMoney(8000, "GBP"),
			OccurredAt: model.NewTimestamp(base.Add(5*time.Minute), 0),
		},
	}
	receipts := []model.Receipt{
		{ID: "r1", Supplier: "Shell", Total: model.NewMoney(8000, "GBP"), IssuedAt: model.NewTimestamp(base, 0)},
		{
			ID:       "r2",
			Supplier: "Shell",
			Total:    model.NewMoney(8000, "GBP"),
			IssuedAt: model.NewTimestamp(base.Add(5*time.Minute), 0),
		},
	}

	result := engine.Match(txns, receipts)
	metrics := harness.EvaluateResult(txns, receipts, result, []harness.LabelledPair{
		{TransactionID: "t1", ReceiptID: "r1"},
		{TransactionID: "t2", ReceiptID: "r2"},
	})
	if metrics.FalseMatchRate > 0 {
		t.Fatalf("false match rate %.2f, result=%+v", metrics.FalseMatchRate, result)
	}
}

func FuzzEngine_NoFalsePositiveOnCurrencyMismatch(f *testing.F) {
	f.Add(int64(1000), int64(2000), int64(0))
	engine := match.NewEngine(match.DefaultConfig())
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	ts := model.NewTimestamp(base, 0)

	f.Fuzz(func(t *testing.T, txnAmount, recAmount, offsetHours int64) {
		if txnAmount < 0 {
			txnAmount = -txnAmount
		}
		if recAmount < 0 {
			recAmount = -recAmount
		}
		txnAmount %= 1_000_000
		recAmount %= 1_000_000
		if offsetHours < 0 {
			offsetHours = -offsetHours
		}
		offsetHours %= 168

		result := engine.Match(
			[]model.Transaction{{
				ID: "t", Merchant: "TEST MERCHANT",
				Amount: model.NewMoney(txnAmount, "GBP"), OccurredAt: ts,
			}},
			[]model.Receipt{{
				ID: "r", Supplier: "Test Merchant", Total: model.NewMoney(recAmount, "USD"),
				IssuedAt: model.NewTimestamp(base.Add(time.Duration(offsetHours)*time.Hour), 0),
			}},
		)
		if len(result.Matches) > 0 {
			t.Fatalf("currency mismatch must not match: %+v", result.Matches)
		}
	})
}
