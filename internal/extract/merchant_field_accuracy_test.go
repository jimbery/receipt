package extract_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jimbery/receipt/internal/classify"
	"github.com/jimbery/receipt/internal/emailtypes"
	"github.com/jimbery/receipt/internal/extract"
	"github.com/jimbery/receipt/internal/mail"
)

type merchantFieldLabel struct {
	merchant   string
	file       string
	kind       emailtypes.DocumentKind
	supplier   string
	totalMinor int64
	orderRef   string
}

func priorityMerchantLabels() []merchantFieldLabel {
	return []merchantFieldLabel{
		{"amazon", "receipt.txt", emailtypes.KindPurchaseReceipt, "Amazon", 2499, "AMZ-1001"},
		{"amazon", "partial-shipment.txt", emailtypes.KindUnknown, "Amazon", 1299, "AMZ-1001"},
		{"amazon", "credit-note.txt", emailtypes.KindCreditNote, "Amazon", -500, "AMZ-1001"},
		{"screwfix", "receipt.txt", emailtypes.KindPurchaseReceipt, "Screwfix", 4500, "SF2001"},
		{"screwfix", "partial.txt", emailtypes.KindPurchaseReceipt, "Screwfix", 4500, ""},
		{"screwfix", "refund.txt", emailtypes.KindCreditNote, "Screwfix", -1200, "SF2001"},
		{"toolstation", "receipt.txt", emailtypes.KindPurchaseReceipt, "Toolstation", 2250, "3001"},
		{"toolstation", "partial-shipment.txt", emailtypes.KindPurchaseReceipt, "Toolstation", 1500, "3001"},
		{"bandq", "receipt.txt", emailtypes.KindPurchaseReceipt, "B&Q", 6780, "4001"},
		{"bandq", "vat-exempt.txt", emailtypes.KindPurchaseReceipt, "B&Q", 1000, ""},
		{"bandq", "confirmation-4001.txt", emailtypes.KindPurchaseReceipt, "B&Q", 6780, "4001"},
		{"bp", "receipt.txt", emailtypes.KindPurchaseReceipt, "BP", 5500, ""},
		{"bp", "app-receipt.txt", emailtypes.KindPurchaseReceipt, "BP", 4200, ""},
		{"shell", "receipt.txt", emailtypes.KindPurchaseReceipt, "Shell", 6200, ""},
		{"shell", "app-receipt.txt", emailtypes.KindPurchaseReceipt, "Shell", 4850, ""},
		{"esso", "app-receipt.txt", emailtypes.KindPurchaseReceipt, "Esso", 3875, ""},
		{"edf", "invoice.txt", emailtypes.KindPurchaseReceipt, "EDF", 8900, "EDF7001"},
		{"edf", "standing-charge.txt", emailtypes.KindPurchaseReceipt, "EDF", 7250, "EDF7002"},
		{"octopus", "invoice.txt", emailtypes.KindPurchaseReceipt, "Octopus Energy", 12000, "ACC6001"},
	}
}

func TestMerchantFieldAccuracy_OnDiskFixtures(t *testing.T) {
	root := filepath.Join("..", "..", "test", "testdata", "email", "merchants")
	clf := classify.New()
	reg := extract.NewRegistry()

	var checks, correct int
	for _, label := range priorityMerchantLabels() {
		path := filepath.Join(root, label.merchant, label.file)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("%s/%s: %v", label.merchant, label.file, err)
		}
		msg, err := mail.ParseRFC822(data, label.file, "field-accuracy")
		if err != nil {
			t.Fatalf("%s/%s: parse: %v", label.merchant, label.file, err)
		}
		kind := clf.Classify(msg)
		if kind != label.kind {
			t.Errorf("%s/%s: kind %s want %s", label.merchant, label.file, kind, label.kind)
		}
		ex, ok := reg.Extract(msg, kind)
		if !ok {
			t.Fatalf("%s/%s: extract failed", label.merchant, label.file)
		}
		if ex.Supplier != label.supplier {
			t.Errorf("%s/%s: supplier %q want %q", label.merchant, label.file, ex.Supplier, label.supplier)
		} else {
			checks++
			correct++
		}
		if label.totalMinor != 0 {
			checks++
			if ex.TotalMinor != label.totalMinor {
				t.Errorf("%s/%s: total %d want %d", label.merchant, label.file, ex.TotalMinor, label.totalMinor)
			} else {
				correct++
			}
		}
		if label.orderRef != "" {
			checks++
			if ex.OrderRef != label.orderRef {
				t.Errorf("%s/%s: order ref %q want %q", label.merchant, label.file, ex.OrderRef, label.orderRef)
			} else {
				correct++
			}
		}
	}
	accuracy := float64(correct) / float64(checks)
	if accuracy < 0.98 {
		t.Fatalf("field accuracy %.3f below gate 0.98 (%d/%d)", accuracy, correct, checks)
	}
}

