package synth

import (
	"time"

	"github.com/jimbery/receipt/internal/model"
)

// Class identifies an adversarial scenario category for per-class evaluation (ADR D7).
type Class string

const (
	ClassExact           Class = "exact"
	ClassTipAdjusted     Class = "tip_adjusted"
	ClassTimezoneShift   Class = "timezone_shift"
	ClassNearDuplicate   Class = "near_duplicate"
	ClassSettlementDelay Class = "settlement_delay"
	ClassAmbiguous       Class = "ambiguous"
	ClassExactRef        Class = "exact_ref"
)

// Scenario is a labelled synthetic dataset.
type Scenario struct {
	Name         string
	Class        Class
	Transactions []model.Transaction
	Receipts     []model.Receipt
	Labels       []model.LabelledPair
	Expectations model.Expectations
}

func scenarioBase() time.Time {
	return time.Date(2026, 3, 15, 14, 30, 0, 0, time.UTC)
}

// All returns the Phase 0 adversarial scenario set.
func All() []Scenario {
	return []Scenario{
		exact(),
		tipAdjusted(),
		timezoneShift(),
		nearDuplicate(),
		settlementDelay(),
		ambiguous(),
		exactRef(),
	}
}

func exact() Scenario {
	ts := model.NewTimestamp(scenarioBase(), 0)
	return Scenario{
		Name:  "exact",
		Class: ClassExact,
		Transactions: []model.Transaction{{
			ID: "t-exact", Merchant: "TESCO STORES 3344 LONDN", MCC: "5411",
			Amount: model.NewMoney(4523, "GBP"), OccurredAt: ts,
		}},
		Receipts: []model.Receipt{{
			ID: "r-exact", Supplier: "Tesco", Total: model.NewMoney(4523, "GBP"),
			IssuedAt:  model.NewTimestamp(scenarioBase().Add(2*time.Hour), 0),
			LineItems: []model.LineItem{{Description: "groceries", NetAmount: model.NewMoney(4523, "GBP")}},
		}},
		Labels: []model.LabelledPair{{TransactionID: "t-exact", ReceiptID: "r-exact"}},
		Expectations: model.Expectations{
			TransactionOutcomes: map[string]model.Outcome{"t-exact": model.OutcomeMatched},
			ReceiptOutcomes:     map[string]model.Outcome{"r-exact": model.OutcomeMatched},
		},
	}
}

func tipAdjusted() Scenario {
	ts := model.NewTimestamp(scenarioBase(), 0)
	return Scenario{
		Name:  "tip_adjusted",
		Class: ClassTipAdjusted,
		Transactions: []model.Transaction{{
			ID: "t-tip", Merchant: "SCREWFIX", MCC: "5251",
			Amount: model.NewMoney(10000, "GBP"), OccurredAt: ts,
		}},
		Receipts: []model.Receipt{{
			ID: "r-tip", Supplier: "Screwfix", Total: model.NewMoney(9990, "GBP"),
			IssuedAt: model.NewTimestamp(scenarioBase().Add(-15*time.Minute), 0),
		}},
		Labels: []model.LabelledPair{{TransactionID: "t-tip", ReceiptID: "r-tip"}},
	}
}

func timezoneShift() Scenario {
	// Transaction recorded in US Eastern (-5h), receipt in UK — UTC normalised.
	eastern := time.FixedZone("EST", -5*3600)
	localTxn := time.Date(2026, 3, 15, 9, 30, 0, 0, eastern)
	return Scenario{
		Name:  "timezone_shift",
		Class: ClassTimezoneShift,
		Transactions: []model.Transaction{{
			ID: "t-tz", Merchant: "AMAZON UK", MCC: "5399",
			Amount:     model.NewMoney(2499, "GBP"),
			OccurredAt: model.TimestampFromTime(localTxn),
		}},
		Receipts: []model.Receipt{{
			ID: "r-tz", Supplier: "Amazon", Total: model.NewMoney(2499, "GBP"),
			IssuedAt: model.NewTimestamp(localTxn.UTC().Add(30*time.Minute), 0),
		}},
		Labels: []model.LabelledPair{{TransactionID: "t-tz", ReceiptID: "r-tz"}},
	}
}

