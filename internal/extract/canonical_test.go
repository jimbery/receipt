package extract_test

import (
	"testing"
	"time"

	"github.com/jimbery/receipt/internal/emailtypes"
	"github.com/jimbery/receipt/internal/extract"
	"github.com/jimbery/receipt/internal/model"
)

func TestCanonicalise_RejectsMissingFields(t *testing.T) {
	_, err := extract.Canonicalise(emailtypes.ExtractedReceipt{}, "mb-1")
	if err == nil {
		t.Fatal("expected error for empty receipt")
	}
}

func TestCanonicalise_MapsFields(t *testing.T) {
	now := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	in := emailtypes.ExtractedReceipt{
		SourceMessageID: "msg-1",
		Supplier:        "Screwfix",
		OrderRef:        "SF-99",
		TotalMinor:      1200,
		Currency:        "GBP",
		IssuedAt:        now,
		Category:        extract.SACategoryMaterials,
		LineItems: []model.LineItem{
			{NetAmount: model.NewMoney(1000, "GBP"), VATAmount: model.NewMoney(200, "GBP")},
		},
		VATMinor: 200,
	}
	r, err := extract.Canonicalise(in, "mb-1")
	if err != nil {
		t.Fatal(err)
	}
	if r.Supplier != "Screwfix" || r.Total.Amount != 1200 || r.Category != extract.SACategoryMaterials {
		t.Fatalf("unexpected receipt: %+v", r)
	}
}

func TestGrade_NeverItemisedWhenLinesDoNotReconcile(t *testing.T) {
	now := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	line := model.LineItem{NetAmount: model.NewMoney(1000, "GBP"), VATAmount: model.NewMoney(200, "GBP")}
	cases := []emailtypes.ExtractedReceipt{
		{
			Supplier: "Screwfix", IssuedAt: now, TotalMinor: 5000, Currency: "GBP",
			LineItems: []model.LineItem{line}, VATMinor: 200,
		},
		{
			Supplier: "Screwfix", IssuedAt: now, TotalMinor: 1000, Currency: "GBP",
			LineItems: []model.LineItem{{NetAmount: model.NewMoney(800, "GBP")}}, VATMinor: 0,
		},
		{
			Supplier: "Screwfix", IssuedAt: now, TotalMinor: 3000, Currency: "GBP",
			LineItems: []model.LineItem{
				{NetAmount: model.NewMoney(500, "GBP"), VATAmount: model.NewMoney(100, "GBP")},
				{NetAmount: model.NewMoney(700, "GBP"), VATAmount: model.NewMoney(140, "GBP")},
			},
			VATMinor: 240,
		},
	}
	for i, in := range cases {
		if extract.Grade(in) == emailtypes.GradeItemised {
			t.Fatalf("case %d: reconciling lines must not grade itemised", i)
		}
		if _, err := extract.Canonicalise(in, "mb-1"); err != nil {
			t.Fatalf("case %d: canonicalise: %v", i, err)
		}
	}
}
