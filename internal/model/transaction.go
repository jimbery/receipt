package model

type TransactionSource string

const (
	TransactionSourceOpenBanking TransactionSource = "open_banking"
	TransactionSourceCardAuth    TransactionSource = "card_auth"
	TransactionSourceSynthetic   TransactionSource = "synthetic"
)

type Transaction struct {
	ID          string
	Source      TransactionSource
	Merchant    string
	MCC         string
	Amount      Money
	OccurredAt  Timestamp
	SettledAt   *Timestamp
	AccountID   string
	ExternalRef string
}
