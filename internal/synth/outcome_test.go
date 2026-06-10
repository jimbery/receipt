package synth_test

import (
	"testing"

	"github.com/jimbery/receipt/internal/match"
	"github.com/jimbery/receipt/internal/model"
	"github.com/jimbery/receipt/internal/synth"
)

func TestAllExtended_OutcomeClasses(t *testing.T) {
	t.Parallel()
	engine := match.NewEngine(match.DefaultConfig())
	for _, s := range synth.AllExtended() {
		if len(s.Expectations.TransactionOutcomes) == 0 && len(s.Expectations.ReceiptOutcomes) == 0 {
			continue
		}
		t.Run(string(s.Class), func(t *testing.T) {
			t.Parallel()
			result := engine.Match(s.Transactions, s.Receipts)
			if v := model.ValidateOutcomes(result, s.Expectations); len(v) > 0 {
				t.Fatalf("outcome violations: %v", v)
			}
		})
	}
}

func TestGenerator_NoiseBounds(t *testing.T) {
	gen := synth.NewGenerator(99)
	ds := gen.Generate(50, synth.NoiseProfile{TipMinor: 10, SettlementLagHours: 1})
	if !ds.VerifyNoiseBounds() {
		t.Fatal("generated pairs exceed claimed noise bounds")
	}
}

func FuzzGenerator_NoiseBounds(f *testing.F) {
	f.Add(int64(1), int64(10))
	f.Fuzz(func(t *testing.T, seed, n int64) {
		if n < 1 {
			n = 1
		}
		if n > 200 {
			n = 200
		}
		gen := synth.NewGenerator(seed)
		ds := gen.Generate(int(n), synth.NoiseProfile{TipMinor: 50, SettlementLagHours: 3})
		if !ds.VerifyNoiseBounds() {
			t.Fatal("noise bounds violated")
		}
	})
}
