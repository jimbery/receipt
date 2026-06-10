package synth_test

import (
	"testing"

	"github.com/jimbery/receipt/internal/match"
	"github.com/jimbery/receipt/internal/synth"
)

func TestAllScenarios_NonEmpty(t *testing.T) {
	scenarios := synth.All()
	if len(scenarios) < 5 {
		t.Fatalf("expected adversarial set, got %d scenarios", len(scenarios))
	}
	engine := match.NewEngine(match.DefaultConfig())
	for _, s := range scenarios {
		if len(s.Transactions) == 0 || len(s.Receipts) == 0 {
			t.Fatalf("scenario %s missing data", s.Name)
		}
		result := engine.Match(s.Transactions, s.Receipts)
		if s.Class == synth.ClassAmbiguous && len(result.Conflicts) == 0 && len(result.Matches) > 0 {
			t.Fatalf("ambiguous scenario should conflict or refuse, got matches %v", result.Matches)
		}
	}
}
