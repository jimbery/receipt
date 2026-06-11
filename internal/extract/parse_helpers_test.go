package extract_test

import (
	"testing"

	"github.com/jimbery/receipt/internal/extract"
)

func TestParseReceiptTotal_TableAndPaid(t *testing.T) {
	cases := []struct {
		body string
		want int64
	}{
		{`<table><tr><td>Total</td><td>£12.99</td></tr></table>`, 1299},
		{"Total paid: £42.00", 4200},
		{"Order total: £24.99", 2499},
		{"Total\n49.98 GBP", 4998},
		{"Total (inc. VAT) £9.99 Total (ex. VAT) £8.32", 999},
		{"Refund: £-12.00", -1200},
	}
	for _, tc := range cases {
		got, ok := extract.ParseReceiptTotal(tc.body)
		if !ok || got != tc.want {
			t.Fatalf("body %q: got %d ok=%v want %d", tc.body, got, ok, tc.want)
		}
	}
}
