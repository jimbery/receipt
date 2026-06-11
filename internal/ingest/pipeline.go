package ingest

import (
	"context"
	"time"

	"github.com/jimbery/receipt/internal/classify"
	"github.com/jimbery/receipt/internal/dedup"
	"github.com/jimbery/receipt/internal/emailtypes"
	"github.com/jimbery/receipt/internal/extract"
	"github.com/jimbery/receipt/internal/mail"
	"github.com/jimbery/receipt/internal/model"
)

// Pipeline wires source → classify → extract → dedup → canonicalise.
type Pipeline struct {
	Source     mail.MailSource
	Classifier *classify.Classifier
	Registry   *extract.Registry
}

// MailboxResult is the output of a mailbox ingestion run.
type MailboxResult struct {
	Receipts         []model.Receipt
	RequiresOCRCount int
}

// ProcessMailbox ingests messages since the given time.
func (p *Pipeline) ProcessMailbox(ctx context.Context, since time.Time) (MailboxResult, error) {
	msgs, err := p.Source.Fetch(ctx, since)
	if err != nil {
		return MailboxResult{}, err
	}
	var extracted []dedup.ExtractedWithProvenance
	var ocrRouted []emailtypes.ExtractedReceipt
	for _, msg := range msgs {
		kind := p.Classifier.Classify(msg)
		if !kind.ShouldExtract() && kind != emailtypes.KindMarketing && kind != emailtypes.KindStatement {
			continue
		}
		if kind == emailtypes.KindMarketing || kind == emailtypes.KindStatement {
			continue
		}
		ex, ok := p.Registry.Extract(msg, kind)
		if !ok {
			continue
		}
		if ex.Grade == emailtypes.GradeRequiresOCR {
			ocrRouted = append(ocrRouted, ex)
		}
		extracted = append(extracted, dedup.ExtractedWithProvenance{
			Extracted: ex,
			MessageID: msg.ID,
			Kind:      kind,
		})
	}
	resolved := dedup.ResolveFamilies(extracted)
	receipts := make([]model.Receipt, 0, len(resolved))
	for _, e := range resolved {
		r, canonErr := extract.Canonicalise(e, p.Source.ID())
		if canonErr != nil {
			continue
		}
		receipts = append(receipts, r)
	}
	return MailboxResult{
		Receipts:         receipts,
		RequiresOCRCount: extract.CountRequiresOCR(ocrRouted),
	}, nil
}