func nearDuplicate() Scenario {
	t1 := scenarioBase()
	t2 := scenarioBase().Add(5 * time.Minute)
	return Scenario{
		Name:  "near_duplicate",
		Class: ClassNearDuplicate,
		Transactions: []model.Transaction{
			{
				ID:         "t-nd1",
				Merchant:   "SHELL FUEL",
				MCC:        "5541",
				Amount:     model.NewMoney(8000, "GBP"),
				OccurredAt: model.NewTimestamp(t1, 0),
			},
			{
				ID:         "t-nd2",
				Merchant:   "SHELL FUEL",
				MCC:        "5541",
				Amount:     model.NewMoney(8000, "GBP"),
				OccurredAt: model.NewTimestamp(t2, 0),
			},
		},
		Receipts: []model.Receipt{
			{ID: "r-nd1", Supplier: "Shell", Total: model.NewMoney(8000, "GBP"), IssuedAt: model.NewTimestamp(t1, 0)},
			{ID: "r-nd2", Supplier: "Shell", Total: model.NewMoney(8000, "GBP"), IssuedAt: model.NewTimestamp(t2, 0)},
		},
		Labels: []model.LabelledPair{
			{TransactionID: "t-nd1", ReceiptID: "r-nd1"},
			{TransactionID: "t-nd2", ReceiptID: "r-nd2"},
		},
	}
}

func settlementDelay() Scenario {
	auth := scenarioBase()
	settled := scenarioBase().Add(72 * time.Hour)
	return Scenario{
		Name:  "settlement_delay",
		Class: ClassSettlementDelay,
		Transactions: []model.Transaction{{
			ID: "t-settle", Merchant: "TOOLSTATION", MCC: "5251",
			Amount: model.NewMoney(6750, "GBP"), OccurredAt: model.NewTimestamp(auth, 0),
			SettledAt: ptrTimestamp(model.NewTimestamp(settled, 0)),
		}},
		Receipts: []model.Receipt{{
			ID: "r-settle", Supplier: "Toolstation", Total: model.NewMoney(6750, "GBP"),
			IssuedAt: model.NewTimestamp(auth.Add(10*time.Minute), 0),
		}},
		Labels: []model.LabelledPair{{TransactionID: "t-settle", ReceiptID: "r-settle"}},
	}
}

func ambiguous() Scenario {
	// Same amount, same merchant, same time — must conflict, not false-match.
	ts := model.NewTimestamp(scenarioBase(), 0)
	return Scenario{
		Name:  "ambiguous",
		Class: ClassAmbiguous,
		Transactions: []model.Transaction{
			{ID: "t-a1", Merchant: "BP CONNECT", MCC: "5541", Amount: model.NewMoney(5000, "GBP"), OccurredAt: ts},
			{ID: "t-a2", Merchant: "BP CONNECT", MCC: "5541", Amount: model.NewMoney(5000, "GBP"), OccurredAt: ts},
		},
		Receipts: []model.Receipt{
			{ID: "r-a1", Supplier: "BP", Total: model.NewMoney(5000, "GBP"), IssuedAt: ts},
			{ID: "r-a2", Supplier: "BP", Total: model.NewMoney(5000, "GBP"), IssuedAt: ts},
		},
		Labels: []model.LabelledPair{
			{TransactionID: "t-a1", ReceiptID: "r-a1"},
			{TransactionID: "t-a2", ReceiptID: "r-a2"},
		},
		Expectations: model.Expectations{
			GenuinelyAmbiguousTxnIDs:     []string{"t-a1", "t-a2"},
			GenuinelyAmbiguousReceiptIDs: []string{"r-a1", "r-a2"},
			TransactionOutcomes: map[string]model.Outcome{
				"t-a1": model.OutcomeConflict, "t-a2": model.OutcomeConflict,
			},
			ReceiptOutcomes: map[string]model.Outcome{
				"r-a1": model.OutcomeConflict, "r-a2": model.OutcomeConflict,
			},
		},
	}
}

func exactRef() Scenario {
	ts := model.NewTimestamp(scenarioBase(), 0)
	return Scenario{
		Name:  "exact_ref",
		Class: ClassExactRef,
		Transactions: []model.Transaction{{
			ID: "t-pos", Merchant: "SQ *CAFE", ExternalRef: "pos-ref-99",
			Amount: model.NewMoney(350, "GBP"), OccurredAt: ts,
		}},
		Receipts: []model.Receipt{{
			ID: "r-pos", Source: model.ReceiptSourcePOS, Supplier: "Cafe",
			TransactionRef: "pos-ref-99", Total: model.NewMoney(350, "GBP"),
			IssuedAt: ts,
		}},
		Labels: []model.LabelledPair{{TransactionID: "t-pos", ReceiptID: "r-pos"}},
	}
}

func ptrTimestamp(t model.Timestamp) *model.Timestamp {
	return &t
}
