package extract_test

import (
	"testing"

	"github.com/jimbery/receipt/internal/extract"
)

func TestParsePoundsToMinor(t *testing.T) {
	cases := []struct {
		in   string
		want int64
		ok   bool
	}{
		{"12.99", 1299, true},
		{"45", 4500, true},
		{"0.01", 1, true},
		{"-5.00", -500, true},
		{"-12.99", -1299, true},
		{"", 0, false},
		{"abc", 0, false},
	}
	for _, tc := range cases {
		got, ok := extract.ParsePoundsToMinor(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Fatalf("%q: got (%d,%v) want (%d,%v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}

func TestParsePoundsToMinor_CreditNoteNegative(t *testing.T) {
	// -5.00 must be exactly -500 minor, not -499 from float rounding.
	got, ok := extract.ParsePoundsToMinor("-5.00")
	if !ok || got != -500 {
		t.Fatalf("got %d ok=%v", got, ok)
	}
}
