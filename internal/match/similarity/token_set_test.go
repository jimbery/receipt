package similarity_test

import (
	"testing"

	"github.com/jimbery/receipt/internal/match/similarity"
)

func TestTokenSetRatio_KnownValues(t *testing.T) {
	t.Parallel()
	if got := similarity.TokenSetRatio("screwfix", "screwfix"); got != 1 {
		t.Fatalf("identical want 1, got %f", got)
	}
	if got := similarity.TokenSetRatio("screwfix london", "screwfix"); got <= 0 {
		t.Fatalf("subset want >0, got %f", got)
	}
	if got := similarity.TokenSetRatio("alpha", "beta"); got != 0 {
		t.Fatalf("disjoint want 0, got %f", got)
	}
}