type emlFieldLabel struct {
	merchant   string
	file       string
	supplier   string
	totalMinor int64
	orderRef   string // empty skips check (scrubbed refs)
}

func scrubbedEMLLabels() []emlFieldLabel {
	return []emlFieldLabel{
		{"amazon", "ordered-2-amazon-fire-tv-stick-4k.eml", "Amazon", 4998, ""},
		{"amazon", "ordered-american-tourister-soundbox-and-2-more-items.eml", "Amazon", 11819, ""},
		{"amazon", "ordered-usb-c-plug-anker-40w-usb-c-and-1-more-item.eml", "Amazon", 3497, ""},
		{"screwfix", "confirmation-of-your-order-order.eml", "Screwfix", 999, ""},
		{"screwfix", "confirmation-of-your-order-order-2.eml", "Screwfix", 169, ""},
		{"toolstation", "toolstation-order-order-confirmed.eml", "Toolstation", 528, "REDACTED_ORDER"},
		{"toolstation", "toolstation-order-order-confirmed-2.eml", "Toolstation", 1687, "REDACTED_ORDER"},
		{"toolstation", "toolstation-order-order-confirmed-3.eml", "Toolstation", 998, "REDACTED_ORDER"},
	}
}

func TestMerchantEMLFixtures_FieldAccuracy(t *testing.T) {
	root := filepath.Join("..", "..", "test", "testdata", "email", "merchants")
	clf := classify.New()
	reg := extract.NewRegistry()

	var checks, correct int
	for _, label := range scrubbedEMLLabels() {
		path := filepath.Join(root, label.merchant, label.file)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("%s/%s: %v", label.merchant, label.file, err)
		}
		msg, err := mail.ParseRFC822(data, label.file, "eml-fixture")
		if err != nil {
			t.Fatalf("%s/%s: parse: %v", label.merchant, label.file, err)
		}
		kind := clf.Classify(msg)
		if kind != emailtypes.KindPurchaseReceipt {
			t.Fatalf("%s/%s: kind %s want purchase_receipt", label.merchant, label.file, kind)
		}
		ex, ok := reg.Extract(msg, kind)
		if !ok {
			t.Fatalf("%s/%s: extract failed", label.merchant, label.file)
		}
		checks++
		if ex.Supplier == label.supplier {
			correct++
		} else {
			t.Errorf("%s/%s: supplier %q want %q", label.merchant, label.file, ex.Supplier, label.supplier)
		}
		checks++
		if ex.TotalMinor == label.totalMinor {
			correct++
		} else {
			t.Errorf("%s/%s: total %d want %d", label.merchant, label.file, ex.TotalMinor, label.totalMinor)
		}
		if label.orderRef != "" {
			checks++
			if ex.OrderRef == label.orderRef {
				correct++
			} else {
				t.Errorf("%s/%s: order_ref %q want %q", label.merchant, label.file, ex.OrderRef, label.orderRef)
			}
		}
	}
	accuracy := float64(correct) / float64(checks)
	if accuracy < 0.98 {
		t.Fatalf("eml field accuracy %.3f below gate 0.98 (%d/%d)", accuracy, correct, checks)
	}
}
