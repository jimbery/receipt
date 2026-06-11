package extract_test

import (
	"testing"
	"time"

	"github.com/jimbery/receipt/internal/emailtypes"
	"github.com/jimbery/receipt/internal/extract"
)

func TestMerchantFixtures_EdgeForms(t *testing.T) {
	reg := extract.NewRegistry()
	day := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	cases := []struct {
		name      string
		msg       emailtypes.RawMessage
		kind      emailtypes.DocumentKind
		supplier  string
		wantGrade emailtypes.CompletenessGrade
	}{
		{
			name: "amazon partial shipment",
			msg: emailtypes.RawMessage{
				ID: "amz-ship", From: "orders@amazon.co.uk", Date: day,
				Subject: "Shipment dispatched",
				BodyHTML: `<p>Shipment ID SHIP9</p><p>Order #999</p>
<table><tr><td>Cable</td><td>£10.82</td></tr><tr><td>Total</td><td>£12.99</td></tr></table>`,
			},
			kind: emailtypes.KindPurchaseReceipt, supplier: "Amazon",
		},
		{
			name: "screwfix partial grade",
			msg: emailtypes.RawMessage{
				ID: "sf-part", From: "receipts@screwfix.com", Date: day,
				BodyHTML: `<table><tr><td>Total</td><td>£45.00</td></tr></table>`,
			},
			kind: emailtypes.KindPurchaseReceipt, supplier: "Screwfix",
			wantGrade: emailtypes.GradePartial,
		},
		{
			name: "bandq vat exempt",
			msg: emailtypes.RawMessage{
				ID: "bq-exempt", From: "receipts@diy.com", Date: day,
				BodyHTML: `<table><tr><td>Zero-rated</td><td>£10.00</td></tr>
<tr><td>Total</td><td>£10.00</td></tr></table>`,
			},
			kind: emailtypes.KindPurchaseReceipt, supplier: "B&Q",
		},
		{
			name: "amazon credit refund",
			msg: emailtypes.RawMessage{
				ID: "amz-cn", From: "returns@amazon.co.uk", Date: day,
				Subject:  "Refund confirmation",
				BodyHTML: `<p>Credit note</p><table><tr><td>Refund total</td><td>£-5.00</td></tr></table>`,
			},
			kind: emailtypes.KindCreditNote, supplier: "Amazon",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ex, ok := reg.Extract(tc.msg, tc.kind)
			if !ok {
				t.Fatal("expected extraction")
			}
			if ex.Supplier != tc.supplier {
				t.Fatalf("supplier %q", ex.Supplier)
			}
			if tc.wantGrade != "" && ex.Grade != tc.wantGrade {
				t.Fatalf("grade %s want %s", ex.Grade, tc.wantGrade)
			}
		})
	}
}
