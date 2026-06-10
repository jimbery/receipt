// Command evaluate runs the Phase 0 synthetic harness and formal gate (M0.5).
package main

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/jimbery/receipt/internal/harness"
	"github.com/jimbery/receipt/internal/match"
	"github.com/jimbery/receipt/internal/synth"
)

const engineVersion = "phase-0.2"

func main() {
	var (
		seed      = flag.Int64("seed", 42, "RNG seed for generated dataset")
		generated = flag.Int("generated", 100, "number of generated pairs (0 to skip)")
		jsonOut   = flag.String("json", "", "write machine-readable gate report to path")
		compare   = flag.Bool("compare", false, "run baseline vs tightened MinConfidence comparison")
	)
	flag.Parse()

	cfg := match.DefaultConfig()
	engine := match.NewEngine(cfg)
	scenarios := synth.AllExtended()

	if *generated > 0 {
		gen := synth.NewGenerator(*seed)
		scenarios = append(scenarios, gen.GenerateSuite(*generated)...)
	}

	thresholds := harness.DefaultGateThresholds()
	report := harness.RunGateWithOptions(engine, scenarios, thresholds, harness.GateRunOptions{
		Seed:          *seed,
		ConfigHash:    cfg.Hash(),
		DatasetScale:  *generated,
		EngineVersion: engineVersion,
	})

	if *compare {
		tight := cfg
		tight.MinConfidence = 0.85
		base, cand := harness.CompareConfigs(cfg, tight, scenarios)
		printCompare(os.Stdout, base, cand)
	}

	printSummary(os.Stdout, report)

	if *jsonOut != "" {
		data, err := harness.MarshalReport(report)
		if err != nil {
			fmt.Fprintf(os.Stderr, "marshal: %v\n", err)
			os.Exit(1)
		}
		if writeErr := os.WriteFile(*jsonOut, data, 0o600); writeErr != nil {
			fmt.Fprintf(os.Stderr, "write json: %v\n", writeErr)
			os.Exit(1)
		}
	}

	if !report.Passed {
		for _, r := range report.FailureReasons {
			fmt.Fprintf(os.Stderr, "gate failed: %s\n", r)
		}
		os.Exit(1)
	}
}

func printSummary(out *os.File, report harness.GateReport) {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	mustWrite(func() error {
		_, err := fmt.Fprintln(w, "CLASS\tRECALL\tFMR\tCONFLICT_RATE\tCONFLICT_OK\tOUTCOME_OK\tPASS")
		return err
	})
	defer func() {
		if err := w.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "flush: %v\n", err)
			os.Exit(1)
		}
	}()

	for _, cm := range report.PerClass {
		outcomeOK := cm.OutcomeViolations == 0
		mustWrite(func() error {
			_, err := fmt.Fprintf(w, "%s\t%.3f\t%.3f\t%.3f\t%.3f\t%v\t%v\n",
				cm.Class, cm.Recall, cm.FalseMatchRate, cm.ConflictRate,
				cm.ConflictCorrectness, outcomeOK, cm.Pass)
			return err
		})
	}

	mustWrite(func() error {
		_, err := fmt.Fprintf(out,
			"\nOVERALL recall=%.3f fmr=%.3f conflict_rate=%.3f conflict_ok=%.3f deterministic=%v passed=%v\n"+
				"seed=%d config_hash=%s scale=%d engine=%s schema=%s\n",
			report.Overall.Recall, report.Overall.FalseMatchRate, report.Overall.ConflictRate,
			report.Overall.ConflictCorrectness, report.Deterministic, report.Passed,
			report.Seed, report.ConfigHash, report.DatasetScale, report.EngineVersion, report.SchemaVersion)
		return err
	})
}

func printCompare(out *os.File, base, cand harness.GateReport) {
	mustWrite(func() error {
		_, err := fmt.Fprintf(
			out,
			"\nCOMPARE baseline_fmr=%.4f candidate_fmr=%.4f baseline_recall=%.3f candidate_recall=%.3f\n",
			base.Overall.FalseMatchRate,
			cand.Overall.FalseMatchRate,
			base.Overall.Recall,
			cand.Overall.Recall,
		)
		return err
	})
}

func mustWrite(fn func() error) {
	if err := fn(); err != nil {
		fmt.Fprintf(os.Stderr, "write: %v\n", err)
		os.Exit(1)
	}
}
