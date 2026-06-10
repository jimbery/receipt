package synth

import (
	"time"

	"github.com/jimbery/receipt/internal/model"
)

const (
	ClassDuplicateReceipt    Class = "duplicate_receipt"
	ClassUnmatchedTxn        Class = "unmatched_txn"
	ClassUnmatchedReceipt    Class = "unmatched_receipt"
	ClassMerchantMangling    Class = "merchant_mangling"
	ClassPartialCapture      Class = "partial_capture"
	ClassFuelPreAuth         Class = "fuel_preauth"
	ClassSplitTenderRefusal  Class = "split_tender_refusal"
	ClassRefundRefusal       Class = "refund_refusal"
	ClassDensityStress       Class = "density_stress"
	ClassGenerated           Class = "generated"
	ClassGeneratedTip        Class = "generated_tip"
	ClassGeneratedSettlement Class = "generated_settlement"
	ClassGeneratedMangled    Class = "generated_mangled"
	ClassGeneratedFX         Class = "generated_fx"
	ClassGeneratedAmbiguous  Class = "generated_ambiguous"
	ClassGeneratedNearDup    Class = "generated_near_duplicate"
	ClassGeneratedRefund     Class = "generated_refund"
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
		fuelPreAuth(),
		splitTenderRefusal(),
		refundRefusal(),
		DensityStress(10),
	)
	return out
}

// DensityStress builds identical-merchant, identical-amount pairs at close temporal
// spacing — the regime that produced high FMR in b25c990. Permanent gate class:
// engine must emit conflict (zero false matches), not silent resolution.
func DensityStress(n int) Scenario {
	if n < 2 {
		n = 2
	}
	ts := model.NewTimestamp(scenarioBase().Add(500*time.Hour), 0)
	return buildIdenticalCluster(n, identicalClusterSpec{
		Name: "density_stress", Class: ClassDensityStress, IDPrefix: "ds",
		Merchant: "SCREWFIX 1234 LON", Supplier: "Screwfix", MCC: "5251", Amount: 5000, TS: ts,
	})
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
			GenuinelyAmbiguousTxnIDs:     []string{"t-dup"},
			GenuinelyAmbiguousReceiptIDs: []string{"r-dup-1", "r-dup-2"},
			TransactionOutcomes:          map[string]model.Outcome{"t-dup": model.OutcomeConflict},
			ReceiptOutcomes: map[string]model.Outcome{
				"r-dup-1": model.OutcomeConflict, "r-dup-2": model.OutcomeConflict,
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

func fuelPreAuth() Scenario {
	ts := model.NewTimestamp(scenarioBase(), 0)
	return Scenario{
		Name:  "fuel_preauth",
		Class: ClassFuelPreAuth,
		Transactions: []model.Transaction{{
			ID: "t-fuel", Merchant: "SHELL FUEL", MCC: "5541",
			Amount: model.NewMoney(10000, "GBP"), OccurredAt: ts,
		}},
		Receipts: []model.Receipt{{
			ID: "r-fuel", Supplier: "Shell", Total: model.NewMoney(8500, "GBP"),
			IssuedAt: model.NewTimestamp(scenarioBase().Add(10*time.Minute), 0),
		}},
		Labels: []model.LabelledPair{{TransactionID: "t-fuel", ReceiptID: "r-fuel"}},
		Expectations: model.Expectations{
			TransactionOutcomes: map[string]model.Outcome{"t-fuel": model.OutcomeMatched},
			ReceiptOutcomes:     map[string]model.Outcome{"r-fuel": model.OutcomeMatched},
		},
	}
}

func splitTenderRefusal() Scenario {
	ts := model.NewTimestamp(scenarioBase(), 0)
	return Scenario{
		Name:  "split_tender_refusal",
		Class: ClassSplitTenderRefusal,
		Transactions: []model.Transaction{{
			ID: "t-split", Merchant: "SCREWFIX", MCC: "5251",
			Amount: model.NewMoney(10000, "GBP"), OccurredAt: ts,
		}},
		Receipts: []model.Receipt{{
			ID: "r-split", Supplier: "Screwfix", Total: model.NewMoney(5000, "GBP"), IssuedAt: ts,
		}},
		Expectations: model.Expectations{
			TransactionOutcomes: map[string]model.Outcome{"t-split": model.OutcomeUnmatched},
			ReceiptOutcomes:     map[string]model.Outcome{"r-split": model.OutcomeUnmatched},
		},
	}
}

func refundRefusal() Scenario {
	ts := model.NewTimestamp(scenarioBase(), 0)
	return Scenario{
		Name:  "refund_refusal",
		Class: ClassRefundRefusal,
		Transactions: []model.Transaction{{
			ID: "t-refund", Merchant: "AMAZON UK", MCC: "5399",
			Amount: model.NewMoney(4500, "GBP"), OccurredAt: ts,
		}},
		Receipts: []model.Receipt{{
			ID: "r-refund", Supplier: "Amazon", Total: model.NewMoney(-4500, "GBP"), IssuedAt: ts,
		}},
		Expectations: model.Expectations{
			TransactionOutcomes: map[string]model.Outcome{"t-refund": model.OutcomeUnmatched},
			ReceiptOutcomes:     map[string]model.Outcome{"r-refund": model.OutcomeUnmatched},
		},
	}
}
