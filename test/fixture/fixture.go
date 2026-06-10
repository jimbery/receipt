// Package fixture loads versioned JSON scenarios from test/testdata for gate tests.
package fixture

import (
	"encoding/json"
	"os"
	"time"

	"github.com/jimbery/receipt/internal/harness"
	"github.com/jimbery/receipt/internal/model"
)

type Scenario struct {
	Name            string                 `json:"name"`
	TxnFixtures     []TransactionFixture   `json:"transactions"`
	ReceiptFixtures []ReceiptFixture       `json:"receipts"`
	Labels          []harness.LabelledPair `json:"labels"`
	ExpectMatch     bool                   `json:"expect_match"`
	ExpectConflict  bool                   `json:"expect_conflict"`
	ExpectPrecision *float64               `json:"expect_precision_min"`
	ExpectFMR       *float64               `json:"expect_fmr_max"`
}

type TransactionFixture struct {
	ID         string `json:"id"`
	Merchant   string `json:"merchant"`
	MCC        string `json:"mcc,omitempty"`
	Amount     int64  `json:"amount_minor"`
	Currency   string `json:"currency"`
	OccurredAt string `json:"occurred_at"`
}

type ReceiptFixture struct {
	ID             string `json:"id"`
	Supplier       string `json:"supplier"`
	Amount         int64  `json:"amount_minor"`
	Currency       string `json:"currency"`
	IssuedAt       string `json:"issued_at"`
	Itemised       bool   `json:"itemised"`
	Category       string `json:"category,omitempty"`
	Source         string `json:"source,omitempty"`
	TransactionRef string `json:"transaction_ref,omitempty"`
}

func LoadScenario(path string) (Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Scenario{}, err
	}
	var s Scenario
	if unmarshalErr := json.Unmarshal(data, &s); unmarshalErr != nil {
		return Scenario{}, unmarshalErr
	}
	return s, nil
}

func (s Scenario) Transactions() []model.Transaction {
	out := make([]model.Transaction, len(s.TxnFixtures))
	for i, t := range s.TxnFixtures {
		out[i] = model.Transaction{
			ID:         t.ID,
			Source:     model.TransactionSourceSynthetic,
			Merchant:   t.Merchant,
			MCC:        t.MCC,
			Amount:     model.NewMoney(t.Amount, t.Currency),
			OccurredAt: parseTimestamp(t.OccurredAt),
		}
	}
	return out
}

func (s Scenario) Receipts() []model.Receipt {
	out := make([]model.Receipt, len(s.ReceiptFixtures))
	for i, r := range s.ReceiptFixtures {
		source := model.ReceiptSourceEmail
		if r.Source == "pos" {
			source = model.ReceiptSourcePOS
		}
		rec := model.Receipt{
			ID:             r.ID,
			Source:         source,
			Supplier:       r.Supplier,
			Total:          model.NewMoney(r.Amount, r.Currency),
			IssuedAt:       parseTimestamp(r.IssuedAt),
			Category:       r.Category,
			TransactionRef: r.TransactionRef,
		}
		if r.Itemised {
			rec.LineItems = []model.LineItem{{
				Description: "item",
				NetAmount:   rec.Total,
			}}
		}
		out[i] = rec
	}
	return out
}

func parseTimestamp(s string) model.Timestamp {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic("fixture: invalid time " + s)
	}
	return model.TimestampFromTime(t)
}
