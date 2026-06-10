package similarity_test

import (
	"testing"

	"github.com/jimbery/receipt/internal/match/similarity"
)

func FuzzJaroWinkler_Bounded(f *testing.F) {
	f.Add("screwfix", "screwfix")
	f.Add("martha", "marhta")
	f.Fuzz(func(t *testing.T, a, b string) {
		got := similarity.JaroWinkler(a, b)
		if got < 0 || got > 1 {
			t.Fatalf("JaroWinkler(%q,%q)=%f outside [0,1]", a, b, got)
		}
		if a == b && got != 1 {
			t.Fatalf("identical strings want 1, got %f", got)
		}
	})
}

func FuzzTokenSetRatio_Bounded(f *testing.F) {
	f.Add("screwfix london", "screwfix")
	f.Add("a b", "b a")
	f.Fuzz(func(t *testing.T, a, b string) {
		got := similarity.TokenSetRatio(a, b)
		if got < 0 || got > 1 {
			t.Fatalf("TokenSetRatio(%q,%q)=%f outside [0,1]", a, b, got)
		}
	})
}
