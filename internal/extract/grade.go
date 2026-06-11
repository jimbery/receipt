package extract

import (
	"math"

	"github.com/jimbery/receipt/internal/emailtypes"
	"github.com/jimbery/receipt/internal/model"
)

const arithmeticToleranceMinor = 2

// Grade assigns CompletenessGrade per ADR-002 D3.
func Grade(r emailtypes.ExtractedReceipt) emailtypes.CompletenessGrade {
	if r.Grade == emailtypes.GradeRequiresOCR {
		return emailtypes.GradeRequiresOCR
	}
	hasIdentity := r.Supplier != "" && !r.IssuedAt.IsZero() && r.TotalMinor > 0
	if !hasIdentity {
		if len(r.AttachmentRefs) > 0 {
			return emailtypes.GradeRequiresOCR
		}
		return emailtypes.GradeEnvelope
	}
	if len(r.LineItems) == 0 {
		return emailtypes.GradeEnvelope
	}
	hasVAT := r.VATExempt || r.VATMinor > 0 || lineItemsHaveVAT(r.LineItems)
	if !hasVAT {
		return emailtypes.GradePartial
	}
	if !arithmeticOK(r) {
		return emailtypes.GradePartial
	}
	return emailtypes.GradeItemised
}

func lineItemsHaveVAT(items []model.LineItem) bool {
	for _, li := range items {
		if li.VATAmount.Amount > 0 {
			return true
		}
	}
	return false
}

func arithmeticOK(r emailtypes.ExtractedReceipt) bool {
	var sum int64
	for _, li := range r.LineItems {
		sum += li.NetAmount.Amount
		if li.VATAmount.Amount > 0 {
			sum += li.VATAmount.Amount
		}
	}
	diff := math.Abs(float64(sum - r.TotalMinor))
	return diff <= arithmeticToleranceMinor
}
