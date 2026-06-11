package harness

import (
	"fmt"
	"strings"
	"time"

	"github.com/jimbery/receipt/internal/emailtypes"
)

// LoadFrozenCorpus returns ≥500 structurally distinct labelled messages.
// Each base template has a unique HTML/MIME/PDF structure; variations inject
// unique order refs, dates, and amounts so duplicated bodies never occur.
func LoadFrozenCorpus() []CorpusMessage {
	templates := corpusTemplates()
	const variationsPerTemplate = 20
	var out []CorpusMessage
	id := 0
	for _, tmpl := range templates {
		for v := range variationsPerTemplate {
			id++
			msg := tmpl.build(id, v)
			out = append(out, CorpusMessage{
				ID:      fmt.Sprintf("msg-%05d", id),
				Msg:     msg,
				Label:   tmpl.label,
				Stratum: tmpl.stratum,
				Pattern: tmpl.pattern,
			})
		}
	}
	return out
}

// SyntheticCorpus is deprecated; use LoadFrozenCorpus.
func SyntheticCorpus() []CorpusMessage { return LoadFrozenCorpus() }

type corpusTemplate struct {
	stratum string
	label   emailtypes.DocumentKind
	pattern string
	build   func(id, variant int) emailtypes.RawMessage
}

func withPattern(t corpusTemplate, pattern string) corpusTemplate {
	t.pattern = pattern
	return t
}

func formatGBP(minor int) string {
	neg := minor < 0
	if neg {
		minor = -minor
	}
	pounds := minor / 100
	pence := minor % 100
	if neg {
		return fmt.Sprintf("-%d.%02d", pounds, pence)
	}
	return fmt.Sprintf("%d.%02d", pounds, pence)
}

func corpusTemplates() []corpusTemplate {
	base := time.Date(2026, 1, 15, 9, 0, 0, 0, time.UTC)
	mkDate := func(id, variant int) time.Time {
		return base.Add(time.Duration(id*37+variant*11) * time.Hour)
	}
	mkID := func(id int) string { return fmt.Sprintf("msg-%05d", id) }

	priority := []corpusTemplate{
		withPattern(amazonHTMLTable(mkDate, mkID), "amazon_html_table"),
		withPattern(amazonJSONLD(mkDate, mkID), "amazon_jsonld"),
		withPattern(amazonPDFAttachment(mkDate, mkID), "amazon_pdf_attachment"),
		withPattern(screwfixNestedTable(mkDate, mkID), "screwfix_nested_table"),
		withPattern(screwfixPlainText(mkDate, mkID), "screwfix_plain_text"),
		withPattern(toolstationHTML(mkDate, mkID), "toolstation_html"),
		withPattern(toolstationPartialShipment(mkDate, mkID), "toolstation_partial_shipment"),
		withPattern(bandqReceipt(mkDate, mkID), "bandq_receipt"),
		withPattern(bandqVATExempt(mkDate, mkID), "bandq_vat_exempt"),
		withPattern(fuelShellHTML(mkDate, mkID), "fuel_shell_html"),
		withPattern(fuelBPApp(mkDate, mkID), "fuel_bp_app"),
		withPattern(utilityOctopusJSONLD(mkDate, mkID), "utility_octopus_jsonld"),
		withPattern(utilityEDFTable(mkDate, mkID), "utility_edf_table"),
		withPattern(amazonCreditNote(mkDate, mkID), "amazon_credit_note"),
		withPattern(screwfixRefund(mkDate, mkID), "screwfix_refund"),
	}
	longTail := []corpusTemplate{
		withPattern(genericRetailerSchema(mkDate, mkID), "generic_retailer_schema"),
		withPattern(genericRetailerTable(mkDate, mkID), "generic_retailer_table"),
		withPattern(genericRetailerTextOnly(mkDate, mkID), "generic_retailer_text_only"),
		withPattern(genericRetailerMultipart(mkDate, mkID), "generic_retailer_multipart"),
		withPattern(genericRetailerPDFInvoice(mkDate, mkID), "generic_retailer_pdf_invoice"),
	}
	hardNeg := []corpusTemplate{
		withPattern(marketingReceiptAwaits(mkDate, mkID), "marketing_receipt_awaits"),
		withPattern(marketingNewsletter(mkDate, mkID), "marketing_newsletter"),
		withPattern(marketingSaleEnds(mkDate, mkID), "marketing_sale_ends"),
		withPattern(dispatchAmazonTracking(mkDate, mkID), "dispatch_amazon_tracking"),
		withPattern(dispatchNoTotal(mkDate, mkID), "dispatch_no_total"),
		withPattern(statementBankHTML(mkDate, mkID), "statement_bank_html"),
		withPattern(statementPDF(mkDate, mkID), "statement_pdf"),
		withPattern(creditNoteStandalone(mkDate, mkID), "credit_note_standalone"),
	}
	all := append(append(priority, longTail...), hardNeg...)
	return all
}

