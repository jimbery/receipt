package extract_test

import (
	"testing"
	"time"

	"github.com/jimbery/receipt/internal/emailtypes"
	"github.com/jimbery/receipt/internal/extract"
)

func TestCountRequiresOCR(t *testing.T) {
	day := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	reg := extract.NewRegistry()
	imagePDF := emailtypes.RawMessage{
		ID: "ocr-1", From: "shop@example.com", Date: day,
		Attachments: []emailtypes.Attachment{{
			Filename: "scan.pdf", ContentType: "application/pdf",
			Content: []byte("%PDF-1.1 image only binary \x00\x01"),
		}},
	}
	ex, ok := reg.Extract(imagePDF, emailtypes.KindPurchaseReceipt)
	if !ok {
		t.Fatal("expected extraction attempt")
	}
	if ex.Grade != emailtypes.GradeRequiresOCR {
		t.Fatalf("grade %s want requires_ocr", ex.Grade)
	}
	if extract.CountRequiresOCR([]emailtypes.ExtractedReceipt{ex}) != 1 {
		t.Fatal("expected requires_ocr count 1")
	}
}
