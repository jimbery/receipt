package harness

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/jimbery/receipt/internal/classify"
	"github.com/jimbery/receipt/internal/emailtypes"
	"github.com/jimbery/receipt/internal/extract"
	"github.com/jimbery/receipt/internal/model"
)

// FixtureSmokeReport is the Phase 1 measurement harness output (ADR-002 D6).
// CI runs this on the frozen labelled corpus; moat and cohort-weighted metrics
// are deferred to Phase 2 when transaction ground truth exists.
type FixtureSmokeReport struct {
	SchemaVersion           string             `json:"schema_version"`
	GeneratedAt             time.Time          `json:"generated_at"`
	CorpusHash              string             `json:"corpus_hash"`
	ClassifierRulesHash     string             `json:"classifier_rules_hash"`
	ExtractorVersion        string             `json:"extractor_version"`
	ExactPrecision          float64            `json:"exact_precision"`
	ExactRecall             float64            `json:"exact_recall"`
	ReceiptBearingPrecision float64            `json:"receipt_bearing_precision"`
	ReceiptBearingRecall    float64            `json:"receipt_bearing_recall"`
	StratumExactAccuracy    map[string]float64 `json:"stratum_exact_accuracy"`
	RequiresOCRCount        int                `json:"requires_ocr_count"`
	Passed                  bool               `json:"passed"`
	Failures                []string           `json:"failures,omitempty"`
	Note                    string             `json:"note"`
}

// SmokeConfig holds classifier thresholds for the fixture smoke check only.
type SmokeConfig struct {
	ExactPrecisionMin          float64
	ExactRecallMin             float64
	ReceiptBearingPrecisionMin float64
	ReceiptBearingRecallMin    float64
}

// DefaultSmokeConfig returns classifier smoke thresholds (exact + receipt-bearing).
func DefaultSmokeConfig() SmokeConfig {
	return SmokeConfig{
		ExactPrecisionMin:          0.90,
		ExactRecallMin:             0.90,
		ReceiptBearingPrecisionMin: 0.97,
		ReceiptBearingRecallMin:    0.92,
	}
}

// RunFixtureSmoke evaluates classifier metrics on the committed corpus.
func RunFixtureSmoke(corpus []CorpusMessage, cfg SmokeConfig) FixtureSmokeReport {
	clf := classify.New()
	reg := extract.NewRegistry()
	metrics := EvaluateClassifierMetrics(corpus, clf)

	stratumTP := map[string]int{}
	stratumTotal := map[string]int{}
	for _, c := range corpus {
		stratumTotal[c.Stratum]++
		if clf.Classify(c.Msg) == c.Label {
			stratumTP[c.Stratum]++
		}
	}
	stratumAcc := make(map[string]float64, len(stratumTotal))
	for stratum, total := range stratumTotal {
		stratumAcc[stratum] = ratio(stratumTP[stratum], total)
	}

	requiresOCR := 0
	for _, c := range corpus {
		if c.Label != emailtypes.KindPurchaseReceipt {
			continue
		}
		kind := clf.Classify(c.Msg)
		ex, ok := reg.Extract(c.Msg, kind)
		if ok && ex.Grade == emailtypes.GradeRequiresOCR {
			requiresOCR++
		}
	}

	report := FixtureSmokeReport{
		SchemaVersion:           "ingest-fixture-smoke-v2",
		GeneratedAt:             time.Now().UTC(),
		CorpusHash:              hashCorpus(corpus),
		ClassifierRulesHash:     clf.RulesetHash(),
		ExtractorVersion:        reg.Version(),
		ExactPrecision:          metrics.ExactPrecision,
		ExactRecall:             metrics.ExactRecall,
		ReceiptBearingPrecision: metrics.ReceiptBearingPrecision,
		ReceiptBearingRecall:    metrics.ReceiptBearingRecall,
		StratumExactAccuracy:    stratumAcc,
		RequiresOCRCount:        requiresOCR,
		Note:                    "Fixture-lab gate on frozen corpus. Classifier exact + receipt-bearing metrics.",
	}
	report.Passed, report.Failures = checkSmokeThresholds(report, cfg)
	return report
}

func receiptBearingPrecisionFail(got, wantMin float64) string {
	return fmt.Sprintf("receipt-bearing precision %.3f < %.3f", got, wantMin)
}

func receiptBearingRecallFail(got, wantMin float64) string {
	return fmt.Sprintf("receipt-bearing recall %.3f < %.3f", got, wantMin)
}

func checkSmokeThresholds(r FixtureSmokeReport, cfg SmokeConfig) (bool, []string) {
	var fails []string
	if r.ExactPrecision < cfg.ExactPrecisionMin {
		fails = append(fails, fmt.Sprintf("exact precision %.3f < %.3f", r.ExactPrecision, cfg.ExactPrecisionMin))
	}
	if r.ExactRecall < cfg.ExactRecallMin {
		fails = append(fails, fmt.Sprintf("exact recall %.3f < %.3f", r.ExactRecall, cfg.ExactRecallMin))
	}
	if r.ReceiptBearingPrecision < cfg.ReceiptBearingPrecisionMin {
		fails = append(fails, receiptBearingPrecisionFail(r.ReceiptBearingPrecision, cfg.ReceiptBearingPrecisionMin))
	}
	if r.ReceiptBearingRecall < cfg.ReceiptBearingRecallMin {
		fails = append(fails, receiptBearingRecallFail(r.ReceiptBearingRecall, cfg.ReceiptBearingRecallMin))
	}
	return len(fails) == 0, fails
}

// WriteSmokeReport persists fixture smoke output.
func WriteSmokeReport(path string, report FixtureSmokeReport) error {
	b, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if mkErr := os.MkdirAll(filepath.Dir(path), 0o750); mkErr != nil {
		return mkErr
	}
	return os.WriteFile(path, b, 0o600)
}

// EvaluateExtractorFixtures scores field accuracy on labelled merchant fixtures.
func EvaluateExtractorFixtures(labels []FieldLabel, receipts []model.Receipt) float64 {
	return EvaluateFields(labels, receipts)
}
