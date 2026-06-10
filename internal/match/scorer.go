package match

import (
	"math"
	"time"

	"github.com/jimbery/receipt/internal/model"
)

const temporalDecayFactor = 0.35

type candidate struct {
	Transaction model.Transaction
	Receipt     model.Receipt
}

func (e *Engine) scoreCandidate(c candidate) (float64, map[string]float64) {
	cfg := e.cfg

	amountScore := amountSimilarity(cfg, c.Transaction.Amount, c.Receipt.Total)
	temporalScore, temporalDelta := temporalSimilarity(cfg, c.Transaction.OccurredAt.UTC, c.Receipt.IssuedAt.UTC)
	merchantScore := e.merchant.Similarity(
		c.Transaction.Merchant,
		c.Receipt.Supplier,
		c.Receipt.RawMerchantHint,
	)
	mccScore := mccCorroboration(c.Transaction.MCC, c.Receipt.Supplier, c.Receipt.RawMerchantHint)

	signals := map[string]float64{
		"amount":              amountScore,
		"temporal":            temporalScore,
		"temporal_delta_secs": temporalDelta,
		"merchant":            merchantScore,
		"mcc":                 mccScore,
	}

	confidence := cfg.WeightAmount*amountScore +
		cfg.WeightTemporal*temporalScore +
		cfg.WeightMerchant*merchantScore +
		cfg.WeightMCC*mccScore

	if boost, ok := cfg.SourceBoost[c.Receipt.Source]; ok && boost > 0 && boost != 1 {
		confidence *= boost
	}
	if confidence > 1 {
		confidence = 1
	}

	return confidence, signals
}

func amountSimilarity(cfg Config, txn, receipt model.Money) float64 {
	if txn.Currency != receipt.Currency {
		return 0
	}
	if txn.Equal(receipt) {
		return 1
	}
	if txn.WithinTolerance(receipt, cfg.AbsAmountTolerance, cfg.RelAmountTolerance) {
		diff, _ := txn.AbsDiff(receipt)
		larger := max(txn.Amount, receipt.Amount)
		if larger == 0 {
			return 1
		}
		maxDiff := float64(cfg.AbsAmountTolerance)
		if rel := cfg.RelAmountTolerance * float64(larger); rel > maxDiff {
			maxDiff = rel
		}
		return math.Max(0, 1-float64(diff)/maxDiff*0.25)
	}
	return 0
}

func temporalSimilarity(cfg Config, txnTime, receiptTime time.Time) (float64, float64) {
	delta := receiptTime.Sub(txnTime)
	window := cfg.MaxReceiptLag
	if delta < 0 {
		delta = -delta
		window = cfg.MaxReceiptLead
	}
	deltaSecs := delta.Seconds()
	if delta > window {
		return 0, deltaSecs
	}
	score := 1 - float64(delta)/float64(window)*temporalDecayFactor
	return score, deltaSecs
}
