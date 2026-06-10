package match

import (
	"fmt"
	"math"
	"sort"

	"github.com/jimbery/receipt/internal/model"
)

const (
	minAmbiguityCandidates   = 2
	signalDistinguishEpsilon = 0.01
	temporalDistinguishSecs  = 120.0
)

type scoredCandidate struct {
	TransactionID string
	ReceiptID     string
	Confidence    float64
	Signals       map[string]float64
	Method        model.MatchMethod
}

func compareCandidates(a, b scoredCandidate) bool {
	if a.Confidence != b.Confidence {
		return a.Confidence > b.Confidence
	}
	if a.TransactionID != b.TransactionID {
		return a.TransactionID < b.TransactionID
	}
	return a.ReceiptID < b.ReceiptID
}

func resolve(cfg Config, candidates []scoredCandidate) model.MatchResult {
	sort.Slice(candidates, func(i, j int) bool {
		return compareCandidates(candidates[i], candidates[j])
	})

	conflictedTxn := detectAmbiguity(cfg, candidates, byTransaction)
	conflictedReceipt := detectAmbiguity(cfg, candidates, byReceipt)
	expandReceiptConflicts(candidates, conflictedTxn, conflictedReceipt)

	conflicts := buildConflicts(candidates, conflictedTxn, conflictedReceipt)

	usedTxn := make(map[string]struct{})
	usedReceipt := make(map[string]struct{})
	var matches []model.Match

	for _, c := range candidates {
		if c.Confidence < cfg.MinConfidence {
			continue
		}
		if _, ok := conflictedTxn[c.TransactionID]; ok {
			continue
		}
		if _, ok := conflictedReceipt[c.ReceiptID]; ok {
			continue
		}
		if _, ok := usedTxn[c.TransactionID]; ok {
			continue
		}
		if _, ok := usedReceipt[c.ReceiptID]; ok {
			continue
		}

		usedTxn[c.TransactionID] = struct{}{}
		usedReceipt[c.ReceiptID] = struct{}{}
		matches = append(matches, model.Match{
			TransactionID: c.TransactionID,
			ReceiptID:     c.ReceiptID,
			Outcome:       model.OutcomeMatched,
			Confidence:    c.Confidence,
			Method:        c.Method,
			Signals:       c.Signals,
		})
	}

	return model.MatchResult{Matches: matches, Conflicts: conflicts}
}

type groupKey func(scoredCandidate) string

func byTransaction(c scoredCandidate) string { return c.TransactionID }
func byReceipt(c scoredCandidate) string     { return c.ReceiptID }

func detectAmbiguity(cfg Config, candidates []scoredCandidate, keyFn groupKey) map[string]struct{} {
	byKey := make(map[string][]scoredCandidate)
	for _, c := range candidates {
		if c.Confidence < cfg.MinConfidence {
			continue
		}
		k := keyFn(c)
		byKey[k] = append(byKey[k], c)
	}

	conflicted := make(map[string]struct{})
	for k, group := range byKey {
		if len(group) < minAmbiguityCandidates {
			continue
		}
		sort.Slice(group, func(i, j int) bool {
			return compareCandidates(group[i], group[j])
		})
		if group[0].Confidence-group[1].Confidence <= cfg.AmbiguityMargin &&
			!candidatesDistinguishable(group[0], group[1]) {
			conflicted[k] = struct{}{}
		}
	}
	return conflicted
}

// candidatesDistinguishable reports whether two candidates differ on a primary signal
// enough to resolve without emitting a conflict (e.g. near-duplicate temporal separation).
func candidatesDistinguishable(a, b scoredCandidate) bool {
	if math.Abs(a.Signals["amount"]-b.Signals["amount"]) > signalDistinguishEpsilon {
		return true
	}
	if math.Abs(a.Signals["merchant"]-b.Signals["merchant"]) > signalDistinguishEpsilon {
		return true
	}
	if math.Abs(a.Signals["temporal_delta_secs"]-b.Signals["temporal_delta_secs"]) > temporalDistinguishSecs {
		return true
	}
	return false
}

// expandReceiptConflicts marks receipts tied to an ambiguous transaction as conflicted
// (e.g. duplicate receipts forwarded for the same card transaction).
func expandReceiptConflicts(
	candidates []scoredCandidate,
	conflictedTxn map[string]struct{},
	conflictedReceipt map[string]struct{},
) {
	for txnID := range conflictedTxn {
		seen := make(map[string]struct{})
		for _, c := range candidates {
			if c.TransactionID != txnID {
				continue
			}
			seen[c.ReceiptID] = struct{}{}
		}
		if len(seen) < minAmbiguityCandidates {
			continue
		}
		for rid := range seen {
			conflictedReceipt[rid] = struct{}{}
		}
	}
}

