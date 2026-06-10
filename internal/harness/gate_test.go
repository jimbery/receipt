package harness_test

import (
	"testing"

	"github.com/jimbery/receipt/internal/harness"
	"github.com/jimbery/receipt/internal/match"
	"github.com/jimbery/receipt/internal/synth"
)

func TestCheckDeterminism(t *testing.T) {
	engine := match.NewEngine(match.DefaultConfig())
	if !harness.CheckDeterminism(engine, synth.AllExtended()) {
		t.Fatal("engine must be deterministic")
	}
}

func TestRunGate_HandBuiltScenarios(t *testing.T) {
	engine := match.NewEngine(match.DefaultConfig())
	report := harness.RunGate(engine, synth.AllExtended(), harness.DefaultGateThresholds())
	if !report.Deterministic {
		t.Fatal("expected deterministic")
	}
	if report.Overall.FalseMatchRate > harness.DefaultGateThresholds().MaxFMR {
		t.Fatalf("FMR %.4f exceeds threshold", report.Overall.FalseMatchRate)
	}
}

func TestConflictCorrectness_Ambiguous(t *testing.T) {
	engine := match.NewEngine(match.DefaultConfig())
	for _, s := range synth.AllExtended() {
		if s.Class != synth.ClassAmbiguous {
			continue
		}
		result := engine.Match(s.Transactions, s.Receipts)
		cc := harness.ConflictCorrectness(result, s.Expectations)
		if cc < 0.9 {
			t.Fatalf("conflict correctness %.2f want >=0.9", cc)
		}
	}
}
