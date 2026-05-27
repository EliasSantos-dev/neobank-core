package api

import (
	"errors"
	"net/http"

	"github.com/EliasSantos-dev/neobank-core/internal/ledger"
)

// statusFor mapeia erros de domínio para status HTTP.
func statusFor(err error) int {
	switch {
	case errors.Is(err, ledger.ErrAccountNotFound):
		return http.StatusNotFound
	case errors.Is(err, ledger.ErrInsufficientFunds),
		errors.Is(err, ledger.ErrCurrencyMismatch),
		errors.Is(err, ledger.ErrSameAccount):
		return http.StatusUnprocessableEntity
	case errors.Is(err, ledger.ErrInvalidAmount):
		return http.StatusBadRequest
	case errors.Is(err, ledger.ErrIdempotencyConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
