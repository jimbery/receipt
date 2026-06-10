package harness

import (
	"github.com/jimbery/receipt/internal/model"
)

// LabelledPair is ground truth for evaluation (defined in model).
type LabelledPair = model.LabelledPair

type Metrics struct {
	ItemisationRate           float64
	CohortWeightedItemisation float64
	MatchRate                 float64
	FalseMatchRate            float64
	ConflictRate              float64
	Precision                 float64
	Recall                    float64
	CategorisationAccuracy    float64
	Matched                   int
	Conflicts                 int
	UnmatchedReceipts         int
	UnmatchedTransactions     int
	FalseMatches              int
}

type CohortWeight struct {
	Merchant string  `json:"merchant"`
	Weight   float64 `json:"weight"`
}

func EvaluateResult(
	transactions []model.Transaction,
	receipts []model.Receipt,
	result model.MatchResult,
	labels []LabelledPair,
) Metrics {
	return EvaluateWithWeights(
		transactions, receipts, result.MatchedOnly(), labels,
		nil, nil, len(result.Conflicts), len(transactions), len(receipts),
	)
}

func Evaluate(
	transactions []model.Transaction,
	receipts []model.Receipt,
	matches []model.Match,
	labels []LabelledPair,
) Metrics {
	return EvaluateWithWeights(transactions, receipts, matches, labels, nil, nil, 0, len(transactions), len(receipts))
}

func EvaluateWithWeights(
	transactions []model.Transaction,
	receipts []model.Receipt,
	matches []model.Match,
	labels []LabelledPair,
	cohortWeights []CohortWeight,
	categoryLabels map[string]string,
	conflicts int,
	txnCount int,
	receiptCount int,
) Metrics {
	itemised := 0
	for _, r := range receipts {
		if r.Itemised() {
			itemised++
		}
	}

	labelMap := make(map[string]string, len(labels))
	for _, l := range labels {
		labelMap[l.TransactionID] = l.ReceiptID
	}

	correct := 0
	falseMatches := 0
	matchedTxn := make(map[string]struct{})
	correctCategories := 0
	categoryTotal := 0

	for _, m := range matches {
		matchedTxn[m.TransactionID] = struct{}{}
		expected, ok := labelMap[m.TransactionID]
		if !ok {
			falseMatches++
			continue
		}
		if expected == m.ReceiptID {
			correct++
		} else {
			falseMatches++
		}
	}

	if categoryLabels != nil {
		receiptByID := make(map[string]model.Receipt, len(receipts))
		for _, r := range receipts {
			receiptByID[r.ID] = r
		}
		for receiptID, expectedCat := range categoryLabels {
			categoryTotal++
			if r, ok := receiptByID[receiptID]; ok && r.Category == expectedCat {
				correctCategories++
			}
		}
	}

	labelledCount := len(labels)
	recallDenom := max(labelledCount, 1)
	precisionDenom := max(len(matches), 1)
	receiptDenom := max(receiptCount, 1)
	txnDenom := max(txnCount, 1)
	conflictDenom := max(txnDenom, 1)

	cohortWeighted := CohortWeightedItemisationRate(receipts, cohortWeights)

	catAcc := 0.0
	if categoryTotal > 0 {
		catAcc = float64(correctCategories) / float64(categoryTotal)
	}

	return Metrics{
		ItemisationRate:           float64(itemised) / float64(receiptDenom),
		CohortWeightedItemisation: cohortWeighted,
		MatchRate:                 float64(len(matches)) / float64(txnDenom),
		FalseMatchRate:            float64(falseMatches) / float64(precisionDenom),
		ConflictRate:              float64(conflicts) / float64(conflictDenom),
		Precision:                 float64(correct) / float64(precisionDenom),
		Recall:                    float64(correct) / float64(recallDenom),
		CategorisationAccuracy:    catAcc,
		Matched:                   len(matches),
		Conflicts:                 conflicts,
		UnmatchedReceipts:         receiptDenom - len(matches),
		UnmatchedTransactions:     txnDenom - len(matchedTxn),
		FalseMatches:              falseMatches,
	}
}

func CohortWeightedItemisationRate(receipts []model.Receipt, weights []CohortWeight) float64 {
	if len(weights) == 0 {
		return 0
	}

	var totalWeight float64
	for _, w := range weights {
		totalWeight += w.Weight
	}
	if totalWeight == 0 {
		return 0
	}

	var covered float64
	for _, w := range weights {
		for _, r := range receipts {
			if r.Supplier == w.Merchant && r.Itemised() {
				covered += w.Weight
				break
			}
		}
	}
	return covered / totalWeight
}