func buildConflicts(
	candidates []scoredCandidate,
	conflictedTxn, conflictedReceipt map[string]struct{},
) []model.Conflict {
	var txnIDs, receiptIDs []string
	for id := range conflictedTxn {
		txnIDs = append(txnIDs, id)
	}
	for id := range conflictedReceipt {
		receiptIDs = append(receiptIDs, id)
	}
	sort.Strings(txnIDs)
	sort.Strings(receiptIDs)

	conflicts := make([]model.Conflict, 0, len(txnIDs)+len(receiptIDs))
	for _, id := range txnIDs {
		conflicts = append(conflicts, model.Conflict{
			TransactionID: id,
			CompetingIDs:  competingReceipts(candidates, id),
			TopConfidence: topConfidenceForTxn(candidates, id),
			Reason:        "ambiguous transaction candidates within margin",
		})
	}
	for _, id := range receiptIDs {
		conflicts = append(conflicts, model.Conflict{
			ReceiptID:     id,
			CompetingIDs:  competingTransactions(candidates, id),
			TopConfidence: topConfidenceForReceipt(candidates, id),
			Reason:        "ambiguous receipt candidates within margin",
		})
	}
	return conflicts
}

func competingReceipts(candidates []scoredCandidate, txnID string) []string {
	return competing(
		candidates,
		func(c scoredCandidate) bool { return c.TransactionID == txnID },
		func(c scoredCandidate) string { return c.ReceiptID },
	)
}

func competingTransactions(candidates []scoredCandidate, receiptID string) []string {
	return competing(
		candidates,
		func(c scoredCandidate) bool { return c.ReceiptID == receiptID },
		func(c scoredCandidate) string { return c.TransactionID },
	)
}

func competing(
	candidates []scoredCandidate,
	match func(scoredCandidate) bool,
	id func(scoredCandidate) string,
) []string {
	var out []string
	for _, c := range candidates {
		if match(c) {
			out = append(out, fmt.Sprintf("%s:%.3f", id(c), c.Confidence))
		}
	}
	sort.Strings(out)
	return out
}

func topConfidenceForTxn(candidates []scoredCandidate, txnID string) float64 {
	return topConfidence(candidates, func(c scoredCandidate) bool { return c.TransactionID == txnID })
}

func topConfidenceForReceipt(candidates []scoredCandidate, receiptID string) float64 {
	return topConfidence(candidates, func(c scoredCandidate) bool { return c.ReceiptID == receiptID })
}

func topConfidence(candidates []scoredCandidate, match func(scoredCandidate) bool) float64 {
	best := 0.0
	for _, c := range candidates {
		if match(c) && c.Confidence > best {
			best = c.Confidence
		}
	}
	return best
}

func finalizeResult(
	result model.MatchResult,
	transactions []model.Transaction,
	receipts []model.Receipt,
) model.MatchResult {
	matchedTxn := make(map[string]struct{}, len(result.Matches))
	matchedReceipt := make(map[string]struct{}, len(result.Matches))
	conflictTxn := make(map[string]struct{})
	conflictReceipt := make(map[string]struct{})

	for _, m := range result.Matches {
		matchedTxn[m.TransactionID] = struct{}{}
		matchedReceipt[m.ReceiptID] = struct{}{}
	}
	for _, c := range result.Conflicts {
		if c.TransactionID != "" {
			conflictTxn[c.TransactionID] = struct{}{}
		}
		if c.ReceiptID != "" {
			conflictReceipt[c.ReceiptID] = struct{}{}
		}
	}

	for _, t := range transactions {
		if _, ok := matchedTxn[t.ID]; ok {
			continue
		}
		if _, ok := conflictTxn[t.ID]; ok {
			continue
		}
		result.UnmatchedTxnIDs = append(result.UnmatchedTxnIDs, t.ID)
	}
	for _, r := range receipts {
		if _, ok := matchedReceipt[r.ID]; ok {
			continue
		}
		if _, ok := conflictReceipt[r.ID]; ok {
			continue
		}
		result.UnmatchedReceiptIDs = append(result.UnmatchedReceiptIDs, r.ID)
	}

	return result
}
