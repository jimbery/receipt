package synth

import (
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/jimbery/receipt/internal/model"
)

// NoiseProfile controls adversarial noise applied during generation (M0.4).
type NoiseProfile struct {
	Class              Class
	TimezoneShiftSec   int
	SettlementLagHours int
	TipMinor           int64
	PartialCapturePct  float64
	FXMismatchMinor    int64
	CashbackMinor      int64
	MerchantPrefix     string
	StoreNumberSuffix  string
	MerchantIndex      int
}

// Generator produces seeded, reproducible labelled datasets.
type Generator struct {
	rng  *rand.Rand
	seed uint64
}

func NewGenerator(seed int64) *Generator {
	s := uint64(seed) //nolint:gosec // deterministic synthetic data
	return &Generator{
		rng:  rand.New(rand.NewPCG(s, s^0x9e3779b97f4a7c15)), //nolint:gosec
		seed: s,
	}
}

func (g *Generator) Seed() int64 {
	return int64(g.seed) //nolint:gosec
}

var cohortMerchants = []struct {
	supplier   string
	descriptor string
	mcc        string
}{
	{"Screwfix", "SCREWFIX 1234 LON", "5251"},
	{"Toolstation", "TOOLSTATION 882", "5251"},
	{"Amazon", "AMAZON UK MARKETPLACE", "5399"},
	{"Shell", "SHELL FUEL GB", "5541"},
	{"B&Q", "B AND Q 4421", "5251"},
	{"BP", "BP CONNECT 991", "5541"},
}

// GenerateSuite builds n pairs split across noise-profile scenario classes.
func (g *Generator) GenerateSuite(n int) []Scenario {
	if n < 4 {
		n = 4
	}
	q := n / 4
	rem := n % 4
	counts := []int{q, q, q, q}
	for i := range rem {
		counts[i]++
	}
	return []Scenario{
		g.Generate(counts[0], tipProfile()),
		g.Generate(counts[1], settlementProfile()),
		g.Generate(counts[2], mangledProfile()),
		g.Generate(counts[3], fxProfile()),
		g.generateAmbiguousCluster(4),
		g.generateNearDuplicates(4),
		g.generateRefunds(2),
	}
}

func tipProfile() NoiseProfile {
	return NoiseProfile{Class: ClassGeneratedTip, TipMinor: 150}
}

// SettlementProfile returns the settlement-delay noise profile for tests and tooling.
func SettlementProfile() NoiseProfile {
	return NoiseProfile{Class: ClassGeneratedSettlement, SettlementLagHours: 48}
}

func settlementProfile() NoiseProfile { return SettlementProfile() }

func mangledProfile() NoiseProfile {
	return NoiseProfile{Class: ClassGeneratedMangled, MerchantPrefix: "SQ *", StoreNumberSuffix: " 77"}
}

func fxProfile() NoiseProfile {
	return NoiseProfile{Class: ClassGeneratedFX, FXMismatchMinor: 1}
}

// Generate builds matchable pairs for one scenario class.
func (g *Generator) Generate(n int, profile NoiseProfile) Scenario {
	if n < 1 {
		n = 1
	}
	if profile.Class == "" {
		profile.Class = ClassGenerated
	}
	base := scenarioBase()
	d := Scenario{
		Name:  string(profile.Class),
		Class: profile.Class,
		Expectations: model.Expectations{
			TransactionOutcomes: make(map[string]model.Outcome),
			ReceiptOutcomes:     make(map[string]model.Outcome),
		},
	}

	for i := range n {
		id := fmt.Sprintf("%s-%d", profile.Class, i)
		m := cohortMerchants[(profile.MerchantIndex+i)%len(cohortMerchants)]
		// Space amounts >25% apart so tolerance bands cannot cross-match distinct pairs.
		amount := 100000 + int64(i)*30000

		// Prime hour spacing avoids settlement-lag periodic collisions at scale.
		txnTime := base.Add(time.Duration(i) * 37 * time.Hour)
		if profile.TimezoneShiftSec != 0 {
			txnTime = txnTime.Add(time.Duration(profile.TimezoneShiftSec) * time.Second)
		}

		receiptAmount := amount - profile.TipMinor + profile.CashbackMinor
		if profile.PartialCapturePct > 0 {
			receiptAmount = int64(float64(amount) * (1 - profile.PartialCapturePct))
		}
		receiptAmount -= profile.FXMismatchMinor

		descriptor := profile.MerchantPrefix + m.descriptor + profile.StoreNumberSuffix
		lag := time.Duration(profile.SettlementLagHours) * time.Hour
		receiptTime := txnTime.Add(lag + time.Duration(g.rng.IntN(minutes60))*time.Minute)

		txnID := id + "-t"
		recID := id + "-r"

		d.Transactions = append(d.Transactions, model.Transaction{
			ID: txnID, Source: model.TransactionSourceSynthetic,
			Merchant: descriptor, MCC: m.mcc,
			Amount:     model.NewMoney(amount, "GBP"),
			OccurredAt: model.NewTimestamp(txnTime.UTC(), profile.TimezoneShiftSec),
		})
		d.Receipts = append(d.Receipts, model.Receipt{
			ID: recID, Source: model.ReceiptSourceEmail,
			Supplier: m.supplier,
			Total:    model.NewMoney(receiptAmount, "GBP"),
			IssuedAt: model.NewTimestamp(receiptTime.UTC(), 0),
		})
		d.Labels = append(d.Labels, model.LabelledPair{TransactionID: txnID, ReceiptID: recID})
		d.Expectations.TransactionOutcomes[txnID] = model.OutcomeMatched
		d.Expectations.ReceiptOutcomes[recID] = model.OutcomeMatched
		d.NoiseBounds = append(d.NoiseBounds, NoiseBound{
			TransactionID: txnID,
			MaxAmountDiffMinor: max64(
				profile.TipMinor+profile.FXMismatchMinor+profile.CashbackMinor,
				int64(float64(amount)*profile.PartialCapturePct),
			),
			MaxTemporalLagHours: profile.SettlementLagHours + 2,
		})
	}

	return d
}

