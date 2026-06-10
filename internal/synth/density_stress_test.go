package synth_test

import (
	"fmt"
	"testing"

	"github.com/jimbery/receipt/internal/harness"
	"github.com/jimbery/receipt/internal/match"
	"github.com/jimbery/receipt/internal/model"
	"github.com/jimbery/receipt/internal/synth"
)

func TestDensityStress_ZeroFalseMatches(t *testing.T) {
	t.Parallel()
	engine := match.NewEngine(match.DefaultConfig())
	for _, n := range []int{10, 50, 200} {
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			t.Parallel()
			s := synth.DensityStress(n)
			result := engine.Match(s.Transactions, s.Receipts)
			if len(result.Matches) > 0 {
				t.Fatalf("density stress must not emit matches, got %d", len(result.Matches))
			}
			if len(result.Conflicts) == 0 {
				t.Fatal("expected conflicts")
			}
			m := harness.EvaluateResult(s.Transactions, s.Receipts, result, s.Labels)
			if m.FalseMatchRate > 0 {
				t.Fatalf("FMR %.3f must be zero", m.FalseMatchRate)
			}
			if v := model.ValidateOutcomes(result, s.Expectations); len(v) > 0 {
				t.Fatalf("outcome violations: %v", v)
			}
		})
	}
}
