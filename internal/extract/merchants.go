package extract

import (
	"regexp"
	"strings"

	"github.com/jimbery/receipt/internal/emailtypes"
	"github.com/jimbery/receipt/internal/model"
)

type domainExtractor struct {
	domain string
}

func (d *domainExtractor) matchDomain(from string) bool {
	return strings.Contains(strings.ToLower(from), d.domain)
}

// AmazonExtractor handles Amazon order, shipment, and invoice emails.
type AmazonExtractor struct{ domainExtractor }

func (a *AmazonExtractor) Name() string { return "amazon" }

func (a *AmazonExtractor) Matches(msg emailtypes.RawMessage) bool {
	a.domain = "amazon."
	return a.matchDomain(msg.From)
}

var (
	amazonOrderRe    = regexp.MustCompile(`(?i)order\s*#\s*([A-Z0-9\-]{3,})`)
	amazonOrderURLRe = regexp.MustCompile(`(?i)orderID=3D?([0-9]{8,})`)
	amazonShipmentRe = regexp.MustCompile(`(?i)shipment\s*(?:id|#)?\s*([A-Z0-9]+)`)
)

func (a *AmazonExtractor) Extract(msg emailtypes.RawMessage) (emailtypes.ExtractedReceipt, error) {
	body := merchantBody(msg)
	r := baseExtract(msg, "Amazon", "materials")
	r.FieldProvenance = map[string]emailtypes.Provenance{}
	if m := amazonOrderRe.FindStringSubmatch(body); len(m) > 1 {
		r.OrderRef = m[1]
		r.FieldProvenance["order_ref"] = emailtypes.ProvHTML
	}
	if m := amazonShipmentRe.FindStringSubmatch(body); len(m) > 1 {
		r.OrderRef = r.OrderRef + "-SHIP-" + m[1]
		r.FieldProvenance["shipment"] = emailtypes.ProvHTML
	}
	if total, ok := parseReceiptTotalFromBodies(msg.BodyText, ExtractVisibleText(msg.BodyHTML)); ok {
		r.TotalMinor = total
		r.FieldProvenance["total"] = emailtypes.ProvHTML
	}
	if r.OrderRef == "" {
		if m := amazonOrderURLRe.FindStringSubmatch(body); len(m) > 1 {
			r.OrderRef = m[1]
			r.FieldProvenance["order_ref"] = emailtypes.ProvHTML
		}
	}
	r.LineItems = parseAmazonLines(ExtractVisibleText(body))
	if len(r.LineItems) > 0 {
		r.VATMinor = sumVAT(r.LineItems)
		r.FieldProvenance["line_items"] = emailtypes.ProvHTML
	}
	r.Category = "materials"
	r.FamilyKey = FamilyKey(r.Supplier, r.OrderRef, r.IssuedAt, r.TotalMinor, r.Currency)
	return r, nil
}

func parseAmazonLines(body string) []model.LineItem {
	matches := lineItemRe.FindAllStringSubmatch(body, -1)
	items := make([]model.LineItem, 0, len(matches))
	for i, m := range matches {
		minor, ok := ParsePoundsToMinor(m[2])
		if !ok {
			continue
		}
		items = append(items, model.LineItem{
			Description: strings.TrimSpace(m[1]),
			NetAmount:   model.NewMoney(minor, "GBP"),
		})
		_ = i
	}
	return items
}

func sumVAT(items []model.LineItem) int64 {
	var v int64
	for _, li := range items {
		v += li.VATAmount.Amount
	}
	return v
}

// ScrewfixExtractor handles Screwfix e-receipts.
type ScrewfixExtractor struct{ domainExtractor }

func (s *ScrewfixExtractor) Name() string { return "screwfix" }

func (s *ScrewfixExtractor) Matches(msg emailtypes.RawMessage) bool {
	s.domain = "screwfix.com"
	return s.matchDomain(msg.From)
}

func (s *ScrewfixExtractor) Extract(msg emailtypes.RawMessage) (emailtypes.ExtractedReceipt, error) {
	return merchantHTMLExtract(msg, "Screwfix", "materials", screwfixOrderRe)
}

var screwfixOrderRe = regexp.MustCompile(`(?i)(?:order\s*ref[:\s]*|credit\s*note\s+)([A-Z0-9]+)`)

// ToolstationExtractor handles Toolstation receipts.
type ToolstationExtractor struct{ domainExtractor }

