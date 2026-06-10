package match

import (
	"time"

	"github.com/jimbery/receipt/internal/model"
)

const (
	candidateAbsToleranceMultiplier = 4
	candidateRelToleranceMultiplier = 2
)

// Engine is a deterministic, source-agnostic matching pipeline (ADR-001).
type Engine struct {
	cfg      Config
	merchant MerchantComparer
}

func NewEngine(cfg Config) *Engine {
	return &Engine{cfg: cfg, merchant: BasicMerchantComparer{}}
}

func NewEngineWithComparer(cfg Config, merchant MerchantComparer) *Engine {
	if merchant == nil {
		merchant = BasicMerchantComparer{}
	}
	return &Engine{cfg: cfg, merchant: merchant}
}

func (e *Engine) Config() Config {
	return e.cfg
}

// Match runs candidate generation → scoring → resolution and returns matched,
// unmatched, and conflict outcomes (ADR D2, D3).
func (e *Engine) Match(transactions []model.Transaction, receipts []model.Receipt) model.MatchResult {
	exact := e.exactRefMatches(transactions, receipts)
	exactTxn := make(map[string]struct{}, len(exact))
	exactReceipt := make(map[string]struct{}, len(exact))
	for _, m := range exact {
		exactTxn[m.TransactionID] = struct{}{}
		exactReceipt[m.ReceiptID] = struct{}{}
	}

	scored := e.scoreCandidates(transactions, receipts, exactTxn, exactReceipt)
	result := resolve(e.cfg, scored)
	result.Matches = append(exact, result.Matches...)
	return finalizeResult(result, transactions, receipts)
}

func (e *Engine) exactRefMatches(transactions []model.Transaction, receipts []model.Receipt) []model.Match {
	var matches []model.Match
	for _, rec := range receipts {
		if rec.Source != model.ReceiptSourcePOS || rec.TransactionRef == "" {
			continue
		}
		for _, txn := range transactions {
			if txn.ExternalRef == rec.TransactionRef || txn.ID == rec.TransactionRef {
				matches = append(matches, model.Match{
					TransactionID: txn.ID,
					ReceiptID:     rec.ID,
					Outcome:       model.OutcomeMatched,
					Confidence:    1,
					Method:        model.MatchMethodExactRef,
					Signals:       map[string]float64{"exact_ref": 1},
				})
				break
			}
		}
	}
	return matches
}

func (e *Engine) scoreCandidates(
	transactions []model.Transaction,
	receipts []model.Receipt,
	skipTxn, skipReceipt map[string]struct{},
) []scoredCandidate {
	candidates := e.generateCandidates(transactions, receipts, skipTxn, skipReceipt)

	scored := make([]scoredCandidate, 0, len(candidates))
	for _, c := range candidates {
		confidence, signals := e.scoreCandidate(c)
		if confidence <= 0 {
			continue
		}
		scored = append(scored, scoredCandidate{
			TransactionID: c.Transaction.ID,
			ReceiptID:     c.Receipt.ID,
			Confidence:    confidence,
			Signals:       signals,
			Method:        model.MatchMethodScored,
		})
	}
	return scored
}

func (e *Engine) generateCandidates(
	transactions []model.Transaction,
	receipts []model.Receipt,
	skipTxn, skipReceipt map[string]struct{},
) []candidate {
	var out []candidate
	for _, txn := range transactions {
		if _, skip := skipTxn[txn.ID]; skip {
			continue
		}
		for _, rec := range receipts {
			if _, skip := skipReceipt[rec.ID]; skip {
				continue
			}
			if !e.withinTemporalWindow(txn.OccurredAt.UTC, rec.IssuedAt.UTC) {
				continue
			}
			if txn.Amount.Currency != rec.Total.Currency {
				continue
			}
			if !txn.Amount.WithinTolerance(
				rec.Total,
				e.cfg.AbsAmountTolerance*candidateAbsToleranceMultiplier,
				e.cfg.RelAmountTolerance*candidateRelToleranceMultiplier,
			) {
				continue
			}
			out = append(out, candidate{Transaction: txn, Receipt: rec})
		}
	}
	return out
}

func (e *Engine) withinTemporalWindow(txnTime, receiptTime time.Time) bool {
	if receiptTime.Before(txnTime) {
		return txnTime.Sub(receiptTime) <= e.cfg.MaxReceiptLead
	}
	return receiptTime.Sub(txnTime) <= e.cfg.MaxReceiptLag
}