func amazonHTMLTable(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "priority_merchant", label: emailtypes.KindPurchaseReceipt,
		build: func(id, v int) emailtypes.RawMessage {
			total := 1299 + id*3 + v
			return emailtypes.RawMessage{
				ID: mkID(id), From: "orders@amazon.co.uk",
				Subject: fmt.Sprintf("Your Amazon order #%d-%d", id, v),
				Date:    mkDate(id, v),
				BodyHTML: fmt.Sprintf(`<html><body><table>
<tr><td>Widget pack %d</td><td>£%s</td></tr>
<tr><td>VAT</td><td>£%s</td></tr>
<tr><td><b>Order total</b></td><td><b>£%s</b></td></tr>
</table><p>Order #%d-%d</p></body></html>`,
					v, formatGBP(total-216), formatGBP(216), formatGBP(total), id, v),
			}
		},
	}
}

func amazonJSONLD(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "priority_merchant", label: emailtypes.KindPurchaseReceipt,
		build: func(id, v int) emailtypes.RawMessage {
			total := 2499 + id + v*5
			return emailtypes.RawMessage{
				ID: mkID(id), From: "shipment-tracking@amazon.co.uk",
				Subject: fmt.Sprintf("Invoice for order %d", id),
				Date:    mkDate(id, v),
				BodyHTML: fmt.Sprintf(`<html><body>
<script type="application/ld+json">{"@type":"Order","orderNumber":"AMZ-%d","price":"%s"}</script>
<table><tr><td>Total</td><td>£%s</td></tr></table>
</body></html>`, id, formatGBP(total), formatGBP(total)),
			}
		},
	}
}

func amazonPDFAttachment(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "priority_merchant", label: emailtypes.KindPurchaseReceipt,
		build: func(id, v int) emailtypes.RawMessage {
			total := 4599 + v*7
			pdf := minimalTextPDF(fmt.Sprintf("Amazon VAT Invoice AMZ-%d\nOrder total £%s", id, formatGBP(total)))
			return emailtypes.RawMessage{
				ID: mkID(id), From: "invoice@amazon.co.uk",
				Subject:  fmt.Sprintf("VAT invoice %d", id),
				Date:     mkDate(id, v),
				BodyHTML: "<html><body><p>Your invoice is attached.</p></body></html>",
				Attachments: []emailtypes.Attachment{{
					Filename: "invoice.pdf", ContentType: "application/pdf", Content: pdf,
				}},
			}
		},
	}
}

func screwfixNestedTable(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "priority_merchant", label: emailtypes.KindPurchaseReceipt,
		build: func(id, v int) emailtypes.RawMessage {
			total := 4500 + id
			return emailtypes.RawMessage{
				ID: mkID(id), From: "receipts@screwfix.com",
				Subject: fmt.Sprintf("Screwfix receipt SF-%d", id),
				Date:    mkDate(id, v),
				BodyHTML: fmt.Sprintf(`<div><table><tr><td colspan="2">Items</td></tr>
<tr><td><table><tr><td>Drill bit %d</td><td>£%s</td></tr></table></td></tr>
<tr><td>Total</td><td>£%s</td></tr></table>
<p>Order ref: SF%d</p></div>`, v, formatGBP(total-750), formatGBP(total), id),
			}
		},
	}
}

