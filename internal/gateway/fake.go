package gateway

import "context"

// FakeApproves decide de forma determinística se o provedor aprova um valor.
// Recusa quando amount % 100 == 13 (análogo a cartões de teste).
func FakeApproves(amount int64) bool { return amount%100 != 13 }

type FakeProvider struct{}

func NewFakeProvider() *FakeProvider { return &FakeProvider{} }

func (FakeProvider) CreateCharge(_ context.Context, req ChargeRequest) (ProviderRef, error) {
	return ProviderRef("charge_" + req.PaymentID.String()), nil
}

func (FakeProvider) CreatePayout(_ context.Context, req PayoutRequest) (ProviderRef, error) {
	return ProviderRef("payout_" + req.PaymentID.String()), nil
}
