package match_test

import (
	"testing"

	"github.com/jimbery/receipt/internal/harness"
	"github.com/jimbery/receipt/internal/match"
	"github.com/jimbery/receipt/internal/synth"
)

func TestEngine_Determinism(t *testing.T) {
	engine := match.NewEngine(match.DefaultConfig())
	if !harness.CheckDeterminism(engine, synth.AllExtended()) {
		t.Fatal("matching engine output must be deterministic")
	}
}

func TestEngine_Determinism_Ambiguous200Iterations(t *testing.T) {
	engine := match.NewEngine(match.DefaultConfig())
	var ambiguous synth.Scenario
	for _, s := range synth.AllExtended() {
		if s.Class == synth.ClassAmbiguous {
			ambiguous = s
			break
		}
	}
	first := engine.Match(ambiguous.Transactions, ambiguous.Receipts)
	for range 200 {
		next := engine.Match(ambiguous.Transactions, ambiguous.Receipts)
		if !harness.ResultsEqual(first, next) {
			t.Fatal("ambiguous scenario must be deterministic across 200 iterations")
		}
	}
}
