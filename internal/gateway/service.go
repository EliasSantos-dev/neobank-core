package gateway

import (
	"context"
	"encoding/json"

	"github.com/EliasSantos-dev/neobank-core/internal/store"
	"github.com/google/uuid"
)

type Service struct {
	s        *store.Store
	provider PaymentProvider
}

func NewService(s *store.Store, provider PaymentProvider) *Service {
	return &Service{s: s, provider: provider}
}

func (svc *Service) InitiateDeposit(ctx context.Context, accountID uuid.UUID, amount int64, idemKey string) (store.Payment, error) {
	p, err := svc.s.CreatePayment(ctx, store.PaymentParams{
		AccountID: accountID, Kind: "deposit", Amount: amount, Currency: "BRL", IdempotencyKey: idemKey,
	})
	if err != nil {
		return store.Payment{}, err
	}
	ref, err := svc.provider.CreateCharge(ctx, ChargeRequest{PaymentID: p.ID, Amount: amount, Currency: "BRL"})
	if err != nil {
		return store.Payment{}, err
	}
	_ = svc.s.SetPaymentProviderRef(ctx, p.ID, string(ref))
	return p, nil
}

func (svc *Service) InitiateWithdrawal(ctx context.Context, accountID uuid.UUID, amount int64, idemKey string) (store.Payment, error) {
	p, err := svc.s.CreatePayment(ctx, store.PaymentParams{
		AccountID: accountID, Kind: "withdrawal", Amount: amount, Currency: "BRL", IdempotencyKey: idemKey,
	})
	if err != nil {
		return store.Payment{}, err
	}
	// hold (reserva) — falha (ex.: saldo insuficiente) marca o payment como failed.
	if err := svc.s.HoldWithdrawal(ctx, p.ID, accountID, amount); err != nil {
		_ = svc.s.MarkPaymentStatus(ctx, p.ID, "failed")
		return store.Payment{}, err
	}
	ref, err := svc.provider.CreatePayout(ctx, PayoutRequest{PaymentID: p.ID, Amount: amount, Currency: "BRL"})
	if err != nil {
		return store.Payment{}, err
	}
	_ = svc.s.SetPaymentProviderRef(ctx, p.ID, string(ref))
	return p, nil
}

// HandleWebhook processa a confirmação do provedor de forma idempotente.
func (svc *Service) HandleWebhook(ctx context.Context, evt WebhookEvent) error {
	first, err := svc.s.RecordWebhookEvent(ctx, evt.EventID, evt.PaymentID)
	if err != nil {
		return err
	}
	if !first {
		return nil // já processado
	}
	p, err := svc.s.GetPayment(ctx, evt.PaymentID)
	if err != nil {
		return err
	}
	if p.Status != "pending" {
		return nil // guarda de estado terminal
	}

	switch {
	case p.Kind == "deposit" && evt.Status == "succeeded":
		if err := svc.s.CreditDeposit(ctx, p.ID, p.AccountID, p.Amount); err != nil {
			return err
		}
		return svc.finalize(ctx, p, "completed")
	case p.Kind == "deposit" && evt.Status == "failed":
		return svc.finalize(ctx, p, "failed")
	case p.Kind == "withdrawal" && evt.Status == "succeeded":
		return svc.finalize(ctx, p, "completed")
	case p.Kind == "withdrawal" && evt.Status == "failed":
		if err := svc.s.RefundWithdrawal(ctx, p.ID, p.AccountID, p.Amount); err != nil {
			return err
		}
		return svc.finalize(ctx, p, "failed")
	}
	return nil
}

// finalize marca o status e emite o evento payment.* no outbox.
func (svc *Service) finalize(ctx context.Context, p store.Payment, status string) error {
	if err := svc.s.MarkPaymentStatus(ctx, p.ID, status); err != nil {
		return err
	}
	topic := "payment.completed"
	if status == "failed" {
		topic = "payment.failed"
	}
	payload, _ := json.Marshal(map[string]any{
		"payment_id": p.ID,
		"kind":       p.Kind,
		"amount":     p.Amount,
	})
	_, err := svc.s.WriteEventDirect(ctx, topic, payload)
	return err
}
