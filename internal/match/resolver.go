package match

import (
	"fmt"
	"sort"

	"github.com/jimbery/receipt/internal/model"
)

const (
	minAmbiguityCandidates = 2
	signalMatchEpsilon     = 0.01
)

type scoredCandidate struct {
	TransactionID string
	ReceiptID     string
	Confidence    float64
	Signals       map[string]float64
	Method        model.MatchMethod
}

func resolve(cfg Config, candidates []scoredCandidate) model.MatchResult {
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Confidence != candidates[j].Confidence {
			return candidates[i].Confidence > candidates[j].Confidence
		}
		return candidates[i].ReceiptID < candidates[j].ReceiptID
	})

	conflictedReceipt := detectAmbiguity(cfg, candidates, byReceipt)
	conflictedTxn := detectTxnAmbiguity(cfg, candidates, conflictedReceipt)

	usedTxn := make(map[string]struct{})
	usedReceipt := make(map[string]struct{})
	var matches []model.Match
	var conflicts []model.Conflict

	for id := range conflictedTxn {
		conflicts = append(conflicts, model.Conflict{
			TransactionID: id,
			CompetingIDs:  competingReceipts(candidates, id),
			TopConfidence: topConfidenceForTxn(candidates, id),
			Reason:        "ambiguous transaction candidates within margin",
		})
	}
	for id := range conflictedReceipt {
		conflicts = append(conflicts, model.Conflict{
			ReceiptID:     id,
			CompetingIDs:  competingTransactions(candidates, id),
			TopConfidence: topConfidenceForReceipt(candidates, id),
			Reason:        "ambiguous receipt candidates within margin",
		})
	}

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

func byReceipt(c scoredCandidate) string { return c.ReceiptID }

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
			return group[i].Confidence > group[j].Confidence
		})
		if group[0].Confidence-group[1].Confidence <= cfg.AmbiguityMargin &&
			signalIndistinguishable(group[0], group[1]) {
			conflicted[k] = struct{}{}
		}
	}
	return conflicted
}

// detectTxnAmbiguity flags transaction conflicts only when receipt-side competition
// also exists — duplicate receipts for one transaction tiebreak deterministically.
func detectTxnAmbiguity(
	cfg Config,
	candidates []scoredCandidate,
	conflictedReceipt map[string]struct{},
) map[string]struct{} {
	byTxn := make(map[string][]scoredCandidate)
	for _, c := range candidates {
		if c.Confidence < cfg.MinConfidence {
			continue
		}
		byTxn[c.TransactionID] = append(byTxn[c.TransactionID], c)
	}

	conflicted := make(map[string]struct{})
	for txnID, group := range byTxn {
		if len(group) < minAmbiguityCandidates {
			continue
		}
		sort.Slice(group, func(i, j int) bool {
			return group[i].Confidence > group[j].Confidence
		})
		if group[0].Confidence-group[1].Confidence > cfg.AmbiguityMargin ||
			!signalIndistinguishable(group[0], group[1]) {
			continue
		}
		receiptContested := false
		for _, c := range group[:2] {
			if _, ok := conflictedReceipt[c.ReceiptID]; ok {
				receiptContested = true
				break
			}
		}
		if receiptContested {
			conflicted[txnID] = struct{}{}
		}
	}
	return conflicted
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
	return out
}

func topConfidenceForTxn(candidates []scoredCandidate, txnID string) float64 {
	return topConfidence(candidates, func(c scoredCandidate) bool { return c.TransactionID == txnID })
}

func topConfidenceForReceipt(candidates []scoredCandidate, receiptID string) float64 {
	return topConfidence(candidates, func(c scoredCandidate) bool { return c.ReceiptID == receiptID })
}

// signalIndistinguishable returns true when candidates tie on amount, merchant,
// and temporal proximity (ADR D2: surface true ties only).
func signalIndistinguishable(a, b scoredCandidate) bool {
	for _, key := range []string{"amount", "merchant"} {
		if diff := abs(a.Signals[key] - b.Signals[key]); diff > signalMatchEpsilon {
			return false
		}
	}
	const minDeltaSeparation = 60.0 // seconds
	return abs(a.Signals["temporal_delta_secs"]-b.Signals["temporal_delta_secs"]) < minDeltaSeparation
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
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
