// Command ingest-evaluate runs the Phase 1 fixture-lab gate (ADR-002 D6).
package main

import (
	"flag"
	"fmt"
	"os"

	iharness "github.com/jimbery/receipt/internal/ingest/harness"
)

func main() {
	jsonOut := flag.String("json", "", "write fixture smoke report JSON to path")
	flag.Parse()

	corpus := iharness.LoadFrozenCorpus()
	report := iharness.RunFixtureSmoke(corpus, iharness.DefaultSmokeConfig())

	fmt.Fprintf(os.Stdout, "Phase 1 fixture-lab gate\n")
	fmt.Fprintf(os.Stdout, "  corpus size:                %d\n", len(corpus))
	fmt.Fprintf(os.Stdout, "  exact precision/recall:     %.1f%% / %.1f%%\n",
		report.ExactPrecision*100, report.ExactRecall*100)
	fmt.Fprintf(os.Stdout, "  receipt-bearing prec/rec:   %.1f%% / %.1f%%\n",
		report.ReceiptBearingPrecision*100, report.ReceiptBearingRecall*100)
	fmt.Fprintf(os.Stdout, "  requires_ocr_count:         %d\n", report.RequiresOCRCount)
	fmt.Fprintf(os.Stdout, "  passed:                     %v\n", report.Passed)
	for _, f := range report.Failures {
		fmt.Fprintf(os.Stdout, "  FAIL: %s\n", f)
	}

	if *jsonOut != "" {
		if err := iharness.WriteSmokeReport(*jsonOut, report); err != nil {
			fmt.Fprintf(os.Stderr, "write: %v\n", err)
			os.Exit(1)
		}
	}

	if !report.Passed {
		os.Exit(1)
	}
}
