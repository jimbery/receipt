package extract_test

import (
	"testing"

	"github.com/jimbery/receipt/internal/extract"
)

func TestParseJSONLDTotal(t *testing.T) {
	html := `<html><script type="application/ld+json">{"@type":"Order","price":"24.99"}</script></html>`
	got, ok := extract.ParseJSONLDTotal(html)
	if !ok || got != 2499 {
		t.Fatalf("got %d ok=%v", got, ok)
	}
}

func TestParseHTMLTableTotal(t *testing.T) {
	html := `<table><tr><td>Subtotal</td><td>£10.00</td></tr>
<tr><td>Order total</td><td>£12.99</td></tr></table>`
	got, ok := extract.ParseHTMLTableTotal(html)
	if !ok || got != 1299 {
		t.Fatalf("got %d ok=%v", got, ok)
	}
}
