package model_test

import (
	"testing"

	"github.com/jimbery/receipt/internal/model"
)

func TestMoney_WithinTolerance(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		a, b model.Money
		abs  int64
		rel  float64
		want bool
	}{
		{"exact", model.NewMoney(1000, "GBP"), model.NewMoney(1000, "GBP"), 0, 0, true},
		{"within abs", model.NewMoney(1000, "GBP"), model.NewMoney(1040, "GBP"), 50, 0, true},
		{"currency mismatch", model.NewMoney(1000, "GBP"), model.NewMoney(1000, "USD"), 100, 0.1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.a.WithinTolerance(tt.b, tt.abs, tt.rel)
			if got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
}
