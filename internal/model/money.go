package model

import (
	"fmt"
	"math"
)

const minorUnitsPerMajor = 100

// Money is an amount in minor units (pence, cents) with ISO currency code.
type Money struct {
	Amount   int64
	Currency string
}

func NewMoney(amount int64, currency string) Money {
	return Money{Amount: amount, Currency: currency}
}

func (m Money) Equal(other Money) bool {
	return m.Amount == other.Amount && m.Currency == other.Currency
}

func (m Money) AbsDiff(other Money) (int64, bool) {
	if m.Currency != other.Currency {
		return 0, false
	}
	diff := m.Amount - other.Amount
	if diff < 0 {
		diff = -diff
	}
	return diff, true
}

func (m Money) WithinTolerance(other Money, absTol int64, relTol float64) bool {
	diff, ok := m.AbsDiff(other)
	if !ok {
		return false
	}
	if diff <= absTol {
		return true
	}
	larger := max(m.Amount, other.Amount)
	if larger == 0 {
		return diff == 0
	}
	return float64(diff)/float64(larger) <= relTol
}

func (m Money) String() string {
	return fmt.Sprintf("%s %.2f", m.Currency, float64(m.Amount)/minorUnitsPerMajor)
}

func RoundMinor(major float64) int64 {
	return int64(math.Round(major * minorUnitsPerMajor))
}
