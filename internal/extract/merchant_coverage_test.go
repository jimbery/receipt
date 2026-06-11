package extract_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jimbery/receipt/internal/classify"
	"github.com/jimbery/receipt/internal/emailtypes"
	"github.com/jimbery/receipt/internal/extract"
	"github.com/jimbery/receipt/internal/mail"
)

type merchantCoverageSpec struct {
	merchant       string
	partialFixture string
	familyFixtures []string
}

func merchantCoverageSpecs() []merchantCoverageSpec {
	return []merchantCoverageSpec{
		{
			merchant:       "amazon",
			partialFixture: "receipt.txt",
			familyFixtures: []string{"receipt.txt", "partial-shipment.txt"},
		},
		{merchant: "screwfix", partialFixture: "partial.txt", familyFixtures: []string{"receipt.txt", "refund.txt"}},
		{
			merchant:       "toolstation",
			partialFixture: "partial-shipment.txt",
			familyFixtures: []string{"receipt.txt", "partial-shipment.txt"},
		},
		{
			merchant:       "bandq",
			partialFixture: "vat-exempt.txt",
			familyFixtures: []string{"receipt.txt", "confirmation-4001.txt"},
		},
		{merchant: "bp", partialFixture: "envelope.txt", familyFixtures: []string{"receipt.txt", "receipt.txt"}},
		{
			merchant:       "shell",
			partialFixture: "envelope.txt",
			familyFixtures: []string{"app-receipt.txt", "app-receipt.txt"},
		},
		{
			merchant:       "esso",
			partialFixture: "envelope.txt",
			familyFixtures: []string{"app-receipt.txt", "app-receipt.txt"},
		},
		{
			merchant:       "edf",
			partialFixture: "standing-charge.txt",
			familyFixtures: []string{"invoice.txt", "invoice.txt"},
		},
		{merchant: "octopus", partialFixture: "invoice.txt", familyFixtures: []string{"invoice.txt", "invoice.txt"}},
	}
}

func TestMerchantCoverage_PartialAndFamilyDedup(t *testing.T) {
	root := filepath.Join("..", "..", "test", "testdata", "email", "merchants")
	clf := classify.New()
	reg := extract.NewRegistry()

	for _, spec := range merchantCoverageSpecs() {
		t.Run(spec.merchant, func(t *testing.T) {
			partial, err := extractFixture(root, spec.merchant, spec.partialFixture, clf, reg)
			if err != nil {
				t.Fatal(err)
			}
			grade := extract.Grade(partial)
			if grade != emailtypes.GradePartial && grade != emailtypes.GradeEnvelope {
				t.Fatalf("partial fixture grade %s want partial or envelope", grade)
			}

			if len(spec.familyFixtures) < 2 {
				t.Fatal("need family fixture pair")
			}
			first, err := extractFixture(root, spec.merchant, spec.familyFixtures[0], clf, reg)
			if err != nil {
				t.Fatal(err)
			}
			second, err := extractFixture(root, spec.merchant, spec.familyFixtures[1], clf, reg)
			if err != nil {
				t.Fatal(err)
			}
			if first.FamilyKey == "" || first.FamilyKey != second.FamilyKey {
				t.Fatalf("family keys %q vs %q want equal", first.FamilyKey, second.FamilyKey)
			}
		})
	}
}

func extractFixture(
	root, merchant, file string,
	clf *classify.Classifier,
	reg *extract.Registry,
) (emailtypes.ExtractedReceipt, error) {
	data, err := os.ReadFile(filepath.Join(root, merchant, file))
	if err != nil {
		return emailtypes.ExtractedReceipt{}, err
	}
	msg, err := mail.ParseRFC822(data, file, "coverage")
	if err != nil {
		return emailtypes.ExtractedReceipt{}, err
	}
	kind := clf.Classify(msg)
	ex, ok := reg.Extract(msg, kind)
	if !ok {
		return emailtypes.ExtractedReceipt{}, os.ErrInvalid
	}
	return ex, nil
}