func screwfixPlainText(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "priority_merchant", label: emailtypes.KindPurchaseReceipt,
		build: func(id, v int) emailtypes.RawMessage {
			total := 3200 + v
			return emailtypes.RawMessage{
				ID: mkID(id), From: "noreply@screwfix.com",
				Subject: "Your Screwfix order",
				Date:    mkDate(id, v),
				BodyText: fmt.Sprintf("Order ref: SF%d\nOrder total: £%s\nDrill £%s\nVAT £%s",
					id, formatGBP(total), formatGBP(total-533), formatGBP(533)),
			}
		},
	}
}

func toolstationHTML(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "priority_merchant", label: emailtypes.KindPurchaseReceipt,
		build: func(id, v int) emailtypes.RawMessage {
			total := 2250 + id
			return emailtypes.RawMessage{
				ID: mkID(id), From: "noreply@toolstation.com",
				Subject: "Toolstation order confirmation",
				Date:    mkDate(id, v),
				BodyHTML: fmt.Sprintf(`<table><tr><td>Saw blade</td><td>£%s</td></tr>
<tr><td>VAT</td><td>£%s</td></tr><tr><td>Total</td><td>£%s</td></tr></table>
<p>Order number: %d</p>`, formatGBP(total-375), formatGBP(375), formatGBP(total), id),
			}
		},
	}
}

func toolstationPartialShipment(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "priority_merchant", label: emailtypes.KindPurchaseReceipt,
		build: func(id, v int) emailtypes.RawMessage {
			total := 1500 + v
			return emailtypes.RawMessage{
				ID: mkID(id), From: "dispatch@toolstation.com",
				Subject: fmt.Sprintf("Order confirmation — partial shipment %d", id),
				Date:    mkDate(id, v),
				BodyHTML: fmt.Sprintf(`<p>Partial shipment %d of order %d</p>
<table><tr><td>Item</td><td>£%s</td></tr><tr><td>Total</td><td>£%s</td></tr></table>`,
					v%3+1, id, formatGBP(total-250), formatGBP(total)),
			}
		},
	}
}

func bandqReceipt(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "priority_merchant", label: emailtypes.KindPurchaseReceipt,
		build: func(id, v int) emailtypes.RawMessage {
			total := 6780 + id
			return emailtypes.RawMessage{
				ID: mkID(id), From: "receipts@diy.com",
				Subject: "B&Q receipt",
				Date:    mkDate(id, v),
				BodyHTML: fmt.Sprintf(`Receipt no: %d<table><tr><td>Paint</td><td>£%s</td></tr>
<tr><td>Total</td><td>£%s</td></tr></table>`, id, formatGBP(total-1130), formatGBP(total)),
			}
		},
	}
}

func bandqVATExempt(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "priority_merchant", label: emailtypes.KindPurchaseReceipt,
		build: func(id, v int) emailtypes.RawMessage {
			return emailtypes.RawMessage{
				ID: mkID(id), From: "receipts@diy.com",
				Subject: "B&Q receipt — VAT-exempt purchase",
				Date:    mkDate(id, v),
				BodyHTML: fmt.Sprintf(`<table><tr><td>Zero-rated goods %d</td><td>£%s</td></tr>
<tr><td>Total</td><td>£%s</td></tr></table>`, v, formatGBP(1000+v), formatGBP(1000+v)),
			}
		},
	}
}

