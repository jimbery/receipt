package test_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jimbery/receipt/internal/harness"
	"github.com/jimbery/receipt/internal/match"
	"github.com/jimbery/receipt/test/fixture"
)

func TestMatchingScenarios_FromFixtures(t *testing.T) {
	t.Parallel()

	engine := match.NewEngine(match.DefaultConfig())

	files := []string{
		"exact_pair.json",
		"tip_tolerance.json",
		"ambiguous_duplicates.json",
		"ambiguous_same_time.json",
	}

	for _, file := range files {
		t.Run(file, func(t *testing.T) {
			t.Parallel()
			path := fixturePath("matching", file)
			scenario, err := fixture.LoadScenario(path)
			if err != nil {
				t.Fatalf("load fixture: %v", err)
			}

			txns := scenario.Transactions()
			receipts := scenario.Receipts()
			result := engine.Match(txns, receipts)

			if scenario.ExpectMatch && len(result.Matches) == 0 && !scenario.ExpectConflict {
				t.Fatal("expected at least one match")
			}
			if scenario.ExpectConflict && len(result.Conflicts) == 0 {
				t.Fatal("expected conflict")
			}

			metrics := harness.EvaluateResult(txns, receipts, result, scenario.Labels)

			if scenario.ExpectPrecision != nil && metrics.Precision < *scenario.ExpectPrecision {
				t.Fatalf("precision %.3f below minimum %.3f", metrics.Precision, *scenario.ExpectPrecision)
			}
			if scenario.ExpectFMR != nil && metrics.FalseMatchRate > *scenario.ExpectFMR {
				t.Fatalf("FMR %.3f above maximum %.3f", metrics.FalseMatchRate, *scenario.ExpectFMR)
			}

			for _, m := range result.Matches {
				if m.Confidence < engine.Config().MinConfidence && m.Method != "exact_ref" {
					t.Fatalf("match below MinConfidence: %+v", m)
				}
			}
		})
	}
}

func fixturePath(parts ...string) string {
	_, file, _, _ := runtime.Caller(0)
	base := filepath.Join(filepath.Dir(file), "testdata")
	return filepath.Join(append([]string{base}, parts...)...)
}
