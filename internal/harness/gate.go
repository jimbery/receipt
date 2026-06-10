package harness

import (
	"encoding/json"
	"slices"
	"time"

	"github.com/jimbery/receipt/internal/match"
	"github.com/jimbery/receipt/internal/model"
	"github.com/jimbery/receipt/internal/synth"
)

const gateReportSchemaVersion = "1.0"

// GateThresholds are commitments made before the first evaluation run (M0.5).
type GateThresholds struct {
	MaxFMR                 float64  `json:"max_fmr"`
	MinRecall              float64  `json:"min_recall"`
	MinHardClassRecall     float64  `json:"min_hard_class_recall"`
	MinConflictCorrectness float64  `json:"min_conflict_correctness"`
	MaxConflictRate        float64  `json:"max_conflict_rate"`
	RequireDeterminism     bool     `json:"require_determinism"`
	HardScenarioClasses    []string `json:"hard_scenario_classes"`
}

const (
	gateMaxFMR             = 0.005
	gateMinRecall          = 0.85
	gateMinHardClassRecall = 0.70
	gateMinConflictCorrect = 0.90
	gateMaxConflictRate    = 0.05
)

// DefaultGateThresholds returns Phase 0 gate commitments.
func DefaultGateThresholds() GateThresholds {
	return GateThresholds{
		MaxFMR:                 gateMaxFMR,
		MinRecall:              gateMinRecall,
		MinHardClassRecall:     gateMinHardClassRecall,
		MinConflictCorrectness: gateMinConflictCorrect,
		MaxConflictRate:        gateMaxConflictRate,
		RequireDeterminism:     true,
		HardScenarioClasses: []string{
			string(synth.ClassAmbiguous),
			string(synth.ClassNearDuplicate),
			string(synth.ClassMerchantMangling),
			string(synth.ClassTimezoneShift),
			string(synth.ClassSettlementDelay),
		},
	}
}

// ClassMetrics holds per-scenario-class evaluation results.
type ClassMetrics struct {
	Class               string  `json:"class"`
	Precision           float64 `json:"precision"`
	Recall              float64 `json:"recall"`
	FalseMatchRate      float64 `json:"false_match_rate"`
	ConflictRate        float64 `json:"conflict_rate"`
	ConflictCorrectness float64 `json:"conflict_correctness"`
	Matched             int     `json:"matched"`
	Conflicts           int     `json:"conflicts"`
	OutcomeViolations   int     `json:"outcome_violations"`
	Pass                bool    `json:"pass"`
}

// GateReport is the machine-readable output of a formal gate run.
type GateReport struct {
	SchemaVersion  string         `json:"schema_version"`
	Timestamp      time.Time      `json:"timestamp"`
	Seed           int64          `json:"seed"`
	ConfigHash     string         `json:"config_hash"`
	DatasetScale   int            `json:"dataset_scale"`
	EngineVersion  string         `json:"engine_version"`
	Thresholds     GateThresholds `json:"thresholds"`
	Overall        ClassMetrics   `json:"overall"`
	PerClass       []ClassMetrics `json:"per_class"`
	Deterministic  bool           `json:"deterministic"`
	Passed         bool           `json:"passed"`
	FailureReasons []string       `json:"failure_reasons,omitempty"`
}

// GateRunOptions carries metadata for a formal gate evaluation.
type GateRunOptions struct {
	Seed          int64
	ConfigHash    string
	DatasetScale  int
	EngineVersion string
}

// EvaluateScenario runs the matcher on one scenario and returns class metrics.
func EvaluateScenario(engine *match.Engine, s synth.Scenario, t GateThresholds) ClassMetrics {
	result := engine.Match(s.Transactions, s.Receipts)
	metrics := EvaluateResult(s.Transactions, s.Receipts, result, s.Labels)
	cc := ConflictCorrectness(result, s.Expectations)
	violations := len(model.ValidateOutcomes(result, s.Expectations))

	pass := violations == 0 && metrics.FalseMatchRate <= t.MaxFMR
	if expectsHighConflicts(s.Class) {
		pass = pass && metrics.FalseMatchRate == 0 && cc >= t.MinConflictCorrectness
	} else {
		pass = pass && metrics.ConflictRate <= t.MaxConflictRate
	}
	if isHardClass(string(s.Class), t) && s.Class != synth.ClassAmbiguous {
		pass = pass && metrics.Recall >= t.MinHardClassRecall
	}

	return ClassMetrics{
		Class:               string(s.Class),
		Precision:           metrics.Precision,
		Recall:              metrics.Recall,
		FalseMatchRate:      metrics.FalseMatchRate,
		ConflictRate:        metrics.ConflictRate,
		ConflictCorrectness: cc,
		Matched:             metrics.Matched,
		Conflicts:           metrics.Conflicts,
		OutcomeViolations:   violations,
		Pass:                pass,
	}
}

