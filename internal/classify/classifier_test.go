package classify_test

import (
	"testing"

	"github.com/jimbery/receipt/internal/classify"
	"github.com/jimbery/receipt/internal/emailtypes"
)

func TestClassifier_DoesNotUseSenderDomainAlone(t *testing.T) {
	clf := classify.New()
	msg := emailtypes.RawMessage{
		From:     "orders@amazon.co.uk",
		Subject:  "Hello from Amazon",
		BodyText: "Tracking update only.",
	}
	got := clf.Classify(msg)
	if got == emailtypes.KindPurchaseReceipt {
		t.Fatalf("domain alone classified as purchase_receipt")
	}
}

func TestClassifier_HardNegatives(t *testing.T) {
	clf := classify.New()
	cases := []struct {
		subject string
		body    string
		from    string
		want    emailtypes.DocumentKind
	}{
		{"Your receipt awaits", "unsubscribe", "news@shop.com", emailtypes.KindMarketing},
		{"Monthly statement", "balance", "statements@bank.com", emailtypes.KindStatement},
		{"Refund confirmation", "credit note", "returns@shop.com", emailtypes.KindCreditNote},
		{"Your package dispatched", "tracking only", "ship@amazon.co.uk", emailtypes.KindDispatchNotice},
	}
	for _, tc := range cases {
		got := clf.Classify(emailtypes.RawMessage{From: tc.from, Subject: tc.subject, BodyText: tc.body})
		if got != tc.want {
			t.Fatalf("%q: got %s want %s", tc.subject, got, tc.want)
		}
	}
}
