package synth

import (
	"fmt"

	"github.com/jimbery/receipt/internal/model"
)

type identicalClusterSpec struct {
	Name     string
	Class    Class
	IDPrefix string
	Merchant string
	Supplier string
	MCC      string
	Amount   int64
	TS       model.Timestamp
}

func buildIdenticalCluster(n int, spec identicalClusterSpec) Scenario {
	s := Scenario{
		Name:  spec.Name,
		Class: spec.Class,
		Expectations: model.Expectations{
			GenuinelyAmbiguousTxnIDs:     make([]string, 0, n),
			GenuinelyAmbiguousReceiptIDs: make([]string, 0, n),
			TransactionOutcomes:          make(map[string]model.Outcome),
			ReceiptOutcomes:              make(map[string]model.Outcome),
		},
	}
	for i := range n {
		tid := fmt.Sprintf("%s-t-%d", spec.IDPrefix, i)
		rid := fmt.Sprintf("%s-r-%d", spec.IDPrefix, i)
		s.Transactions = append(s.Transactions, model.Transaction{
			ID: tid, Merchant: spec.Merchant, MCC: spec.MCC,
			Amount: model.NewMoney(spec.Amount, "GBP"), OccurredAt: spec.TS,
		})
		s.Receipts = append(s.Receipts, model.Receipt{
			ID: rid, Supplier: spec.Supplier, Total: model.NewMoney(spec.Amount, "GBP"), IssuedAt: spec.TS,
		})
		s.Labels = append(s.Labels, model.LabelledPair{TransactionID: tid, ReceiptID: rid})
		s.Expectations.GenuinelyAmbiguousTxnIDs = append(s.Expectations.GenuinelyAmbiguousTxnIDs, tid)
		s.Expectations.GenuinelyAmbiguousReceiptIDs = append(s.Expectations.GenuinelyAmbiguousReceiptIDs, rid)
		s.Expectations.TransactionOutcomes[tid] = model.OutcomeConflict
		s.Expectations.ReceiptOutcomes[rid] = model.OutcomeConflict
	}
	return s
}
