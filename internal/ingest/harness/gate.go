package harness

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/jimbery/receipt/internal/classify"
	"github.com/jimbery/receipt/internal/emailtypes"
	"github.com/jimbery/receipt/internal/model"
)

// CorpusMessage is a labelled fixture for classifier evaluation.
type CorpusMessage struct {
	ID      string                  `json:"id"`
	Msg     emailtypes.RawMessage   `json:"msg"`
	Label   emailtypes.DocumentKind `json:"label"`
	Stratum string                  `json:"stratum"`
	Pattern string                  `json:"pattern"`
}

// FieldLabel is ground truth for extractor evaluation.
type FieldLabel struct {
	MessageID  string `json:"message_id"`
	Supplier   string `json:"supplier"`
	TotalMinor int64  `json:"total_minor"`
	OrderRef   string `json:"order_ref"`
	Category   string `json:"category"`
}

// ClassifierMetrics holds exact-match and receipt-bearing scores.
type ClassifierMetrics struct {
	ExactPrecision          float64
	ExactRecall             float64
	ReceiptBearingPrecision float64
	ReceiptBearingRecall    float64
}

func isReceiptBearing(k emailtypes.DocumentKind) bool {
	switch k {
	case emailtypes.KindPurchaseReceipt, emailtypes.KindCreditNote, emailtypes.KindDispatchNotice:
		return true
	case emailtypes.KindMarketing, emailtypes.KindStatement, emailtypes.KindUnknown:
		return false
	}
	return false
}

// EvaluateClassifier scores exact DocumentKind match (strict; unknown is never a match).
func EvaluateClassifier(corpus []CorpusMessage, clf *classify.Classifier) (float64, float64) {
	m := EvaluateClassifierMetrics(corpus, clf)
	return m.ExactPrecision, m.ExactRecall
}

// EvaluateClassifierMetrics reports exact and receipt-bearing scores separately.
func EvaluateClassifierMetrics(corpus []CorpusMessage, clf *classify.Classifier) ClassifierMetrics {
	var exactTP, exactFP, exactFN int
	var rbTP, rbFP, rbFN int

	for _, c := range corpus {
		got := clf.Classify(c.Msg)
		if got == c.Label {
			exactTP++
		} else {
			exactFP++
			exactFN++
		}

		wantRB := isReceiptBearing(c.Label)
		gotRB := isReceiptBearing(got)
		switch {
		case wantRB && gotRB:
			rbTP++
		case wantRB && !gotRB:
			rbFN++
		case !wantRB && gotRB:
			rbFP++
		}
	}

	return ClassifierMetrics{
		ExactPrecision:          ratio(exactTP, exactTP+exactFP),
		ExactRecall:             ratio(exactTP, exactTP+exactFN),
		ReceiptBearingPrecision: ratio(rbTP, rbTP+rbFP),
		ReceiptBearingRecall:    ratio(rbTP, rbTP+rbFN),
	}
}

func ratio(num, denom int) float64 {
	if denom == 0 {
		return 0
	}
	return float64(num) / float64(denom)
}

// EvaluateFields scores supplier/total accuracy on labelled extractions.
func EvaluateFields(labels []FieldLabel, receipts []model.Receipt) float64 {
	if len(labels) == 0 {
		return 1
	}
	byID := make(map[string]model.Receipt, len(receipts))
	for _, r := range receipts {
		byID[r.ID] = r
	}
	correct := 0
	for _, l := range labels {
		r, ok := byID[l.MessageID]
		if !ok {
			continue
		}
		if r.Supplier == l.Supplier && r.Total.Amount == l.TotalMinor {
			correct++
		}
	}
	return float64(correct) / float64(len(labels))
}

func hashCorpus(corpus []CorpusMessage) string {
	ids := make([]string, len(corpus))
	for i, c := range corpus {
		ids[i] = c.ID + "|" + string(c.Label)
	}
	b, _ := json.Marshal(ids)
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:8])
}