func (t *ToolstationExtractor) Name() string { return "toolstation" }

func (t *ToolstationExtractor) Matches(msg emailtypes.RawMessage) bool {
	t.domain = "toolstation.com"
	return t.matchDomain(msg.From)
}

func (t *ToolstationExtractor) Extract(msg emailtypes.RawMessage) (emailtypes.ExtractedReceipt, error) {
	r, err := merchantHTMLExtract(msg, "Toolstation", "materials", toolstationOrderRe)
	if err != nil {
		return r, err
	}
	if r.OrderRef == "" {
		if m := toolstationSubjectRe.FindStringSubmatch(msg.Subject); len(m) > 1 {
			r.OrderRef = m[1]
			r.FieldProvenance["order_ref"] = emailtypes.ProvHTML
		}
	}
	return r, nil
}

var (
	toolstationOrderRe = regexp.MustCompile(
		`(?i)(?:order\s+(YWW[0-9]+|REDACTED_ORDER)|order\s*number[:\s]*([0-9]+)|of\s+order\s+([0-9]+))`,
	)
	toolstationSubjectRe = regexp.MustCompile(`(?i)order\s+(YWW[0-9]+|REDACTED_ORDER)`)
)

// BandQExtractor handles B&Q / diy.com receipts.
type BandQExtractor struct{}

func (b *BandQExtractor) Name() string { return "bandq" }

func (b *BandQExtractor) Matches(msg emailtypes.RawMessage) bool {
	from := strings.ToLower(msg.From)
	return strings.Contains(from, "diy.com") || strings.Contains(from, "bandq")
}

func (b *BandQExtractor) Extract(msg emailtypes.RawMessage) (emailtypes.ExtractedReceipt, error) {
	return merchantHTMLExtract(msg, "B&Q", "materials", bandqOrderRe)
}

var bandqOrderRe = regexp.MustCompile(`(?i)receipt\s*no[:\s]*([0-9]+)`)

// FuelExtractor handles Shell/BP/Esso fuel app receipts.
type FuelExtractor struct{}

func (f *FuelExtractor) Name() string { return "fuel" }

func (f *FuelExtractor) Matches(msg emailtypes.RawMessage) bool {
	from := strings.ToLower(msg.From)
	return strings.Contains(from, "bp.com") || strings.Contains(from, "shell.com") || strings.Contains(from, "esso.com")
}

func (f *FuelExtractor) Extract(msg emailtypes.RawMessage) (emailtypes.ExtractedReceipt, error) {
	body := coalesceBody(msg)
	r := baseExtract(msg, fuelSupplier(msg.From), "motor")
	r.FieldProvenance = map[string]emailtypes.Provenance{}
	if total, ok := ParseReceiptTotal(body); ok {
		r.TotalMinor = total
		r.FieldProvenance["total"] = emailtypes.ProvHTML
	}
	net := r.TotalMinor * 5 / 6
	vat := r.TotalMinor - net
	r.LineItems = []model.LineItem{{
		Description: "Fuel",
		NetAmount:   model.NewMoney(net, "GBP"),
		VATAmount:   model.NewMoney(vat, "GBP"),
	}}
	r.VATMinor = vat
	r.VATExempt = false
	r.FamilyKey = FamilyKey(r.Supplier, r.OrderRef, r.IssuedAt, r.TotalMinor, r.Currency)
	return r, nil
}

func fuelSupplier(from string) string {
	from = strings.ToLower(from)
	switch {
	case strings.Contains(from, "bp.com"):
		return "BP"
	case strings.Contains(from, "shell.com"):
		return "Shell"
	case strings.Contains(from, "esso.com"):
		return "Esso"
	default:
		return "Fuel"
	}
}

// UtilityExtractor handles recurring utility invoices.
type UtilityExtractor struct{}

func (u *UtilityExtractor) Name() string { return "utility" }

func (u *UtilityExtractor) Matches(msg emailtypes.RawMessage) bool {
	from := strings.ToLower(msg.From)
	return strings.Contains(from, "edfenergy") || strings.Contains(from, "octopus.energy") ||
		strings.Contains(from, "britishgas")
}

