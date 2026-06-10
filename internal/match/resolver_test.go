package match_test

import (
	"testing"
	"time"

	"github.com/jimbery/receipt/internal/match"
	"github.com/jimbery/receipt/internal/model"
)

func TestResolver_EqualConfidenceNeverSilentlyResolves(t *testing.T) {
	t.Parallel()
	engine := match.NewEngine(match.DefaultConfig())
	ts := model.NewTimestamp(scenarioBase(), 0)

	result := engine.Match(
		[]model.Transaction{
			{ID: "t1", Merchant: "BP CONNECT", Amount: model.NewMoney(5000, "GBP"), OccurredAt: ts},
			{ID: "t2", Merchant: "BP CONNECT", Amount: model.NewMoney(5000, "GBP"), OccurredAt: ts},
		},
		[]model.Receipt{
			{ID: "r1", Supplier: "BP", Total: model.NewMoney(5000, "GBP"), IssuedAt: ts},
			{ID: "r2", Supplier: "BP", Total: model.NewMoney(5000, "GBP"), IssuedAt: ts},
		},
	)

	if len(result.Matches) > 0 {
		t.Fatalf("equal-confidence ambiguous candidates must not silently resolve: matches=%v", result.Matches)
	}
	if len(result.Conflicts) == 0 {
		t.Fatal("expected conflicts for indistinguishable equal-confidence candidates")
	}
}

func TestResolver_Determinism200Iterations(t *testing.T) {
	engine := match.NewEngine(match.DefaultConfig())
	txns, receipts := ambiguousFixture()
	first := engine.Match(txns, receipts)
	for range 200 {
		next := engine.Match(txns, receipts)
		if !matchResultsEqual(first, next) {
			t.Fatal("conflict ordering or contents changed across repeated runs")
		}
	}
}

func ambiguousFixture() ([]model.Transaction, []model.Receipt) {
	ts := model.NewTimestamp(scenarioBase(), 0)
	return []model.Transaction{
			{ID: "t-a1", Merchant: "BP CONNECT", Amount: model.NewMoney(5000, "GBP"), OccurredAt: ts},
			{ID: "t-a2", Merchant: "BP CONNECT", Amount: model.NewMoney(5000, "GBP"), OccurredAt: ts},
		}, []model.Receipt{
			{ID: "r-a1", Supplier: "BP", Total: model.NewMoney(5000, "GBP"), IssuedAt: ts},
			{ID: "r-a2", Supplier: "BP", Total: model.NewMoney(5000, "GBP"), IssuedAt: ts},
		}
}

func scenarioBase() time.Time {
	return time.Date(2026, 3, 15, 14, 30, 0, 0, time.UTC)
}

func matchResultsEqual(a, b model.MatchResult) bool {
	if len(a.Matches) != len(b.Matches) || len(a.Conflicts) != len(b.Conflicts) {
		return false
	}
	for i, m := range a.Matches {
		if m.TransactionID != b.Matches[i].TransactionID ||
			m.ReceiptID != b.Matches[i].ReceiptID ||
			m.Confidence != b.Matches[i].Confidence {
			return false
		}
	}
	for i, c := range a.Conflicts {
		bc := b.Conflicts[i]
		if c.TransactionID != bc.TransactionID ||
			c.ReceiptID != bc.ReceiptID ||
			c.TopConfidence != bc.TopConfidence ||
			c.Reason != bc.Reason ||
			!equalStringSlices(c.CompetingIDs, bc.CompetingIDs) {
			return false
		}
	}
	return equalStringSlices(a.UnmatchedTxnIDs, b.UnmatchedTxnIDs) &&
		equalStringSlices(a.UnmatchedReceiptIDs, b.UnmatchedReceiptIDs)
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
