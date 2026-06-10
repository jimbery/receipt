package match

import (
	"math"

	"github.com/jimbery/receipt/internal/model"
)

const (
	tipAbsMinorDefault            int64   = 500
	tipRelDefault                 float64 = 0.15
	partialCaptureAbsMinorDefault int64   = 2000
	partialCaptureRelDefault      float64 = 0.20
	fuelPreAuthRelDefault         float64 = 0.20
	fxRoundingMinorDefault        int64   = 2
	cashbackAbsMinorDefault       int64   = 500
	fxRoundingScore               float64 = 0.95
	cashbackRelDecay              float64 = 0.05
)

// AmountBands configures tolerance bands for amount scoring (M0.2).
type AmountBands struct {
	TipAbsMinor            int64   `json:"TipAbsMinor"`
	TipRel                 float64 `json:"TipRel"`
	PartialCaptureAbsMinor int64   `json:"PartialCaptureAbsMinor"`
	PartialCaptureRel      float64 `json:"PartialCaptureRel"`
	FuelPreAuthRel         float64 `json:"FuelPreAuthRel"`
	FXRoundingMinor        int64   `json:"FXRoundingMinor"`
	CashbackAbsMinor       int64   `json:"CashbackAbsMinor"`
}

func defaultAmountBands() AmountBands {
	return AmountBands{
		TipAbsMinor:            tipAbsMinorDefault,
		TipRel:                 tipRelDefault,
		PartialCaptureAbsMinor: partialCaptureAbsMinorDefault,
		PartialCaptureRel:      partialCaptureRelDefault,
		FuelPreAuthRel:         fuelPreAuthRelDefault,
		FXRoundingMinor:        fxRoundingMinorDefault,
		CashbackAbsMinor:       cashbackAbsMinorDefault,
	}
}

func amountSimilarity(cfg Config, txn, receipt model.Money) float64 {
	if txn.Currency != receipt.Currency {
		return 0
	}
	if txn.Equal(receipt) {
		return 1
	}

	diff, ok := txn.AbsDiff(receipt)
	if !ok {
		return 0
	}

	bands := cfg.AmountBands
	larger := max(txn.Amount, receipt.Amount)
	if larger == 0 {
		return 1
	}

	// Tip / gratuity: receipt <= transaction.
	if receipt.Amount <= txn.Amount && withinBand(diff, larger, bands.TipAbsMinor, bands.TipRel) {
		return decayScore(diff, larger, bands.TipAbsMinor, bands.TipRel)
	}

	// Partial capture / fuel pre-auth: transaction may exceed receipt.
	if txn.Amount >= receipt.Amount {
		maxDiff := max64(bands.PartialCaptureAbsMinor, int64(float64(larger)*bands.PartialCaptureRel))
		if bands.FuelPreAuthRel > bands.PartialCaptureRel {
			fuelMax := int64(float64(larger) * bands.FuelPreAuthRel)
			if fuelMax > maxDiff {
				maxDiff = fuelMax
			}
		}
		if diff <= maxDiff {
			return decayScore(diff, larger, maxDiff, bands.PartialCaptureRel)
		}
	}

	// FX rounding delta.
	if diff <= bands.FXRoundingMinor {
		return fxRoundingScore
	}

	// Cashback: receipt may exceed transaction slightly.
	if receipt.Amount > txn.Amount && diff <= bands.CashbackAbsMinor {
		return decayScore(diff, larger, bands.CashbackAbsMinor, cashbackRelDecay)
	}

	// Legacy tight tolerance for small skew.
	if txn.WithinTolerance(receipt, cfg.AbsAmountTolerance, cfg.RelAmountTolerance) {
		maxDiff := float64(cfg.AbsAmountTolerance)
		if rel := cfg.RelAmountTolerance * float64(larger); rel > maxDiff {
			maxDiff = rel
		}
		return math.Max(0, 1-float64(diff)/maxDiff*0.25)
	}

	return 0
}

func withinBand(diff, larger, absMinor int64, rel float64) bool {
	if diff <= absMinor {
		return true
	}
	return float64(diff)/float64(larger) <= rel
}

func decayScore(diff, larger, absMinor int64, rel float64) float64 {
	maxDiff := float64(absMinor)
	if r := rel * float64(larger); r > maxDiff {
		maxDiff = r
	}
	if maxDiff == 0 {
		return 1
	}
	return math.Max(0, 1-float64(diff)/maxDiff*0.2)
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
