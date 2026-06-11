package dedup_test

import (
	"testing"
	"time"

	"github.com/jimbery/receipt/internal/extract"
)

func TestFamilyKey_OrderRefPreferred(t *testing.T) {
	day := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	got := extract.FamilyKey("Amazon", "ORD-99", day, 1299, "GBP")
	if got != "amazon|ORD-99" {
		t.Fatalf("got %q", got)
	}
}

func TestFamilyKey_FingerprintFallback(t *testing.T) {
	day := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	got := extract.FamilyKey("Cafe", "", day, 350, "GBP")
	want := "cafe|2026-03-10|350|GBP"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestFamilyKey_AmazonShipmentSplit(t *testing.T) {
	day := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	a := extract.FamilyKey("Amazon", "ORD1-SHIP-A", day, 1500, "GBP")
	b := extract.FamilyKey("Amazon", "ORD1-SHIP-B", day, 2500, "GBP")
	if a == b {
		t.Fatal("shipment-level keys must differ")
	}
}
