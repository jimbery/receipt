package model

type ReceiptSource string

const (
	ReceiptSourceEmail ReceiptSource = "email"
	ReceiptSourceOCR   ReceiptSource = "ocr"
	ReceiptSourcePOS   ReceiptSource = "pos"
)

type LineItem struct {
	Description string
	Quantity    float64
	UnitPrice   Money
	NetAmount   Money
	VATAmount   Money
	VATRate     float64
}

type Receipt struct {
	ID              string
	Source          ReceiptSource
	Supplier        string
	SupplierVATNo   string
	LineItems       []LineItem
	Total           Money
	Category        string
	IssuedAt        Timestamp
	ExternalRef     string
	TransactionRef  string // POS exact reference for short-circuit matching
	RawMerchantHint string
}

func (r Receipt) Itemised() bool {
	return len(r.LineItems) > 0
}
