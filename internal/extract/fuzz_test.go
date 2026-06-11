package extract_test

import (
	"testing"

	"github.com/jimbery/receipt/internal/extract"
)

func FuzzParsePoundsToMinor_NoPanic(f *testing.F) {
	f.Add("12.99")
	f.Add("-5.00")
	f.Add("not money")
	f.Fuzz(func(t *testing.T, s string) {
		if len(s) > 4096 {
			return
		}
		_, _ = extract.ParsePoundsToMinor(s)
	})
}

func FuzzParseJSONLDTotal_NoPanic(f *testing.F) {
	f.Add(`<script type="application/ld+json">{"price":"1.00"}</script>`)
	f.Fuzz(func(t *testing.T, html string) {
		if len(html) > 64*1024 {
			return
		}
		_, _ = extract.ParseJSONLDTotal(html)
	})
}

func FuzzExtractPDFText_NoPanic(f *testing.F) {
	f.Add([]byte("BT (Total £9.99) Tj ET"))
	f.Add([]byte("%PDF-1.1 not valid"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 256*1024 {
			return
		}
		_ = extract.ExtractPDFText(data)
	})
}

func FuzzParseHTMLTableTotal_NoPanic(f *testing.F) {
	f.Add(`<table><tr><td>Total</td><td>£12.99</td></tr></table>`)
	f.Add("<table><tr><td>broken")
	f.Fuzz(func(t *testing.T, html string) {
		if len(html) > 64*1024 {
			return
		}
		_, _ = extract.ParseHTMLTableTotal(html)
	})
}

func FuzzExtractVisibleText_NoPanic(f *testing.F) {
	f.Add(`<div><p>Hello <b>world</b></p></div>`)
	f.Fuzz(func(t *testing.T, html string) {
		if len(html) > 64*1024 {
			return
		}
		_ = extract.ExtractVisibleText(html)
	})
}
