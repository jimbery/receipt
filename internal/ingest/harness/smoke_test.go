package harness_test

import (
	"testing"

	iharness "github.com/jimbery/receipt/internal/ingest/harness"
)

func TestLoadFrozenCorpus_SizeAndStrata(t *testing.T) {
	corpus := iharness.LoadFrozenCorpus()
	if len(corpus) < 500 {
		t.Fatalf("corpus size %d < 500", len(corpus))
	}
	seen := make(map[string]struct{})
	for _, c := range corpus {
		key := c.Msg.BodyHTML + c.Msg.BodyText + c.Msg.Subject
		seen[key] = struct{}{}
	}
	if len(seen) < 400 {
		t.Fatalf("only %d unique bodies; expected structural diversity", len(seen))
	}
}

func TestFixtureSmoke_Passes(t *testing.T) {
	corpus := iharness.LoadFrozenCorpus()
	report := iharness.RunFixtureSmoke(corpus, iharness.DefaultSmokeConfig())
	if !report.Passed {
		t.Fatalf("smoke failed: %v", report.Failures)
	}
}