func (u *UtilityExtractor) Extract(msg emailtypes.RawMessage) (emailtypes.ExtractedReceipt, error) {
	body := coalesceBody(msg)
	r := baseExtract(msg, utilitySupplier(msg.From), "overheads")
	r.FieldProvenance = map[string]emailtypes.Provenance{}
	if total, ok := ParseJSONLDTotal(body); ok {
		r.TotalMinor = total
		r.FieldProvenance["total"] = emailtypes.ProvSchemaOrg
	} else if tableTotal, tableOK := ParseReceiptTotal(body); tableOK {
		r.TotalMinor = tableTotal
		r.FieldProvenance["total"] = emailtypes.ProvHTML
	}
	items := parseLineItems(ExtractVisibleText(body))
	for i := range items {
		v := items[i].NetAmount.Amount / 5
		items[i].VATAmount = model.NewMoney(v, "GBP")
	}
	r.LineItems = items
	if len(items) > 0 {
		r.VATMinor = sumVAT(items)
		r.FieldProvenance["line_items"] = emailtypes.ProvHTML
	}
	accountRefRe := regexp.MustCompile(`(?i)(?:account\s*ref[:\s]*|accountNumber["\s:]+)([A-Z0-9]+)`)
	if m := accountRefRe.FindStringSubmatch(body); len(m) > 1 {
		r.OrderRef = m[1]
		r.FieldProvenance["order_ref"] = emailtypes.ProvHTML
	}
	r.FamilyKey = FamilyKey(r.Supplier, r.OrderRef, r.IssuedAt, r.TotalMinor, r.Currency)
	return r, nil
}

func utilitySupplier(from string) string {
	from = strings.ToLower(from)
	switch {
	case strings.Contains(from, "edf"):
		return "EDF"
	case strings.Contains(from, "octopus"):
		return "Octopus Energy"
	case strings.Contains(from, "britishgas"):
		return "British Gas"
	default:
		return "Utility"
	}
}

func merchantHTMLExtract(
	msg emailtypes.RawMessage,
	supplier, category string,
	orderRe *regexp.Regexp,
) (emailtypes.ExtractedReceipt, error) {
	body := coalesceBody(msg)
	r := baseExtract(msg, supplier, category)
	r.FieldProvenance = map[string]emailtypes.Provenance{}
	if ref := firstCapture(orderRe, body); ref != "" {
		r.OrderRef = ref
		r.FieldProvenance["order_ref"] = emailtypes.ProvHTML
	}
	if total, ok := parseReceiptTotalFromBodies(msg.BodyText, ExtractVisibleText(msg.BodyHTML)); ok {
		r.TotalMinor = total
		r.FieldProvenance["total"] = emailtypes.ProvHTML
	}
	items := parseLineItems(ExtractVisibleText(body))
	for i := range items {
		vat := items[i].NetAmount.Amount / 5
		items[i].VATAmount = model.NewMoney(vat, "GBP")
	}
	r.LineItems = items
	if len(items) > 0 {
		r.VATMinor = sumVAT(items)
		r.FieldProvenance["line_items"] = emailtypes.ProvHTML
	}
	r.FamilyKey = FamilyKey(r.Supplier, r.OrderRef, r.IssuedAt, r.TotalMinor, r.Currency)
	return r, nil
}

func baseExtract(msg emailtypes.RawMessage, supplier, category string) emailtypes.ExtractedReceipt {
	return emailtypes.ExtractedReceipt{
		Supplier: supplier,
		IssuedAt: msg.Date,
		Currency: "GBP",
		Category: category,
	}
}

func coalesceBody(msg emailtypes.RawMessage) string {
	return merchantBody(msg)
}

func merchantBody(msg emailtypes.RawMessage) string {
	if msg.BodyText != "" && msg.BodyHTML != "" {
		return msg.BodyText + "\n" + ExtractVisibleText(msg.BodyHTML)
	}
	if msg.BodyHTML != "" {
		return ExtractVisibleText(msg.BodyHTML)
	}
	return msg.BodyText
}

func firstCapture(re *regexp.Regexp, s string) string {
	m := re.FindStringSubmatch(s)
	for i := 1; i < len(m); i++ {
		if m[i] != "" {
			return m[i]
		}
	}
	return ""
}

func parseReceiptTotalFromBodies(bodies ...string) (int64, bool) {
	for _, body := range bodies {
		if body == "" {
			continue
		}
		if minor, ok := ParseReceiptTotal(body); ok {
			return minor, true
		}
	}
	return 0, false
}
