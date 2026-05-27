// Package money é um value object para valores monetários em unidades mínimas.
package money

import (
	"errors"
	"regexp"
)

var (
	ErrNegativeAmount  = errors.New("money: amount must not be negative")
	ErrInvalidCurrency = errors.New("money: currency must be 3 uppercase letters")
)

var currencyRe = regexp.MustCompile(`^[A-Z]{3}$`)

// Money é um valor em unidades mínimas (ex.: centavos) de uma moeda.
type Money struct {
	Amount   int64  // >= 0
	Currency string // 3 letras maiúsculas
}

// New valida e cria um Money.
func New(amount int64, currency string) (Money, error) {
	if amount < 0 {
		return Money{}, ErrNegativeAmount
	}
	if !currencyRe.MatchString(currency) {
		return Money{}, ErrInvalidCurrency
	}
	return Money{Amount: amount, Currency: currency}, nil
}
