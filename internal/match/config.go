package match

import (
	"time"

	"github.com/jimbery/receipt/internal/model"
)

// Config holds all tunable matcher parameters (ADR D4).
type Config struct {
	MaxReceiptLead     time.Duration
	MaxReceiptLag      time.Duration
	AbsAmountTolerance int64
	RelAmountTolerance float64
	MinConfidence      float64
	AmbiguityMargin    float64
	WeightAmount       float64
	WeightTemporal     float64
	WeightMerchant     float64
	WeightMCC          float64
	SourceBoost        map[model.ReceiptSource]float64
}

// DefaultConfig returns production-oriented defaults.
func DefaultConfig() Config {
	return Config{
		MaxReceiptLead:     48 * time.Hour,
		MaxReceiptLag:      7 * 24 * time.Hour,
		AbsAmountTolerance: 50,
		RelAmountTolerance: 0.02,
		MinConfidence:      0.75,
		AmbiguityMargin:    0.03,
		WeightAmount:       0.42,
		WeightTemporal:     0.23,
		WeightMerchant:     0.30,
		WeightMCC:          0.05,
		SourceBoost: map[model.ReceiptSource]float64{
			model.ReceiptSourceEmail: 1.0,
			model.ReceiptSourceOCR:   1.0,
			model.ReceiptSourcePOS:   1.0,
		},
	}
}
