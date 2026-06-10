// Command evaluate runs the Phase 0 synthetic harness (ADR D7).
package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/jimbery/receipt/internal/harness"
	"github.com/jimbery/receipt/internal/match"
	"github.com/jimbery/receipt/internal/synth"
)

func main() {
	engine := match.NewEngine(match.DefaultConfig())
	scenarios := synth.All()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	mustWrite(func() error {
		_, err := fmt.Fprintln(w, "CLASS\tPRECISION\tRECALL\tFMR\tCONFLICTS\tMATCHED\tPASS")
		return err
	})
	defer func() {
		if err := w.Flush(); err != nil {
			fmt.Fprintf(os.Stderr, "flush: %v\n", err)
			os.Exit(1)
		}
	}()

	var (
		totalPrecision float64
		totalFMR       float64
		count          int
	)

	for _, s := range scenarios {
		result := engine.Match(s.Transactions, s.Receipts)
		metrics := harness.EvaluateResult(s.Transactions, s.Receipts, result, s.Labels)

		pass := "yes"
		if s.Class == synth.ClassAmbiguous {
			if metrics.FalseMatchRate > 0 {
				pass = "no"
			}
		} else if metrics.Precision < 1.0 || metrics.FalseMatchRate > 0 {
			pass = "no"
		}

		mustWrite(func() error {
			_, err := fmt.Fprintf(w, "%s\t%.3f\t%.3f\t%.3f\t%d\t%d\t%s\n",
				s.Class, metrics.Precision, metrics.Recall, metrics.FalseMatchRate,
				metrics.Conflicts, metrics.Matched, pass)
			return err
		})

		if s.Class != synth.ClassAmbiguous {
			totalPrecision += metrics.Precision
			totalFMR += metrics.FalseMatchRate
			count++
		}
	}

	if count > 0 {
		mustWrite(func() error {
			_, err := fmt.Fprintf(os.Stdout, "\nAggregate (excl. ambiguous): precision=%.3f fmr=%.3f\n",
				totalPrecision/float64(count), totalFMR/float64(count))
			return err
		})
	}
}

func mustWrite(fn func() error) {
	if err := fn(); err != nil {
		fmt.Fprintf(os.Stderr, "write: %v\n", err)
		os.Exit(1)
	}
}
