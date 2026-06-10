package model

// LabelledPair is ground-truth for evaluation.
type LabelledPair struct {
	TransactionID string `json:"transaction_id"`
	ReceiptID     string `json:"receipt_id"`
}

// Expectations records ground-truth outcome classes for M0.3/M0.5 evaluation.
type Expectations struct {
	TransactionOutcomes          map[string]Outcome
	ReceiptOutcomes              map[string]Outcome
	GenuinelyAmbiguousTxnIDs     []string
	GenuinelyAmbiguousReceiptIDs []string
}

func (e Expectations) AmbiguousTxnSet() map[string]struct{} {
	return toIDSet(e.GenuinelyAmbiguousTxnIDs)
}

func (e Expectations) AmbiguousReceiptSet() map[string]struct{} {
	return toIDSet(e.GenuinelyAmbiguousReceiptIDs)
}

func toIDSet(ids []string) map[string]struct{} {
	m := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		m[id] = struct{}{}
	}
	return m
}

// ValidateOutcomes checks engine output against ground-truth outcome classes.
func ValidateOutcomes(result MatchResult, exp Expectations) []string {
	var violations []string

	matchedTxn := map[string]string{}
	matchedReceipt := map[string]string{}
	for _, m := range result.Matches {
		matchedTxn[m.TransactionID] = m.ReceiptID
		matchedReceipt[m.ReceiptID] = m.TransactionID
	}

	conflictTxn := map[string]struct{}{}
	conflictReceipt := map[string]struct{}{}
	for _, c := range result.Conflicts {
		if c.TransactionID != "" {
			conflictTxn[c.TransactionID] = struct{}{}
		}
		if c.ReceiptID != "" {
			conflictReceipt[c.ReceiptID] = struct{}{}
		}
	}

	check := func(id string, expected Outcome, actual func() Outcome) {
		if got := actual(); got != expected {
			violations = append(violations, id+": want "+string(expected)+" got "+string(got))
		}
	}

	for id, want := range exp.TransactionOutcomes {
		check(id, want, func() Outcome {
			if _, ok := conflictTxn[id]; ok {
				return OutcomeConflict
			}
			if _, ok := matchedTxn[id]; ok {
				return OutcomeMatched
			}
			return OutcomeUnmatched
		})
	}
	for id, want := range exp.ReceiptOutcomes {
		check(id, want, func() Outcome {
			if _, ok := conflictReceipt[id]; ok {
				return OutcomeConflict
			}
			if _, ok := matchedReceipt[id]; ok {
				return OutcomeMatched
			}
			return OutcomeUnmatched
		})
	}

	return violations
}
