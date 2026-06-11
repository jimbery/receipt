package extract_test

import (
	"testing"
	"time"

	"github.com/jimbery/receipt/internal/emailtypes"
	"github.com/jimbery/receipt/internal/extract"
)

func TestMalformedFixtures_DeclaredOutcomes(t *testing.T) {
	reg := extract.NewRegistry()
	day := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name  string
		msg   emailtypes.RawMessage
		grade emailtypes.CompletenessGrade
	}{
		{
			name: "truncated HTML",
			msg: emailtypes.RawMessage{
				ID: "trunc", From: "shop@example.com", Date: day,
				BodyHTML: `<table><tr><td>Total</td><td>£`,
			},
			grade: emailtypes.GradeEnvelope,
		},
		{
			name: "image-only PDF",
			msg: emailtypes.RawMessage{
				ID: "imgpdf", From: "shop@example.com", Date: day,
				Attachments: []emailtypes.Attachment{{
					Filename: "scan.pdf", ContentType: "application/pdf",
					Content: []byte("%PDF-1.1 image only binary \x00\x01"),
				}},
			},
			grade: emailtypes.GradeRequiresOCR,
		},
		{
			name: "encrypted PDF",
			msg: emailtypes.RawMessage{
				ID: "encpdf", From: "shop@example.com", Date: day,
				Attachments: []emailtypes.Attachment{{
					Filename: "locked.pdf", ContentType: "application/pdf",
					Content: []byte("%PDF-1.4 /Encrypt metadata"),
				}},
			},
			grade: emailtypes.GradeRequiresOCR,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, ok := reg.Extract(tc.msg, emailtypes.KindPurchaseReceipt)
			if !ok && tc.grade != emailtypes.GradeRequiresOCR {
				t.Fatal("expected extraction attempt")
			}
			if ok && out.Grade != tc.grade {
				t.Fatalf("grade %s want %s", out.Grade, tc.grade)
			}
			if ok && out.TotalMinor > 0 && tc.grade == emailtypes.GradeRequiresOCR {
				t.Fatal("must not fabricate total on OCR route")
			}
		})
	}
}
