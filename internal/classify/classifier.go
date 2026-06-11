package classify

import (
	"strings"

	"github.com/jimbery/receipt/internal/emailtypes"
)

// Classifier assigns DocumentKind using subject, body, and MIME signals only.
// Merchant identity is intentionally deferred to the extractor registry (ADR-002 D1/D2).
type Classifier struct {
	rulesetHash string
}

// New returns a rule-based classifier with a stable ruleset identifier.
func New() *Classifier {
	return &Classifier{rulesetHash: "classify-v2-merchant-agnostic"}
}

// RulesetHash identifies the committed rule set for harness reporting.
func (c *Classifier) RulesetHash() string { return c.rulesetHash }

// Classify determines the document kind for a raw message.
func (c *Classifier) Classify(msg emailtypes.RawMessage) emailtypes.DocumentKind {
	from := strings.ToLower(msg.From)
	subject := strings.ToLower(msg.Subject)
	body := strings.ToLower(msg.BodyText + " " + msg.BodyHTML)

	if isMarketing(from, subject, body) {
		return emailtypes.KindMarketing
	}
	if isStatement(from, subject) {
		return emailtypes.KindStatement
	}
	if isCreditNote(subject, body) {
		return emailtypes.KindCreditNote
	}
	if isDispatchOnly(subject, body) {
		return emailtypes.KindDispatchNotice
	}
	if hasPurchaseSignals(subject, body) || hasReceiptAttachment(msg) {
		return emailtypes.KindPurchaseReceipt
	}
	return emailtypes.KindUnknown
}

func isMarketing(from, subject, body string) bool {
	markers := []string{
		"unsubscribe", "sale ends", "limited time offer", "newsletter",
		"your receipt awaits", "complete your purchase",
	}
	for _, m := range markers {
		if strings.Contains(subject, m) || strings.Contains(body, m) {
			return true
		}
	}
	if strings.Contains(from, "marketing@") || strings.Contains(from, "news@") {
		return true
	}
	return false
}

func isStatement(from, subject string) bool {
	if strings.Contains(subject, "monthly statement") || strings.Contains(subject, "account summary") {
		return true
	}
	return strings.Contains(from, "statements@")
}

func isCreditNote(subject, body string) bool {
	markers := []string{"credit note", "refund confirmation", "return processed"}
	for _, m := range markers {
		if strings.Contains(subject, m) || strings.Contains(body, m) {
			return true
		}
	}
	if strings.Contains(subject, "money back") || strings.Contains(body, "money back") {
		return !strings.Contains(body, "money back guarantee")
	}
	return false
}

func isDispatchOnly(subject, body string) bool {
	if strings.Contains(subject, "dispatched") || strings.Contains(subject, "shipped") {
		hasTotal := strings.Contains(body, "order total") ||
			strings.Contains(body, "total paid") ||
			(strings.Contains(body, "total") && strings.Contains(body, "£"))
		return !hasTotal
	}
	return false
}

func hasPurchaseSignals(subject, body string) bool {
	receiptWords := []string{
		"receipt", "invoice", "order confirmation", "your order", "payment received",
	}
	for _, w := range receiptWords {
		if strings.Contains(subject, w) || strings.Contains(body, w) {
			return true
		}
	}
	return strings.Contains(body, "order total") || strings.Contains(body, "total paid")
}

func hasReceiptAttachment(msg emailtypes.RawMessage) bool {
	for _, a := range msg.Attachments {
		name := strings.ToLower(a.Filename)
		if strings.Contains(name, "invoice") || strings.Contains(name, "receipt") {
			return true
		}
	}
	return false
}