func expectsHighConflicts(class synth.Class) bool {
	return class == synth.ClassAmbiguous ||
		class == synth.ClassDuplicateReceipt ||
		class == synth.ClassDensityStress ||
		class == synth.ClassGeneratedAmbiguous
}

func isHardClass(class string, t GateThresholds) bool {
	return slices.Contains(t.HardScenarioClasses, class)
}

// ConflictCorrectness measures how many flagged conflicts are genuinely ambiguous.
func ConflictCorrectness(result model.MatchResult, exp model.Expectations) float64 {
	if len(result.Conflicts) == 0 {
		return 1
	}
	ambTxn := exp.AmbiguousTxnSet()
	ambReceipt := exp.AmbiguousReceiptSet()
	correct := 0
	for _, c := range result.Conflicts {
		if c.TransactionID != "" {
			if _, ok := ambTxn[c.TransactionID]; ok {
				correct++
				continue
			}
		}
		if c.ReceiptID != "" {
			if _, ok := ambReceipt[c.ReceiptID]; ok {
				correct++
			}
		}
	}
	return float64(correct) / float64(len(result.Conflicts))
}

// RunGate evaluates all scenarios and checks gate thresholds.
func RunGate(engine *match.Engine, scenarios []synth.Scenario, thresholds GateThresholds) GateReport {
	return RunGateWithOptions(engine, scenarios, thresholds, GateRunOptions{})
}

// RunGateWithOptions evaluates scenarios with report metadata (M0.5).
func RunGateWithOptions(
	engine *match.Engine,
	scenarios []synth.Scenario,
	thresholds GateThresholds,
	opts GateRunOptions,
) GateReport {
	report := GateReport{
		SchemaVersion: gateReportSchemaVersion,
		Timestamp:     time.Now().UTC(),
		Seed:          opts.Seed,
		ConfigHash:    opts.ConfigHash,
		DatasetScale:  opts.DatasetScale,
		EngineVersion: opts.EngineVersion,
		Thresholds:    thresholds,
		PerClass:      make([]ClassMetrics, 0, len(scenarios)),
	}

	var (
		totalCorrect, totalLabelled, totalMatches, totalFalse, totalConflicts, correctConflicts int
		totalTxn, totalConflictsEmitted                                                         int
	)

	for _, s := range scenarios {
		result := engine.Match(s.Transactions, s.Receipts)
		cm := EvaluateScenario(engine, s, thresholds)
		report.PerClass = append(report.PerClass, cm)

		m := EvaluateResult(s.Transactions, s.Receipts, result, s.Labels)
		totalMatches += m.Matched
		totalFalse += m.FalseMatches
		totalConflicts += m.Conflicts
		totalTxn += len(s.Transactions)
		totalConflictsEmitted += len(result.Conflicts)
		totalCorrect += countCorrect(result.MatchedOnly(), s.Labels)
		totalLabelled += len(s.Labels)
		correctConflicts += int(ConflictCorrectness(result, s.Expectations) * float64(m.Conflicts))
	}

	if totalMatches > 0 {
		report.Overall.FalseMatchRate = float64(totalFalse) / float64(totalMatches)
		report.Overall.Precision = 1 - report.Overall.FalseMatchRate
	}
	if totalLabelled > 0 {
		report.Overall.Recall = float64(totalCorrect) / float64(totalLabelled)
	}
	if totalTxn > 0 {
		report.Overall.ConflictRate = float64(totalConflictsEmitted) / float64(totalTxn)
	}
	if totalConflicts > 0 {
		report.Overall.ConflictCorrectness = float64(correctConflicts) / float64(totalConflicts)
	} else {
		report.Overall.ConflictCorrectness = 1
	}
	report.Overall.Matched = totalMatches
	report.Overall.Conflicts = totalConflicts

	report.Deterministic = CheckDeterminism(engine, scenarios)
	report.Passed, report.FailureReasons = checkThresholds(report, thresholds)
	return report
}

