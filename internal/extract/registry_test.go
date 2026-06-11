package extract_test

import (
	"testing"
	"time"

	"github.com/jimbery/receipt/internal/emailtypes"
	"github.com/jimbery/receipt/internal/extract"
)

type novelExtractor struct{}

func (novelExtractor) Name() string { return "novel" }

func (novelExtractor) Matches(msg emailtypes.RawMessage) bool {
	return msg.From == "novel@pilot.example"
}

func (novelExtractor) Extract(_ emailtypes.RawMessage) (emailtypes.ExtractedReceipt, error) {
	return emailtypes.ExtractedReceipt{
		Supplier: "NovelCo", TotalMinor: 4242, Currency: "GBP",
	}, nil
}

// M1.4 pluggability: register a novel extractor without changing pipeline wiring.
func TestRegistry_PluggableWithoutPipelineChange(t *testing.T) {
	reg := extract.NewRegistryWith([]extract.Extractor{novelExtractor{}})
	msg := emailtypes.RawMessage{
		ID: "plug-1", From: "novel@pilot.example", Date: time.Now(),
		BodyText: "ignored",
	}
	ex, ok := reg.Extract(msg, emailtypes.KindPurchaseReceipt)
	if !ok {
		t.Fatal("expected extraction")
	}
	if ex.Supplier != "NovelCo" || ex.TotalMinor != 4242 {
		t.Fatalf("novel extractor not routed: %+v", ex)
	}
}
