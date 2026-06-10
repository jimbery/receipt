package synth

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/jimbery/receipt/internal/model"
)

// NoiseProfile controls adversarial noise applied during generation (M0.4).
type NoiseProfile struct {
	TimezoneShiftSec   int
	SettlementLagHours int
	TipMinor           int64
	PartialCapturePct  float64
	FXMismatchMinor    int64
	MerchantPrefix     string
	StoreNumberSuffix  string
}

// Generator produces seeded, reproducible labelled datasets.
type Generator struct {
	rng  *rand.Rand
	seed uint64
}

func NewGenerator(seed int64) *Generator {
	s := uint64(seed) //nolint:gosec // deterministic synthetic data, not security-sensitive
	return &Generator{
		rng:  rand.New(rand.NewPCG(s, s^0x9e3779b97f4a7c15)), //nolint:gosec // reproducible evaluation datasets
		seed: s,
	}
}

func (g *Generator) Seed() int64 {
	return int64(g.seed) //nolint:gosec // seed fits int64 by construction
}

// Generate builds n matchable pairs plus unmatched populations.
func (g *Generator) Generate(n int, profile NoiseProfile) Dataset {
	if n < 1 {
		n = 1
	}
	base := scenarioBase()
	d := Dataset{
		Seed:  int64(g.seed), //nolint:gosec // seed fits int64 by construction
		Class: ClassGenerated,
		Expectations: model.Expectations{
			TransactionOutcomes: make(map[string]model.Outcome),
			ReceiptOutcomes:     make(map[string]model.Outcome),
		},
	}

	for i := range n {
		id := fmt.Sprintf("gen-%d", i)
		amount := int64(thousand + g.rng.IntN(fiftyThousand))

		txnTime := base.Add(time.Duration(g.rng.IntN(hours72)) * time.Hour)
		if profile.TimezoneShiftSec != 0 {
			txnTime = txnTime.Add(time.Duration(profile.TimezoneShiftSec) * time.Second)
		}

		receiptAmount := amount - profile.TipMinor
		if profile.PartialCapturePct > 0 {
			receiptAmount = int64(float64(amount) * (1 - profile.PartialCapturePct))
		}
		receiptAmount -= profile.FXMismatchMinor

		merchant := "SCREWFIX"
		descriptor := profile.MerchantPrefix + merchant + profile.StoreNumberSuffix
		if descriptor == merchant {
			descriptor = "SCREWFIX 1234 LON"
		}

		lag := time.Duration(profile.SettlementLagHours) * time.Hour
		receiptTime := txnTime.Add(lag + time.Duration(g.rng.IntN(minutes60))*time.Minute)

		txnID := id + "-t"
		recID := id + "-r"

		d.Transactions = append(d.Transactions, model.Transaction{
			ID: txnID, Source: model.TransactionSourceSynthetic,
			Merchant: descriptor, MCC: "5251",
			Amount:     model.NewMoney(amount, "GBP"),
			OccurredAt: model.NewTimestamp(txnTime.UTC(), profile.TimezoneShiftSec),
		})
		d.Receipts = append(d.Receipts, model.Receipt{
			ID: recID, Source: model.ReceiptSourceEmail,
			Supplier: merchant,
			Total:    model.NewMoney(receiptAmount, "GBP"),
			IssuedAt: model.NewTimestamp(receiptTime.UTC(), 0),
		})
		d.Labels = append(d.Labels, model.LabelledPair{TransactionID: txnID, ReceiptID: recID})
		d.Expectations.TransactionOutcomes[txnID] = model.OutcomeMatched
		d.Expectations.ReceiptOutcomes[recID] = model.OutcomeMatched
		d.NoiseBounds = append(d.NoiseBounds, NoiseBound{
			TransactionID: txnID,
			MaxAmountDiffMinor: max64(
				profile.TipMinor+profile.FXMismatchMinor,
				int64(float64(amount)*profile.PartialCapturePct),
			),
			MaxTemporalLagHours: profile.SettlementLagHours + 1,
		})
	}

	d.Transactions = append(d.Transactions, model.Transaction{
		ID: "gen-unmatched-t", Merchant: "ORPHAN TXN", MCC: "5399",
		Amount: model.NewMoney(1999, "GBP"), OccurredAt: model.NewTimestamp(base, 0),
	})
	d.Receipts = append(d.Receipts, model.Receipt{
		ID: "gen-unmatched-r", Supplier: "Orphan Supplier",
		Total: model.NewMoney(2999, "GBP"), IssuedAt: model.NewTimestamp(base, 0),
	})
	d.Expectations.TransactionOutcomes["gen-unmatched-t"] = model.OutcomeUnmatched
	d.Expectations.ReceiptOutcomes["gen-unmatched-r"] = model.OutcomeUnmatched

	return d
}

const (
	thousand      = 1000
	fiftyThousand = 50000
	hours72       = 72
	minutes60     = 60
)

// Dataset is a generated labelled set with scenario-class tagging.
type Dataset struct {
	Seed         int64
	Class        Class
	Transactions []model.Transaction
	Receipts     []model.Receipt
	Labels       []model.LabelledPair
	Expectations model.Expectations
	NoiseBounds  []NoiseBound
}

func (d Dataset) Scenario() Scenario {
	return Scenario{
		Name:         string(d.Class),
		Class:        d.Class,
		Transactions: d.Transactions,
		Receipts:     d.Receipts,
		Labels:       d.Labels,
		Expectations: d.Expectations,
	}
}

// NoiseBound records generator-claimed noise limits for property verification.
type NoiseBound struct {
	TransactionID       string
	MaxAmountDiffMinor  int64
	MaxTemporalLagHours int
}

// VerifyNoiseBounds checks every labelled pair sits within generator-claimed bounds.
func (d Dataset) VerifyNoiseBounds() bool {
	txnByID := make(map[string]model.Transaction, len(d.Transactions))
	recByID := make(map[string]model.Receipt, len(d.Receipts))
	boundByTxn := make(map[string]NoiseBound, len(d.NoiseBounds))
	for _, t := range d.Transactions {
		txnByID[t.ID] = t
	}
	for _, r := range d.Receipts {
		recByID[r.ID] = r
	}
	for _, b := range d.NoiseBounds {
		boundByTxn[b.TransactionID] = b
	}
	for _, l := range d.Labels {
		txn := txnByID[l.TransactionID]
		rec := recByID[l.ReceiptID]
		bound, ok := boundByTxn[l.TransactionID]
		if !ok {
			return false
		}
		diff, ok := txn.Amount.AbsDiff(rec.Total)
		if !ok || diff > bound.MaxAmountDiffMinor+DefaultConfigAbsTol() {
			return false
		}
		lag := rec.IssuedAt.UTC.Sub(txn.OccurredAt.UTC)
		if lag < 0 {
			lag = -lag
		}
		if lag > time.Duration(bound.MaxTemporalLagHours)*time.Hour {
			return false
		}
	}
	return true
}

func DefaultConfigAbsTol() int64 { return 50 }

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
