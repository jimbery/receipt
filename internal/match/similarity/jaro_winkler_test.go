package similarity_test

import (
	"testing"

	"github.com/jimbery/receipt/internal/match/similarity"
)

func TestJaroWinkler_KnownValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		a, b string
		min  float64
	}{
		{"", "", 1},
		{"martha", "marhta", 0.9},
		{"dwayne", "duane", 0.8},
		{"screwfix", "screwfix", 1},
		{"screwfix", "totally different", 0},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			t.Parallel()
			got := similarity.JaroWinkler(tt.a, tt.b)
			if tt.a == tt.b && got != 1 {
				t.Fatalf("identical strings want 1, got %f", got)
			}
			if tt.min > 0 && got < tt.min {
				t.Fatalf("JaroWinkler(%q,%q)=%f want >= %f", tt.a, tt.b, got, tt.min)
			}
		})
	}
}
