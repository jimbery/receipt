package ingest_test

import (
	"context"
	"testing"
	"time"

	"github.com/jimbery/receipt/internal/classify"
	"github.com/jimbery/receipt/internal/emailtypes"
	"github.com/jimbery/receipt/internal/extract"
	"github.com/jimbery/receipt/internal/ingest"
	iharness "github.com/jimbery/receipt/internal/ingest/harness"
	"github.com/jimbery/receipt/internal/mail"
)

type staticSource struct {
	msgs []emailtypes.RawMessage
}

func (s *staticSource) ID() string { return "determinism-lab" }

func (s *staticSource) Fetch(_ context.Context, _ time.Time) ([]emailtypes.RawMessage, error) {
	return s.msgs, nil
}

func TestPipeline_ProcessMailbox_Deterministic(t *testing.T) {
	corpus := iharness.LoadFrozenCorpus()
	msgs := make([]emailtypes.RawMessage, 0, 40)
	for _, c := range corpus {
		if c.Label == emailtypes.KindPurchaseReceipt {
			msgs = append(msgs, c.Msg)
			if len(msgs) >= 40 {
				break
			}
		}
	}
	if len(msgs) < 20 {
		t.Fatalf("need purchase receipts in corpus, got %d", len(msgs))
	}

	src := &staticSource{msgs: msgs}
	p := &ingest.Pipeline{
		Source:     src,
		Classifier: classify.New(),
		Registry:   extract.NewRegistry(),
	}

	var first []string
	for run := range 3 {
		result, err := p.ProcessMailbox(context.Background(), time.Time{})
		if err != nil {
			t.Fatalf("run %d: %v", run, err)
		}
		receipts := result.Receipts
		ids := make([]string, len(receipts))
		for i, r := range receipts {
			ids[i] = r.ID
		}
		if run == 0 {
			first = ids
			continue
		}
		if len(ids) != len(first) {
			t.Fatalf("run %d: count %d != %d", run, len(ids), len(first))
		}
		for i := range ids {
			if ids[i] != first[i] {
				t.Fatalf("run %d: receipt order/id changed at %d", run, i)
			}
		}
	}
}

var _ mail.MailSource = (*staticSource)(nil)
