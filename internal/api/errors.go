package api

import (
	"errors"
	"net/http"

	"github.com/EliasSantos-dev/neobank-core/internal/fx"
	"github.com/EliasSantos-dev/neobank-core/internal/ledger"
	"github.com/EliasSantos-dev/neobank-core/internal/store"
	"github.com/EliasSantos-dev/neobank-core/internal/user"
)

// statusFor mapeia erros de domínio para status HTTP.
func statusFor(err error) int {
	switch {
	case errors.Is(err, ledger.ErrAccountNotFound), errors.Is(err, user.ErrUserNotFound),
		errors.Is(err, store.ErrIntentNotFound), errors.Is(err, store.ErrPaymentNotFound):
		return http.StatusNotFound
	case errors.Is(err, store.ErrNotUnderReview):
		return http.StatusConflict
	case errors.Is(err, ledger.ErrInsufficientFunds),
		errors.Is(err, ledger.ErrCurrencyMismatch),
		errors.Is(err, ledger.ErrSameAccount):
		return http.StatusUnprocessableEntity
	case errors.Is(err, ledger.ErrInvalidAmount),
		errors.Is(err, user.ErrInvalidEmail),
		errors.Is(err, user.ErrWeakPassword):
		return http.StatusBadRequest
	case errors.Is(err, fx.ErrRateUnavailable):
		return http.StatusUnprocessableEntity
	case errors.Is(err, user.ErrEmailTaken):
		return http.StatusConflict
	case errors.Is(err, user.ErrInvalidCredentials):
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}