func (g *Generator) generateAmbiguousCluster(n int) Scenario {
	ts := model.NewTimestamp(scenarioBase().Add(200*time.Hour), 0)
	s := Scenario{
		Name:  "generated_ambiguous",
		Class: ClassGeneratedAmbiguous,
		Expectations: model.Expectations{
			GenuinelyAmbiguousTxnIDs:     make([]string, 0, n),
			GenuinelyAmbiguousReceiptIDs: make([]string, 0, n),
			TransactionOutcomes:          make(map[string]model.Outcome),
			ReceiptOutcomes:              make(map[string]model.Outcome),
		},
	}
	for i := range n {
		tid := fmt.Sprintf("g-amb-t-%d", i)
		rid := fmt.Sprintf("g-amb-r-%d", i)
		s.Transactions = append(s.Transactions, model.Transaction{
			ID: tid, Merchant: "BP CONNECT", MCC: "5541",
			Amount: model.NewMoney(5000, "GBP"), OccurredAt: ts,
		})
		s.Receipts = append(s.Receipts, model.Receipt{
			ID: rid, Supplier: "BP", Total: model.NewMoney(5000, "GBP"), IssuedAt: ts,
		})
		s.Labels = append(s.Labels, model.LabelledPair{TransactionID: tid, ReceiptID: rid})
		s.Expectations.GenuinelyAmbiguousTxnIDs = append(s.Expectations.GenuinelyAmbiguousTxnIDs, tid)
		s.Expectations.GenuinelyAmbiguousReceiptIDs = append(s.Expectations.GenuinelyAmbiguousReceiptIDs, rid)
		s.Expectations.TransactionOutcomes[tid] = model.OutcomeConflict
		s.Expectations.ReceiptOutcomes[rid] = model.OutcomeConflict
	}
	return s
}

func (g *Generator) generateNearDuplicates(n int) Scenario {
	base := scenarioBase().Add(300 * time.Hour)
	s := Scenario{
		Name:  "generated_near_duplicate",
		Class: ClassGeneratedNearDup,
		Expectations: model.Expectations{
			TransactionOutcomes: make(map[string]model.Outcome),
			ReceiptOutcomes:     make(map[string]model.Outcome),
		},
	}
	for i := range n {
		t := base.Add(time.Duration(i*5) * time.Minute)
		tid := fmt.Sprintf("g-nd-t-%d", i)
		rid := fmt.Sprintf("g-nd-r-%d", i)
		s.Transactions = append(s.Transactions, model.Transaction{
			ID: tid, Merchant: "SHELL FUEL", MCC: "5541",
			Amount: model.NewMoney(8000, "GBP"), OccurredAt: model.NewTimestamp(t, 0),
		})
		s.Receipts = append(s.Receipts, model.Receipt{
			ID: rid, Supplier: "Shell", Total: model.NewMoney(8000, "GBP"),
			IssuedAt: model.NewTimestamp(t, 0),
		})
		s.Labels = append(s.Labels, model.LabelledPair{TransactionID: tid, ReceiptID: rid})
		s.Expectations.TransactionOutcomes[tid] = model.OutcomeMatched
		s.Expectations.ReceiptOutcomes[rid] = model.OutcomeMatched
	}
	return s
}

func (g *Generator) generateRefunds(n int) Scenario {
	ts := model.NewTimestamp(scenarioBase().Add(400*time.Hour), 0)
	s := Scenario{
		Name:  "generated_refund",
		Class: ClassGeneratedRefund,
		Expectations: model.Expectations{
			TransactionOutcomes: make(map[string]model.Outcome),
			ReceiptOutcomes:     make(map[string]model.Outcome),
		},
	}
	for i := range n {
		tid := fmt.Sprintf("g-ref-t-%d", i)
		rid := fmt.Sprintf("g-ref-r-%d", i)
		s.Transactions = append(s.Transactions, model.Transaction{
			ID: tid, Merchant: "AMAZON UK", MCC: "5399",
			Amount: model.NewMoney(3000, "GBP"), OccurredAt: ts,
		})
		s.Receipts = append(s.Receipts, model.Receipt{
			ID: rid, Supplier: "Amazon", Total: model.NewMoney(-3000, "GBP"), IssuedAt: ts,
		})
		s.Expectations.TransactionOutcomes[tid] = model.OutcomeUnmatched
		s.Expectations.ReceiptOutcomes[rid] = model.OutcomeUnmatched
	}
	return s
}

const minutes60 = 60

// NoiseBound records generator-claimed noise limits for property verification.
type NoiseBound struct {
	TransactionID       string
	MaxAmountDiffMinor  int64
	MaxTemporalLagHours int
}

func (s Scenario) VerifyNoiseBounds() bool {
	txnByID := make(map[string]model.Transaction, len(s.Transactions))
	recByID := make(map[string]model.Receipt, len(s.Receipts))
	boundByTxn := make(map[string]NoiseBound, len(s.NoiseBounds))
	for _, t := range s.Transactions {
		txnByID[t.ID] = t
	}
	for _, r := range s.Receipts {
		recByID[r.ID] = r
	}
	for _, b := range s.NoiseBounds {
		boundByTxn[b.TransactionID] = b
	}
	for _, l := range s.Labels {
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
