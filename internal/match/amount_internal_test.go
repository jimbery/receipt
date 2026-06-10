package match

import (
	"testing"

	"github.com/jimbery/receipt/internal/model"
)

func TestAmountSimilarity_TipBand15Pct(t *testing.T) {
	t.Parallel()
	cfg := DefaultConfig()
	txn := model.NewMoney(10000, "GBP")
	receipt := model.NewMoney(8500, "GBP")
	got := amountSimilarity(cfg, txn, receipt)
	if got < 0.5 {
		t.Fatalf("15%% tip should score within tip band, got %.3f", got)
	}
}
