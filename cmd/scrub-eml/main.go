// Command scrub-eml converts raw volunteer .eml files into scrubbed CI fixtures.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jimbery/receipt/internal/scrub"
)

func main() {
	inDir := flag.String("in", "emls", "directory of raw .eml files (gitignored)")
	outDir := flag.String("out", "test/testdata/email/merchants", "destination for scrubbed fixtures")
	dryRun := flag.Bool("dry-run", false, "scrub and audit only; do not write")
	flag.Parse()

	entries, err := os.ReadDir(*inDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read input: %v\n", err)
		os.Exit(1)
	}

	s := scrub.New(nil)
	usedBasenames := map[string]int{}
	var written int
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".eml") {
			continue
		}
		inPath := filepath.Join(*inDir, e.Name())
		raw, readErr := os.ReadFile(inPath)
		if readErr != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", inPath, readErr)
			os.Exit(1)
		}
		merchant, merchErr := scrub.MerchantFromEML(string(raw))
		if merchErr != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", inPath, merchErr)
			os.Exit(1)
		}
		out, scrubErr := scrub.ExportEML(raw, s)
		if scrubErr != nil {
			fmt.Fprintf(os.Stderr, "%s: scrub failed: %v\n", inPath, scrubErr)
			os.Exit(1)
		}
		basename := uniqueBasename(usedBasenames, merchant, scrub.SlugifyFilename(e.Name()))
		if *dryRun {
			fmt.Fprintf(os.Stdout, "ok  %s -> %s/%s.eml\n", e.Name(), merchant, basename)
			written++
			continue
		}
		dest, writeErr := scrub.WriteFixture(*outDir, merchant, basename, out)
		if writeErr != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", inPath, writeErr)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stdout, "wrote %s\n", dest)
		written++
	}
	if written == 0 {
		fmt.Fprintf(os.Stderr, "no .eml files found in %s\n", *inDir)
		os.Exit(1)
	}
}

func uniqueBasename(used map[string]int, merchant, base string) string {
	key := merchant + "/" + base
	n := used[key]
	used[key] = n + 1
	if n == 0 {
		return base
	}
	return fmt.Sprintf("%s-%d", base, n+1)
}
