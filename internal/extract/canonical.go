package extract

import (
	"errors"

	"github.com/jimbery/receipt/internal/emailtypes"
	"github.com/jimbery/receipt/internal/model"
)

// Canonicalise maps an extracted receipt to ADR-001 Receipt.
func Canonicalise(e emailtypes.ExtractedReceipt, mailboxID string) (model.Receipt, error) {
	if e.Supplier == "" {
		return model.Receipt{}, errors.New("missing supplier")
	}
	if e.TotalMinor <= 0 {
		return model.Receipt{}, errors.New("missing total")
	}
	id := e.SourceMessageID
	if id == "" {
		id = FamilyKey(e.Supplier, e.OrderRef, e.IssuedAt, e.TotalMinor, e.Currency)
	}
	ref := e.OrderRef
	if ref == "" {
		ref = id
	}
	cat := e.Category
	if cat == "" {
		cat = "uncategorised"
	}
	grade := Grade(e)
	if grade != emailtypes.GradeItemised && len(e.LineItems) > 0 && !arithmeticOK(e) {
		grade = emailtypes.GradePartial
	}
	_ = grade // grade stored via line item presence for Phase 0 Itemised()
	return model.Receipt{
		ID:          id,
		Source:      model.ReceiptSourceEmail,
		Supplier:    e.Supplier,
		ExternalRef: ref,
		IssuedAt:    model.TimestampFromTime(e.IssuedAt),
		Total:       model.NewMoney(e.TotalMinor, e.Currency),
		LineItems:   e.LineItems,
		Category:    cat,
	}, nil
}
