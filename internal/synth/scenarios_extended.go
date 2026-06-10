package synth

import (
	"time"

	"github.com/jimbery/receipt/internal/model"
)

const (
	ClassDuplicateReceipt Class = "duplicate_receipt"
	ClassUnmatchedTxn     Class = "unmatched_txn"
	ClassUnmatchedReceipt Class = "unmatched_receipt"
	ClassMerchantMangling Class = "merchant_mangling"
	ClassPartialCapture   Class = "partial_capture"
	ClassGenerated        Class = "generated"
)

// AllExtended returns hand-built scenarios plus extended outcome-class fixtures (M0.3).
func AllExtended() []Scenario {
	out := All()
	out = append(out,
		duplicateReceipt(),
		unmatchedTxn(),
		unmatchedReceipt(),
		merchantMangling(),
		partialCapture(),
	)
	return out
}

func duplicateReceipt() Scenario {
	ts := model.NewTimestamp(scenarioBase(), 0)
	return Scenario{
		Name:  "duplicate_receipt",
		Class: ClassDuplicateReceipt,
		Transactions: []model.Transaction{{
			ID: "t-dup", Merchant: "AMAZON UK", MCC: "5399",
			Amount: model.NewMoney(3200, "GBP"), OccurredAt: ts,
		}},
		Receipts: []model.Receipt{
			{ID: "r-dup-1", Supplier: "Amazon", Total: model.NewMoney(3200, "GBP"), IssuedAt: ts},
			{ID: "r-dup-2", Supplier: "Amazon", Total: model.NewMoney(3200, "GBP"), IssuedAt: ts},
		},
		Labels: []model.LabelledPair{{TransactionID: "t-dup", ReceiptID: "r-dup-1"}},
		Expectations: model.Expectations{
			TransactionOutcomes: map[string]model.Outcome{"t-dup": model.OutcomeMatched},
			ReceiptOutcomes: map[string]model.Outcome{
				"r-dup-2": model.OutcomeUnmatched,
			},
		},
	}
}

func unmatchedTxn() Scenario {
	ts := model.NewTimestamp(scenarioBase(), 0)
	return Scenario{
		Name:  "unmatched_txn",
		Class: ClassUnmatchedTxn,
		Transactions: []model.Transaction{{
			ID: "t-alone", Merchant: "UNKNOWN MERCHANT",
			Amount: model.NewMoney(1500, "GBP"), OccurredAt: ts,
		}},
		Expectations: model.Expectations{
			TransactionOutcomes: map[string]model.Outcome{"t-alone": model.OutcomeUnmatched},
		},
	}
}

func unmatchedReceipt() Scenario {
	ts := model.NewTimestamp(scenarioBase(), 0)
	return Scenario{
		Name:  "unmatched_receipt",
		Class: ClassUnmatchedReceipt,
		Receipts: []model.Receipt{{
			ID: "r-alone", Supplier: "Lonely Supplier",
			Total: model.NewMoney(2200, "GBP"), IssuedAt: ts,
		}},
		Expectations: model.Expectations{
			ReceiptOutcomes: map[string]model.Outcome{"r-alone": model.OutcomeUnmatched},
		},
	}
}

func merchantMangling() Scenario {
	ts := model.NewTimestamp(scenarioBase(), 0)
	return Scenario{
		Name:  "merchant_mangling",
		Class: ClassMerchantMangling,
		Transactions: []model.Transaction{{
			ID: "t-mang", Merchant: "SQ *SCREWFIX 4521 LONDON", MCC: "5251",
			Amount: model.NewMoney(4500, "GBP"), OccurredAt: ts,
		}},
		Receipts: []model.Receipt{{
			ID: "r-mang", Supplier: "Screwfix", Total: model.NewMoney(4500, "GBP"), IssuedAt: ts,
		}},
		Labels: []model.LabelledPair{{TransactionID: "t-mang", ReceiptID: "r-mang"}},
		Expectations: model.Expectations{
			TransactionOutcomes: map[string]model.Outcome{"t-mang": model.OutcomeMatched},
			ReceiptOutcomes:     map[string]model.Outcome{"r-mang": model.OutcomeMatched},
		},
	}
}

func partialCapture() Scenario {
	ts := model.NewTimestamp(scenarioBase(), 0)
	return Scenario{
		Name:  "partial_capture",
		Class: ClassPartialCapture,
		Transactions: []model.Transaction{{
			ID: "t-part", Merchant: "TOOLSTATION", MCC: "5251",
			Amount: model.NewMoney(10000, "GBP"), OccurredAt: ts,
		}},
		Receipts: []model.Receipt{{
			ID: "r-part", Supplier: "Toolstation", Total: model.NewMoney(9950, "GBP"),
			IssuedAt: model.NewTimestamp(scenarioBase().Add(5*time.Minute), 0),
		}},
		Labels: []model.LabelledPair{{TransactionID: "t-part", ReceiptID: "r-part"}},
		Expectations: model.Expectations{
			TransactionOutcomes: map[string]model.Outcome{"t-part": model.OutcomeMatched},
			ReceiptOutcomes:     map[string]model.Outcome{"r-part": model.OutcomeMatched},
		},
	}
}