func countCorrect(matches []model.Match, labels []LabelledPair) int {
	labelMap := make(map[string]string, len(labels))
	for _, l := range labels {
		labelMap[l.TransactionID] = l.ReceiptID
	}
	correct := 0
	for _, m := range matches {
		if labelMap[m.TransactionID] == m.ReceiptID {
			correct++
		}
	}
	return correct
}

func checkThresholds(report GateReport, t GateThresholds) (bool, []string) {
	var reasons []string
	if report.Overall.FalseMatchRate > t.MaxFMR {
		reasons = append(reasons, "FMR exceeds threshold")
	}
	if report.Overall.Recall < t.MinRecall {
		reasons = append(reasons, "overall recall below threshold")
	}
	if report.Overall.ConflictCorrectness < t.MinConflictCorrectness {
		reasons = append(reasons, "conflict correctness below threshold")
	}
	if report.Overall.ConflictRate > t.MaxConflictRate {
		reasons = append(reasons, "conflict rate exceeds threshold")
	}
	if t.RequireDeterminism && !report.Deterministic {
		reasons = append(reasons, "determinism check failed")
	}

	hardSet := make(map[string]struct{}, len(t.HardScenarioClasses))
	for _, c := range t.HardScenarioClasses {
		hardSet[c] = struct{}{}
	}
	for _, cm := range report.PerClass {
		if _, hard := hardSet[cm.Class]; !hard || cm.Class == string(synth.ClassAmbiguous) {
			continue
		}
		if cm.Recall < t.MinHardClassRecall {
			reasons = append(reasons, "hard class "+cm.Class+" recall below threshold")
		}
	}

	return len(reasons) == 0, reasons
}

// CheckDeterminism verifies identical output across repeated runs.
func CheckDeterminism(engine *match.Engine, scenarios []synth.Scenario) bool {
	for _, s := range scenarios {
		first := engine.Match(s.Transactions, s.Receipts)
		for range 5 {
			if !resultsEqual(first, engine.Match(s.Transactions, s.Receipts)) {
				return false
			}
		}
	}
	return true
}

// ResultsEqual compares match results including conflict contents and order.
func ResultsEqual(a, b model.MatchResult) bool {
	return resultsEqual(a, b)
}

func resultsEqual(a, b model.MatchResult) bool {
	if len(a.Matches) != len(b.Matches) || len(a.Conflicts) != len(b.Conflicts) {
		return false
	}
	for i, m := range a.Matches {
		if m.TransactionID != b.Matches[i].TransactionID ||
			m.ReceiptID != b.Matches[i].ReceiptID ||
			m.Confidence != b.Matches[i].Confidence {
			return false
		}
	}
	for i, c := range a.Conflicts {
		bc := b.Conflicts[i]
		if c.TransactionID != bc.TransactionID ||
			c.ReceiptID != bc.ReceiptID ||
			c.TopConfidence != bc.TopConfidence ||
			c.Reason != bc.Reason ||
			!stringSlicesEqual(c.CompetingIDs, bc.CompetingIDs) {
			return false
		}
	}
	return stringSlicesEqual(a.UnmatchedTxnIDs, b.UnmatchedTxnIDs) &&
		stringSlicesEqual(a.UnmatchedReceiptIDs, b.UnmatchedReceiptIDs)
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// CompareConfigs runs baseline vs candidate on the same scenarios.
func CompareConfigs(baseline, candidate match.Config, scenarios []synth.Scenario) (GateReport, GateReport) {
	t := DefaultGateThresholds()
	return RunGate(match.NewEngine(baseline), scenarios, t), RunGate(match.NewEngine(candidate), scenarios, t)
}

// MarshalReport serialises a gate report as JSON.
func MarshalReport(r GateReport) ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}
