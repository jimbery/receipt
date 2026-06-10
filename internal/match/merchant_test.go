package match_test

import (
	"testing"

	"github.com/jimbery/receipt/internal/match"
)

func TestBasicMerchantResolver_Table(t *testing.T) {
	t.Parallel()
	r := match.BasicMerchantResolver{}
	cases := []struct {
		txn, supplier string
		min           float64
	}{
		{"SCREWFIX 1234 LON", "Screwfix", 0.8},
		{"SQ *CAFE LONDON", "Cafe", 0.7},
		{"AMAZON UK MARKETPLACE", "Amazon", 0.75},
		{"TESCO STORES 3344", "Tesco", 0.8},
		{"UNRELATED MERCHANT", "Screwfix", 0},
	}
	for _, tc := range cases {
		t.Run(tc.txn, func(t *testing.T) {
			t.Parallel()
			got := r.Similarity(tc.txn, tc.supplier, "")
			if got < tc.min {
				t.Fatalf("similarity %.3f < %.3f", got, tc.min)
			}
		})
	}
}

func FuzzMerchantResolver_Bounded(f *testing.F) {
	f.Add("SCREWFIX", "Screwfix", "")
	f.Fuzz(func(t *testing.T, txn, supplier, hint string) {
		r := match.BasicMerchantResolver{}
		got := r.Similarity(txn, supplier, hint)
		if got < 0 || got > 1 {
			t.Fatalf("score %.3f outside [0,1]", got)
		}
	})
}