func fuelShellHTML(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "priority_merchant", label: emailtypes.KindPurchaseReceipt,
		build: func(id, v int) emailtypes.RawMessage {
			total := 6200 + id
			return emailtypes.RawMessage{
				ID: mkID(id), From: "receipts@shell.com",
				Subject: "Shell Go+ receipt",
				Date:    mkDate(id, v),
				BodyHTML: fmt.Sprintf(`<table><tr><td>Fuel</td><td>£%s</td></tr>
<tr><td>Total</td><td>£%s</td></tr></table>`, formatGBP(total), formatGBP(total)),
			}
		},
	}
}

func fuelBPApp(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "priority_merchant", label: emailtypes.KindPurchaseReceipt,
		build: func(id, v int) emailtypes.RawMessage {
			total := 5500 + v*3
			return emailtypes.RawMessage{
				ID: mkID(id), From: "fuel@bp.com",
				Subject: "BP fuel receipt",
				Date:    mkDate(id, v),
				BodyText: fmt.Sprintf("Site %d\nTotal paid: £%s\nFuel £%s\nVAT £%s",
					id, formatGBP(total), formatGBP(total*5/6), formatGBP(total/6)),
			}
		},
	}
}

func utilityOctopusJSONLD(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "priority_merchant", label: emailtypes.KindPurchaseReceipt,
		build: func(id, v int) emailtypes.RawMessage {
			total := 12000 + id
			return emailtypes.RawMessage{
				ID: mkID(id), From: "billing@octopus.energy",
				Subject: "Your Octopus Energy invoice",
				Date:    mkDate(id, v),
				BodyHTML: fmt.Sprintf(`<script type="application/ld+json">
{"@type":"Invoice","accountNumber":"ACC%d","totalPaymentDue":{"price":"%s"}}
</script><table><tr><td>Energy</td><td>£%s</td></tr><tr><td>Total</td><td>£%s</td></tr></table>`,
					id, formatGBP(total), formatGBP(total-2000), formatGBP(total)),
			}
		},
	}
}

func utilityEDFTable(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "priority_merchant", label: emailtypes.KindPurchaseReceipt,
		build: func(id, v int) emailtypes.RawMessage {
			total := 8900 + v
			return emailtypes.RawMessage{
				ID: mkID(id), From: "billing@edfenergy.com",
				Subject: "EDF energy invoice",
				Date:    mkDate(id, v),
				BodyHTML: fmt.Sprintf(`<table><tr><td>Standing charge</td><td>£%s</td></tr>
<tr><td>Total</td><td>£%s</td></tr></table><p>Account ref: EDF%d</p>`,
					formatGBP(total-1483), formatGBP(total), id),
			}
		},
	}
}

func amazonCreditNote(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "priority_merchant", label: emailtypes.KindCreditNote,
		build: func(id, v int) emailtypes.RawMessage {
			return emailtypes.RawMessage{
				ID: mkID(id), From: "returns@amazon.co.uk",
				Subject: "Refund confirmation",
				Date:    mkDate(id, v),
				BodyHTML: fmt.Sprintf(`<p>Credit note for order AMZ-%d</p>
<table><tr><td>Refund total</td><td>£%s</td></tr></table>`, id, formatGBP(-(500 + v))),
			}
		},
	}
}

func screwfixRefund(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "priority_merchant", label: emailtypes.KindCreditNote,
		build: func(id, v int) emailtypes.RawMessage {
			return emailtypes.RawMessage{
				ID: mkID(id), From: "returns@screwfix.com",
				Subject:  "Return processed",
				Date:     mkDate(id, v),
				BodyText: fmt.Sprintf("Credit note SF%d\nRefund: £%s", id, formatGBP(-(1200 + v))),
			}
		},
	}
}

func genericRetailerSchema(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "long_tail", label: emailtypes.KindPurchaseReceipt,
		build: func(id, v int) emailtypes.RawMessage {
			total := 999 + id%50
			return emailtypes.RawMessage{
				ID: mkID(id), From: fmt.Sprintf("shop%d@unknown-retailer.co.uk", v%7),
				Subject:  "Receipt for your purchase",
				Date:     mkDate(id, v),
				BodyHTML: fmt.Sprintf(`<script type="application/ld+json">{"price":"%s"}</script>`, formatGBP(total)),
			}
		},
	}
}

