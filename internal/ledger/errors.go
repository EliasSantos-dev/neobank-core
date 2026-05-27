package ledger

import "errors"

var (
	ErrAccountNotFound     = errors.New("ledger: account not found")
	ErrInsufficientFunds   = errors.New("ledger: insufficient funds")
	ErrCurrencyMismatch    = errors.New("ledger: currency mismatch")
	ErrIdempotencyConflict = errors.New("ledger: idempotency key reused with different payload")
	ErrInvalidAmount       = errors.New("ledger: amount must be positive")
	ErrSameAccount         = errors.New("ledger: from and to accounts must differ")
)
