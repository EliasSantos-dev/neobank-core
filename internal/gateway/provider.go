// Package gateway integra com um provedor de pagamento externo (cash-in/out).
package gateway

import (
	"context"

	"github.com/google/uuid"
)

type ProviderRef string

type ChargeRequest struct {
	PaymentID uuid.UUID
	Amount    int64
	Currency  string
}

type PayoutRequest struct {
	PaymentID uuid.UUID
	Amount    int64
	Currency  string
}

// PaymentProvider abstrai o provedor externo. FakeProvider implementa para
// dev/teste; um HTTPProvider (Stripe/AbacatePay) plugaria a mesma interface.
type PaymentProvider interface {
	CreateCharge(ctx context.Context, req ChargeRequest) (ProviderRef, error)
	CreatePayout(ctx context.Context, req PayoutRequest) (ProviderRef, error)
}
