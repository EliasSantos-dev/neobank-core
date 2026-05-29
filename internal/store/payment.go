package store

import (
	"context"
	"errors"

	"github.com/EliasSantos-dev/neobank-core/internal/ledger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrPaymentNotFound = errors.New("store: pagamento não encontrado")

type Payment struct {
	ID             uuid.UUID
	AccountID      uuid.UUID
	Kind           string
	Amount         int64
	Currency       string
	Status         string
	ProviderRef    string
	IdempotencyKey string
}

type PaymentParams struct {
	AccountID      uuid.UUID
	Kind           string
	Amount         int64
	Currency       string
	IdempotencyKey string
}

func (s *Store) CreatePayment(ctx context.Context, p PaymentParams) (Payment, error) {
	var out Payment
	err := s.pool.QueryRow(ctx,
		`INSERT INTO payments (account_id, kind, amount, currency, idempotency_key)
		 VALUES ($1,$2,$3,$4,$5)
		 RETURNING id, account_id, kind, amount, currency, status, COALESCE(provider_ref,''), idempotency_key`,
		p.AccountID, p.Kind, p.Amount, p.Currency, p.IdempotencyKey).
		Scan(&out.ID, &out.AccountID, &out.Kind, &out.Amount, &out.Currency, &out.Status, &out.ProviderRef, &out.IdempotencyKey)
	return out, err
}

func (s *Store) GetPayment(ctx context.Context, id uuid.UUID) (Payment, error) {
	var p Payment
	err := s.pool.QueryRow(ctx,
		`SELECT id, account_id, kind, amount, currency, status, COALESCE(provider_ref,''), idempotency_key
		 FROM payments WHERE id=$1`, id).
		Scan(&p.ID, &p.AccountID, &p.Kind, &p.Amount, &p.Currency, &p.Status, &p.ProviderRef, &p.IdempotencyKey)
	if errors.Is(err, pgx.ErrNoRows) {
		return Payment{}, ErrPaymentNotFound
	}
	return p, err
}

func (s *Store) SetPaymentProviderRef(ctx context.Context, id uuid.UUID, ref string) error {
	_, err := s.pool.Exec(ctx, `UPDATE payments SET provider_ref=$2, updated_at=now() WHERE id=$1`, id, ref)
	return err
}

func (s *Store) MarkPaymentStatus(ctx context.Context, id uuid.UUID, status string) error {
	ct, err := s.pool.Exec(ctx, `UPDATE payments SET status=$2, updated_at=now() WHERE id=$1`, id, status)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrPaymentNotFound
	}
	return nil
}

func (s *Store) ListPaymentsByAccount(ctx context.Context, accountID uuid.UUID, limit, offset int32) ([]Payment, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, account_id, kind, amount, currency, status, COALESCE(provider_ref,''), idempotency_key
		 FROM payments WHERE account_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		accountID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Payment
	for rows.Next() {
		var p Payment
		if err := rows.Scan(&p.ID, &p.AccountID, &p.Kind, &p.Amount, &p.Currency, &p.Status, &p.ProviderRef, &p.IdempotencyKey); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// RecordWebhookEvent registra um event_id; retorna true se é a primeira vez
// (false = já processado → idempotência).
func (s *Store) RecordWebhookEvent(ctx context.Context, eventID, paymentID uuid.UUID) (bool, error) {
	ct, err := s.pool.Exec(ctx,
		`INSERT INTO webhook_events (event_id, payment_id) VALUES ($1,$2) ON CONFLICT (event_id) DO NOTHING`,
		eventID, paymentID)
	if err != nil {
		return false, err
	}
	return ct.RowsAffected() == 1, nil
}

// CreditDeposit credita a wallet a partir do gateway (depósito confirmado).
func (s *Store) CreditDeposit(ctx context.Context, paymentID, walletID uuid.UUID, amount int64) error {
	_, err := s.Transfer(ctx, TransferParams{
		IdempotencyKey: "deposit:" + paymentID.String(),
		FromAccountID:  ledger.GatewayBRL,
		ToAccountID:    walletID,
		Amount:         amount,
		Currency:       "BRL",
	})
	return err
}

// HoldWithdrawal reserva os fundos da wallet no gateway (saque solicitado).
func (s *Store) HoldWithdrawal(ctx context.Context, paymentID, walletID uuid.UUID, amount int64) error {
	_, err := s.Transfer(ctx, TransferParams{
		IdempotencyKey: "withdraw-hold:" + paymentID.String(),
		FromAccountID:  walletID,
		ToAccountID:    ledger.GatewayBRL,
		Amount:         amount,
		Currency:       "BRL",
	})
	return err
}

// RefundWithdrawal devolve os fundos à wallet (saque falhou no gateway).
func (s *Store) RefundWithdrawal(ctx context.Context, paymentID, walletID uuid.UUID, amount int64) error {
	_, err := s.Transfer(ctx, TransferParams{
		IdempotencyKey: "withdraw-refund:" + paymentID.String(),
		FromAccountID:  ledger.GatewayBRL,
		ToAccountID:    walletID,
		Amount:         amount,
		Currency:       "BRL",
	})
	return err
}