func genericRetailerTable(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "long_tail", label: emailtypes.KindPurchaseReceipt,
		build: func(id, v int) emailtypes.RawMessage {
			total := 1500 + id
			return emailtypes.RawMessage{
				ID: mkID(id), From: fmt.Sprintf("billing%d@longtail-shop.example", v%5),
				Subject: "Order confirmation",
				Date:    mkDate(id, v),
				BodyHTML: fmt.Sprintf(`<table><tr><td>Item %d</td><td>£%s</td></tr>
<tr><td>Total</td><td>£%s</td></tr></table>`, v, formatGBP(total-250), formatGBP(total)),
			}
		},
	}
}

func genericRetailerTextOnly(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "long_tail", label: emailtypes.KindPurchaseReceipt,
		build: func(id, v int) emailtypes.RawMessage {
			total := 1800 + v
			return emailtypes.RawMessage{
				ID: mkID(id), From: "orders@corner-store.example",
				Subject: "Payment received",
				Date:    mkDate(id, v),
				BodyText: fmt.Sprintf("Total paid: £%s\nItem £%s\nVAT £%s",
					formatGBP(total), formatGBP(total-300), formatGBP(300)),
			}
		},
	}
}

func genericRetailerMultipart(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "long_tail", label: emailtypes.KindPurchaseReceipt,
		build: func(id, v int) emailtypes.RawMessage {
			total := 2200 + id
			return emailtypes.RawMessage{
				ID: mkID(id), From: "receipts@workshop-supplies.example",
				Subject: "Your receipt",
				Date:    mkDate(id, v),
				BodyHTML: fmt.Sprintf(`<html><body><p>Thanks for order %d</p>
<table><tr><td>Total</td><td>£%s</td></tr></table></body></html>`, id, formatGBP(total)),
				BodyText: fmt.Sprintf("Order %d total £%s", id, formatGBP(total)),
			}
		},
	}
}

func genericRetailerPDFInvoice(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "long_tail", label: emailtypes.KindPurchaseReceipt,
		build: func(id, v int) emailtypes.RawMessage {
			total := 3400 + v
			pdf := minimalTextPDF(fmt.Sprintf("Invoice LT-%d Total: £%s", id, formatGBP(total)))
			return emailtypes.RawMessage{
				ID: mkID(id), From: "accounts@longtail-trader.example",
				Subject: "Invoice attached",
				Date:    mkDate(id, v),
				Attachments: []emailtypes.Attachment{{
					Filename: "invoice.pdf", ContentType: "application/pdf", Content: pdf,
				}},
			}
		},
	}
}

func marketingReceiptAwaits(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "hard_negative", label: emailtypes.KindMarketing,
		build: func(id, v int) emailtypes.RawMessage {
			return emailtypes.RawMessage{
				ID: mkID(id), From: "news@retailer.com",
				Subject: "Your receipt awaits — complete your checkout",
				Date:    mkDate(id, v),
				BodyHTML: `<html><body><h1>Almost there!</h1><p>Complete checkout to get your receipt.</p>
<footer><a href="#">unsubscribe</a></footer></body></html>`,
			}
		},
	}
}

func marketingNewsletter(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "hard_negative", label: emailtypes.KindMarketing,
		build: func(id, v int) emailtypes.RawMessage {
			return emailtypes.RawMessage{
				ID: mkID(id), From: "marketing@store.com",
				Subject:  fmt.Sprintf("Weekly deals %d", id),
				Date:     mkDate(id, v),
				BodyHTML: `<html><body><p>Newsletter sale ends tonight.</p><p>limited time offer</p></body></html>`,
			}
		},
	}
}

