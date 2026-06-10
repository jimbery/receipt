package match

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/jimbery/receipt/internal/model"
)

// Config holds all tunable matcher parameters (ADR D4).
type Config struct {
	MaxReceiptLead          time.Duration                   `json:"MaxReceiptLead"`
	MaxReceiptLag           time.Duration                   `json:"MaxReceiptLag"`
	AbsAmountTolerance      int64                           `json:"AbsAmountTolerance"`
	RelAmountTolerance      float64                         `json:"RelAmountTolerance"`
	AmountBands             AmountBands                     `json:"AmountBands"`
	MinConfidence           float64                         `json:"MinConfidence"`
	AmbiguityMargin         float64                         `json:"AmbiguityMargin"`
	MinMerchantForCandidate float64                         `json:"MinMerchantForCandidate"`
	WeightAmount            float64                         `json:"WeightAmount"`
	WeightTemporal          float64                         `json:"WeightTemporal"`
	WeightMerchant          float64                         `json:"WeightMerchant"`
	WeightMCC               float64                         `json:"WeightMCC"`
	SourceBoost             map[model.ReceiptSource]float64 `json:"SourceBoost"`
}

// DefaultConfig returns the committed Phase 0 default configuration.
func DefaultConfig() Config {
	return Config{
		MaxReceiptLead:          48 * time.Hour,
		MaxReceiptLag:           7 * 24 * time.Hour,
		AbsAmountTolerance:      50,
		RelAmountTolerance:      0.02,
		AmountBands:             defaultAmountBands(),
		MinConfidence:           0.82,
		AmbiguityMargin:         0.02,
		MinMerchantForCandidate: 0.55,
		WeightAmount:            0.42,
		WeightTemporal:          0.23,
		WeightMerchant:          0.30,
		WeightMCC:               0.05,
		SourceBoost: map[model.ReceiptSource]float64{
			model.ReceiptSourceEmail: 1.0,
			model.ReceiptSourceOCR:   1.0,
			model.ReceiptSourcePOS:   1.0,
		},
	}
}

// Hash returns a stable SHA-256 hash of the serialisable config fields.
func (c Config) Hash() string {
	b, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
