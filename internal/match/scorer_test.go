package match_test

import (
	"math"
	"testing"
	"time"

	"github.com/jimbery/receipt/internal/match"
	"github.com/jimbery/receipt/internal/model"
)

func TestAmountSimilarity_Bands(t *testing.T) {
	t.Parallel()
	cfg := match.DefaultConfig()
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name    string
		txn     int64
		receipt int64
		min     float64
	}{
		{"exact", 5000, 5000, 1.0},
		{"tip", 5000, 4990, 0.9},
		{"outside", 5000, 4800, 0},
	}

	engine := match.NewEngine(cfg)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := engine.Match(
				[]model.Transaction{{
					ID: "t", Merchant: "TEST", Amount: model.NewMoney(tc.txn, "GBP"),
					OccurredAt: model.NewTimestamp(base, 0),
				}},
				[]model.Receipt{{
					ID: "r", Supplier: "Test", Total: model.NewMoney(tc.receipt, "GBP"),
					IssuedAt: model.NewTimestamp(base, 0),
				}},
			)
			if tc.min == 0 {
				if len(result.Matches) > 0 {
					t.Fatal("expected no match")
				}
				return
			}
			if len(result.Matches) == 0 {
				t.Fatal("expected match")
			}
			if result.Matches[0].Signals["amount"] < tc.min {
				t.Fatalf("amount signal %.3f < %.3f", result.Matches[0].Signals["amount"], tc.min)
			}
		})
	}
}

func TestTemporalScorer_MonotonicDecay(t *testing.T) {
	t.Parallel()
	cfg := match.DefaultConfig()
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	engine := match.NewEngine(cfg)

	prev := 2.0
	for _, lag := range []time.Duration{0, time.Hour, 6 * time.Hour, 24 * time.Hour} {
		result := engine.Match(
			[]model.Transaction{{
				ID: "t", Merchant: "SCREWFIX", Amount: model.NewMoney(1000, "GBP"),
				OccurredAt: model.NewTimestamp(base, 0),
			}},
			[]model.Receipt{{
				ID: "r", Supplier: "Screwfix", Total: model.NewMoney(1000, "GBP"),
				IssuedAt: model.NewTimestamp(base.Add(lag), 0),
			}},
		)
		if len(result.Matches) == 0 {
			continue
		}
		score := result.Matches[0].Signals["temporal"]
		if score > prev && lag > 0 {
			t.Fatalf("temporal score should decay: lag=%s score=%.3f prev=%.3f", lag, score, prev)
		}
		prev = score
	}
}

func TestScorerSignals_InUnitInterval(t *testing.T) {
	t.Parallel()
	cfg := match.DefaultConfig()
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	engine := match.NewEngine(cfg)
	result := engine.Match(
		[]model.Transaction{{
			ID: "t", Merchant: "TESCO STORES", Amount: model.NewMoney(1234, "GBP"),
			OccurredAt: model.NewTimestamp(base, 0),
		}},
		[]model.Receipt{{
			ID: "r", Supplier: "Tesco", Total: model.NewMoney(1234, "GBP"),
			IssuedAt: model.NewTimestamp(base.Add(time.Hour), 0),
		}},
	)
	if len(result.Matches) == 0 {
		t.Fatal("expected match")
	}
	for k, v := range result.Matches[0].Signals {
		if k == "temporal_delta_secs" {
			continue
		}
		if v < 0 || v > 1+1e-9 {
			t.Fatalf("signal %s=%f outside [0,1]", k, v)
		}
		if math.IsNaN(v) {
			t.Fatalf("signal %s is NaN", k)
		}
	}
}
