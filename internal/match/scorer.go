package match

import (
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

	amountScore := amountSimilarity(cfg, c.Transaction.Amount, c.Receipt.Total) // bands in amount.go
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
