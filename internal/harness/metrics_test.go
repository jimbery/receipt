package harness_test

import (
	"testing"

	"github.com/jimbery/receipt/internal/harness"
	"github.com/jimbery/receipt/internal/model"
)

func TestEvaluate_PrecisionRecall(t *testing.T) {
	txns := []model.Transaction{{ID: "t1"}, {ID: "t2"}}
	receipts := []model.Receipt{
		{ID: "r1", LineItems: []model.LineItem{{Description: "x"}}},
		{ID: "r2"},
	}
	matches := []model.Match{{TransactionID: "t1", ReceiptID: "r1", Confidence: 0.9}}
	labels := []harness.LabelledPair{{TransactionID: "t1", ReceiptID: "r1"}}

	m := harness.Evaluate(txns, receipts, matches, labels)

	if m.Precision != 1.0 || m.Recall != 1.0 {
		t.Fatalf("precision=%f recall=%f", m.Precision, m.Recall)
	}
	if m.ItemisationRate != 0.5 {
		t.Fatalf("itemisation rate = %f", m.ItemisationRate)
	}
}

func TestCohortWeightedItemisationRate(t *testing.T) {
	weights := []harness.CohortWeight{
		{Merchant: "Screwfix", Weight: 0.5},
		{Merchant: "Amazon", Weight: 0.5},
	}
	receipts := []model.Receipt{
		{Supplier: "Screwfix", LineItems: []model.LineItem{{Description: "drill"}}},
	}

	rate := harness.CohortWeightedItemisationRate(receipts, weights)
	if rate != 0.5 {
		t.Fatalf("cohort weighted rate = %f, want 0.5", rate)
	}
}
