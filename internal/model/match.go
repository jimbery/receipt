package model

// Outcome is the resolution state for a transaction or receipt.
type Outcome string

const (
	OutcomeMatched   Outcome = "matched"
	OutcomeUnmatched Outcome = "unmatched"
	OutcomeConflict  Outcome = "conflict"
)

// MatchMethod describes how a match was produced.
type MatchMethod string

const (
	MatchMethodScored   MatchMethod = "scored"
	MatchMethodExactRef MatchMethod = "exact_ref"
)

// Match is a confident 1:1 link between a transaction and receipt.
type Match struct {
	TransactionID string
	ReceiptID     string
	Outcome       Outcome
	Confidence    float64
	Method        MatchMethod
	Signals       map[string]float64
}

// Conflict records ambiguity the engine refused to resolve.
type Conflict struct {
	TransactionID string
	ReceiptID     string
	CompetingIDs  []string
	TopConfidence float64
	Reason        string
}

// MatchResult is the full deterministic output of the matching engine.
type MatchResult struct {
	Matches             []Match
	UnmatchedTxnIDs     []string
	UnmatchedReceiptIDs []string
	Conflicts           []Conflict
}

// MatchedOnly returns matches with OutcomeMatched for harness evaluation.
func (r MatchResult) MatchedOnly() []Match {
	out := make([]Match, 0, len(r.Matches))
	for _, m := range r.Matches {
		if m.Outcome == OutcomeMatched {
			out = append(out, m)
		}
	}
	return out
}
