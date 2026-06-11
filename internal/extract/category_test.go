package extract_test

import (
	"testing"

	"github.com/jimbery/receipt/internal/extract"
)

func TestSACategoryRuleTable(t *testing.T) {
	cases := []struct {
		merchant string
		want     string
	}{
		{"Amazon", extract.SACategoryMaterials},
		{"Screwfix", extract.SACategoryMaterials},
		{"Toolstation", extract.SACategoryMaterials},
		{"B&Q", extract.SACategoryMaterials},
		{"Shell", extract.SACategoryMotor},
		{"BP", extract.SACategoryMotor},
		{"Esso", extract.SACategoryMotor},
		{"Octopus", extract.SACategoryOverheads},
		{"EDF", extract.SACategoryOverheads},
	}
	for _, tc := range cases {
		got, ok := extract.SACategoryForMerchant(tc.merchant)
		if !ok {
			t.Fatalf("%s: missing SA category rule", tc.merchant)
		}
		if got != tc.want {
			t.Fatalf("%s: got %q want %q", tc.merchant, got, tc.want)
		}
	}
}
