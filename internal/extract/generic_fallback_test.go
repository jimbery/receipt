package extract_test

import (
	"testing"
	"time"

	"github.com/jimbery/receipt/internal/emailtypes"
	"github.com/jimbery/receipt/internal/extract"
)

func TestGenericFallback_SchemaOrgBeforeHTML(t *testing.T) {
	reg := extract.NewRegistry()
	msg := emailtypes.RawMessage{
		ID:   "fb-1",
		From: "shop@longtail-trader.example",
		Date: time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC),
		BodyHTML: `<html><body>
<script type="application/ld+json">{"@type":"Order","price":"99.99"}</script>
<table><tr><td>Order total</td><td>£1.00</td></tr></table>
</body></html>`,
	}
	ex, ok := reg.Extract(msg, emailtypes.KindPurchaseReceipt)
	if !ok {
		t.Fatal("expected extraction")
	}
	if ex.TotalMinor != 9999 {
		t.Fatalf("total %d want schema.org 9999", ex.TotalMinor)
	}
	if ex.FieldProvenance["total"] != emailtypes.ProvSchemaOrg {
		t.Fatalf("provenance %v want schema_org", ex.FieldProvenance["total"])
	}
}

func TestGenericFallback_PDFTextLayer(t *testing.T) {
	reg := extract.NewRegistry()
	pdf := []byte("%PDF-1.1\nBT (Order total £45.50) Tj ET")
	msg := emailtypes.RawMessage{
		ID:   "fb-2",
		From: "accounts@longtail-trader.example",
		Date: time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC),
		Attachments: []emailtypes.Attachment{{
			Filename: "invoice.pdf", ContentType: "application/pdf", Content: pdf,
		}},
	}
	ex, ok := reg.Extract(msg, emailtypes.KindPurchaseReceipt)
	if !ok {
		t.Fatal("expected extraction")
	}
	if ex.TotalMinor != 4550 {
		t.Fatalf("total %d want 4550 from PDF text layer", ex.TotalMinor)
	}
	if ex.FieldProvenance["total"] != emailtypes.ProvPDFText {
		t.Fatalf("provenance %v want pdf_text", ex.FieldProvenance["total"])
	}
}

func TestRegistry_VersionAndGrade(t *testing.T) {
	reg := extract.NewRegistry()
	if reg.Version() == "" {
		t.Fatal("registry missing version string")
	}
	msg := emailtypes.RawMessage{
		ID: "g-1", From: "receipts@screwfix.com",
		Date:     time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC),
		BodyHTML: `<table><tr><td>Total</td><td>£12.00</td></tr></table>`,
	}
	ex, ok := reg.Extract(msg, emailtypes.KindPurchaseReceipt)
	if !ok || ex.Grade == "" {
		t.Fatalf("expected graded extraction, got ok=%v grade=%s", ok, ex.Grade)
	}
}
