package extract

import (
	"strings"

	"github.com/jimbery/receipt/internal/emailtypes"
	"github.com/jimbery/receipt/internal/model"
)

// GenericExtractor handles schema.org JSON-LD, HTML tables, and text-layer PDF.
type GenericExtractor struct{}

func (g *GenericExtractor) Name() string { return "generic" }

func (g *GenericExtractor) Matches(_ emailtypes.RawMessage) bool { return true }

func (g *GenericExtractor) Extract(msg emailtypes.RawMessage) (emailtypes.ExtractedReceipt, error) {
	r := emailtypes.ExtractedReceipt{
		Supplier:        SupplierFromFrom(msg.From),
		IssuedAt:        msg.Date,
		Currency:        "GBP",
		FieldProvenance: map[string]emailtypes.Provenance{},
	}
	extractHTMLFields(&r, msg.BodyHTML)
	if r.TotalMinor == 0 && msg.BodyText != "" {
		if textTotal, textOK := parseHTMLTotal(msg.BodyText); textOK {
			r.TotalMinor = textTotal
			r.FieldProvenance["total"] = emailtypes.ProvHTML
		}
	}
	extractPDFFields(&r, msg.Attachments)
	r.FamilyKey = FamilyKey(r.Supplier, r.OrderRef, r.IssuedAt, r.TotalMinor, r.Currency)
	return r, nil
}

func extractHTMLFields(r *emailtypes.ExtractedReceipt, html string) {
	if html == "" {
		return
	}
	if schemaTotal, schemaOK := ParseJSONLDTotal(html); schemaOK {
		r.TotalMinor = schemaTotal
		r.FieldProvenance["total"] = emailtypes.ProvSchemaOrg
	} else if tableTotal, tableOK := ParseHTMLTableTotal(html); tableOK {
		r.TotalMinor = tableTotal
		r.FieldProvenance["total"] = emailtypes.ProvHTML
	}
	if ref, refOK := parseOrderRef(ExtractVisibleText(html)); refOK {
		r.OrderRef = ref
		r.FieldProvenance["order_ref"] = emailtypes.ProvHTML
	}
	if items := parseLineItems(ExtractVisibleText(html)); len(items) > 0 {
		r.LineItems = items
		r.FieldProvenance["line_items"] = emailtypes.ProvHTML
	}
}

func extractPDFFields(r *emailtypes.ExtractedReceipt, attachments []emailtypes.Attachment) {
	for _, a := range attachments {
		if !strings.Contains(strings.ToLower(a.ContentType), "pdf") {
			continue
		}
		text := ExtractPDFText(a.Content)
		if text == "" {
			r.Grade = emailtypes.GradeRequiresOCR
			r.AttachmentRefs = append(r.AttachmentRefs, a.Filename)
			continue
		}
		if r.TotalMinor == 0 {
			if pdfTotal, pdfOK := ParsePDFTotal(text); pdfOK {
				r.TotalMinor = pdfTotal
				r.FieldProvenance["total"] = emailtypes.ProvPDFText
			}
		}
		if len(r.LineItems) == 0 {
			if items := parseLineItems(text); len(items) > 0 {
				r.LineItems = items
				r.FieldProvenance["line_items"] = emailtypes.ProvPDFText
			}
		}
	}
}

func parseLineItems(body string) []model.LineItem {
	matches := lineItemRe.FindAllStringSubmatch(body, -1)
	if len(matches) == 0 {
		return nil
	}
	items := make([]model.LineItem, 0, len(matches))
	for _, m := range matches {
		minor, ok := ParsePoundsToMinor(m[2])
		if !ok {
			continue
		}
		desc := strings.TrimSpace(m[1])
		if strings.EqualFold(desc, "vat") || strings.HasPrefix(strings.ToLower(desc), "vat ") {
			continue
		}
		items = append(items, model.LineItem{
			Description: desc,
			NetAmount:   model.NewMoney(minor, "GBP"),
		})
	}
	return items
}
