package extract_test

import (
	"testing"
	"time"

	"github.com/jimbery/receipt/internal/emailtypes"
	"github.com/jimbery/receipt/internal/extract"
	"github.com/jimbery/receipt/internal/model"
)

func TestGrade_Rubric(t *testing.T) {
	now := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		name string
		in   emailtypes.ExtractedReceipt
		want emailtypes.CompletenessGrade
	}{
		{
			name: "itemised",
			in: emailtypes.ExtractedReceipt{
				Supplier: "Screwfix", IssuedAt: now, TotalMinor: 1200, Currency: "GBP",
				LineItems: []model.LineItem{
					{NetAmount: model.NewMoney(1000, "GBP"), VATAmount: model.NewMoney(200, "GBP")},
				},
				VATMinor: 200,
			},
			want: emailtypes.GradeItemised,
		},
		{
			name: "partial arithmetic",
			in: emailtypes.ExtractedReceipt{
				Supplier: "Screwfix", IssuedAt: now, TotalMinor: 5000, Currency: "GBP",
				LineItems: []model.LineItem{
					{NetAmount: model.NewMoney(1000, "GBP"), VATAmount: model.NewMoney(200, "GBP")},
				},
				VATMinor: 200,
			},
			want: emailtypes.GradePartial,
		},
		{
			name: "zero_line_envelope",
			in: emailtypes.ExtractedReceipt{
				Supplier: "BP", IssuedAt: now, TotalMinor: 5500, Currency: "GBP",
			},
			want: emailtypes.GradeEnvelope,
		},
		{
			name: "vat_exempt_itemised",
			in: emailtypes.ExtractedReceipt{
				Supplier: "B&Q", IssuedAt: now, TotalMinor: 1000, Currency: "GBP",
				VATExempt: true,
				LineItems: []model.LineItem{
					{NetAmount: model.NewMoney(1000, "GBP")},
				},
			},
			want: emailtypes.GradeItemised,
		},
		{
			name: "vat_number_absent_partial",
			in: emailtypes.ExtractedReceipt{
				Supplier: "Screwfix", IssuedAt: now, TotalMinor: 1000, Currency: "GBP",
				LineItems: []model.LineItem{
					{NetAmount: model.NewMoney(1000, "GBP")},
				},
			},
			want: emailtypes.GradePartial,
		},
		{
			name: "requires ocr",
			in: emailtypes.ExtractedReceipt{
				Grade: emailtypes.GradeRequiresOCR, AttachmentRefs: []string{"scan.pdf"},
			},
			want: emailtypes.GradeRequiresOCR,
		},
	}
	for _, tc := range cases {
		got := extract.Grade(tc.in)
		if got != tc.want {
			t.Fatalf("%s: got %s want %s", tc.name, got, tc.want)
		}
	}
}
