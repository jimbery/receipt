package harness_test

import (
	"testing"

	"github.com/jimbery/receipt/internal/classify"
	iharness "github.com/jimbery/receipt/internal/ingest/harness"
)

func hardNegativePatterns() []string {
	return []string{
		"marketing_receipt_awaits",
		"marketing_newsletter",
		"marketing_sale_ends",
		"dispatch_amazon_tracking",
		"dispatch_no_total",
		"statement_bank_html",
		"statement_pdf",
		"credit_note_standalone",
	}
}

func TestFrozenCorpus_HardNegativePatterns(t *testing.T) {
	clf := classify.New()
	corpus := iharness.LoadFrozenCorpus()

	seen := make(map[string]struct{}, len(hardNegativePatterns()))
	for _, c := range corpus {
		if c.Stratum != "hard_negative" {
			continue
		}
		if c.Pattern == "" {
			t.Fatalf("%s: missing pattern id", c.ID)
		}
		seen[c.Pattern] = struct{}{}
		got := clf.Classify(c.Msg)
		if got != c.Label {
			t.Errorf("%s (%s): got %s want %s", c.ID, c.Pattern, got, c.Label)
		}
	}
	for _, p := range hardNegativePatterns() {
		if _, ok := seen[p]; !ok {
			t.Errorf("corpus missing hard-negative pattern %q", p)
		}
	}
}

func TestFrozenCorpus_PriorityMerchantLabels(t *testing.T) {
	clf := classify.New()
	corpus := iharness.LoadFrozenCorpus()

	seen := make(map[string]struct{})
	for _, c := range corpus {
		if c.Stratum != "priority_merchant" {
			continue
		}
		if c.Pattern == "" {
			t.Fatalf("%s: missing pattern id", c.ID)
		}
		seen[c.Pattern] = struct{}{}
		got := clf.Classify(c.Msg)
		if got != c.Label {
			t.Errorf("%s (%s): got %s want %s", c.ID, c.Pattern, got, c.Label)
		}
	}
	if len(seen) != 15 {
		t.Fatalf("expected 15 priority-merchant patterns, got %d", len(seen))
	}
}

func TestFrozenCorpus_LongTailLabels(t *testing.T) {
	clf := classify.New()
	corpus := iharness.LoadFrozenCorpus()

	seen := make(map[string]struct{})
	for _, c := range corpus {
		if c.Stratum != "long_tail" {
			continue
		}
		seen[c.Pattern] = struct{}{}
		got := clf.Classify(c.Msg)
		if got != c.Label {
			t.Errorf("%s (%s): got %s want %s", c.ID, c.Pattern, got, c.Label)
		}
	}
	if len(seen) != 5 {
		t.Fatalf("expected 5 long-tail patterns, got %d", len(seen))
	}
}
