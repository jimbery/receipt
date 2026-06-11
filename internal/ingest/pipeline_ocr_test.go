package ingest_test

import (
	"context"
	"testing"
	"time"

	"github.com/jimbery/receipt/internal/classify"
	"github.com/jimbery/receipt/internal/emailtypes"
	"github.com/jimbery/receipt/internal/extract"
	"github.com/jimbery/receipt/internal/ingest"
)

func TestPipeline_RequiresOCRCounter(t *testing.T) {
	day := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	src := &staticSource{msgs: []emailtypes.RawMessage{
		{
			ID: "ocr-1", From: "shop@example.com", Date: day,
			Subject: "Invoice",
			Attachments: []emailtypes.Attachment{{
				Filename: "scan.pdf", ContentType: "application/pdf",
				Content: []byte("%PDF-1.1 /Encrypt image-only"),
			}},
		},
	}}
	p := &ingest.Pipeline{
		Source:     src,
		Classifier: classify.New(),
		Registry:   extract.NewRegistry(),
	}
	result, err := p.ProcessMailbox(context.Background(), time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if result.RequiresOCRCount != 1 {
		t.Fatalf("requires_ocr count %d want 1", result.RequiresOCRCount)
	}
	if len(result.Receipts) != 0 {
		t.Fatalf("expected no canonical receipts for OCR-only message, got %d", len(result.Receipts))
	}
}