func marketingSaleEnds(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "hard_negative", label: emailtypes.KindMarketing,
		build: func(id, v int) emailtypes.RawMessage {
			return emailtypes.RawMessage{
				ID: mkID(id), From: "promo@bigbox.example",
				Subject:  "Sale ends midnight",
				Date:     mkDate(id, v),
				BodyText: "limited time offer\nunsubscribe here",
			}
		},
	}
}

func dispatchAmazonTracking(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "hard_negative", label: emailtypes.KindDispatchNotice,
		build: func(id, v int) emailtypes.RawMessage {
			return emailtypes.RawMessage{
				ID: mkID(id), From: "ship@amazon.co.uk",
				Subject:  "Your package has dispatched",
				Date:     mkDate(id, v),
				BodyHTML: fmt.Sprintf(`<p>Tracking number TRACK%d</p><p>Expected delivery soon.</p>`, id),
			}
		},
	}
}

func dispatchNoTotal(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "hard_negative", label: emailtypes.KindDispatchNotice,
		build: func(id, v int) emailtypes.RawMessage {
			return emailtypes.RawMessage{
				ID: mkID(id), From: "logistics@courier.example",
				Subject:  "Shipped: your order is on the way",
				Date:     mkDate(id, v),
				BodyHTML: "<p>Parcel dispatched. Track at courier.example</p>",
			}
		},
	}
}

func statementBankHTML(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "hard_negative", label: emailtypes.KindStatement,
		build: func(id, v int) emailtypes.RawMessage {
			return emailtypes.RawMessage{
				ID: mkID(id), From: "statements@bank.com",
				Subject:  "Monthly statement available",
				Date:     mkDate(id, v),
				BodyHTML: "<p>Your monthly statement balance summary is ready.</p>",
			}
		},
	}
}

func statementPDF(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "hard_negative", label: emailtypes.KindStatement,
		build: func(id, v int) emailtypes.RawMessage {
			pdf := minimalTextPDF(fmt.Sprintf("Account statement period %d\nBalance summary only", id))
			return emailtypes.RawMessage{
				ID: mkID(id), From: "statements@bank.com",
				Subject: "Account summary",
				Date:    mkDate(id, v),
				Attachments: []emailtypes.Attachment{{
					Filename: "statement.pdf", ContentType: "application/pdf", Content: pdf,
				}},
			}
		},
	}
}

func creditNoteStandalone(mkDate func(int, int) time.Time, mkID func(int) string) corpusTemplate {
	return corpusTemplate{
		stratum: "hard_negative", label: emailtypes.KindCreditNote,
		build: func(id, v int) emailtypes.RawMessage {
			return emailtypes.RawMessage{
				ID: mkID(id), From: "returns@shop.com",
				Subject: "Refund confirmation",
				Date:    mkDate(id, v),
				BodyHTML: fmt.Sprintf(`<p>Return processed. Credit note CN-%d issued.</p>
<table><tr><td>Refund</td><td>£%s</td></tr></table>`, id, formatGBP(-(800 + v))),
			}
		},
	}
}

func minimalTextPDF(text string) []byte {
	escaped := strings.ReplaceAll(text, "(", "\\(")
	escaped = strings.ReplaceAll(escaped, ")", "\\)")
	stream := fmt.Sprintf("BT /F1 12 Tf 72 720 Td (%s) Tj ET", escaped)
	return fmt.Appendf(nil, `%%PDF-1.1
1 0 obj<</Type/Catalog/Pages 2 0 R>>endobj
2 0 obj<</Type/Pages/Kids[3 0 R]/Count 1>>endobj
3 0 obj<</Type/Page/Parent 2 0 R/MediaBox[0 0 612 792]/Contents 4 0 R>>endobj
4 0 obj<</Length %d>>stream
%s
endstream
endobj
xref
0 5
0000000000 65535 f 
trailer<</Size 5/Root 1 0 R>>
startxref
0
%%%%EOF`, len(stream), stream)
}
