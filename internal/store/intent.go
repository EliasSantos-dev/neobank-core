package store

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrIntentNotFound = errors.New("store: intenção não encontrada")
	ErrNotUnderReview = errors.New("store: intenção não está em revisão")
)

type Intent struct {
	ID             uuid.UUID
	IdempotencyKey string
	FromAccountID  uuid.UUID
	ToAccountID    uuid.UUID
	Amount         int64
	Currency       string
	Status         string
	RiskScore      int
	RiskLevel      string
	RiskReasons    []byte
	TransferID     *uuid.UUID
}

type IntentParams struct {
	IdempotencyKey string
	FromAccountID  uuid.UUID
	ToAccountID    uuid.UUID
	Amount         int64
	Currency       string
}

const intentCols = `id, idempotency_key, from_account_id, to_account_id, amount, currency, status, risk_score, risk_level, risk_reasons, transfer_id`

func scanIntent(row pgx.Row) (Intent, error) {
	var it Intent
	var score *int
	var level *string
	err := row.Scan(&it.ID, &it.IdempotencyKey, &it.FromAccountID, &it.ToAccountID,
		&it.Amount, &it.Currency, &it.Status, &score, &level, &it.RiskReasons, &it.TransferID)
	if score != nil {
		it.RiskScore = *score
	}
	if level != nil {
		it.RiskLevel = *level
	}
	return it, err
}

func (s *Store) CreateIntent(ctx context.Context, p IntentParams) (Intent, error) {
	var it Intent
	err := s.pool.QueryRow(ctx,
		`INSERT INTO transfer_intents (idempotency_key, from_account_id, to_account_id, amount, currency)
		 VALUES ($1,$2,$3,$4,$5)
		 RETURNING id, idempotency_key, from_account_id, to_account_id, amount, currency, status`,
		p.IdempotencyKey, p.FromAccountID, p.ToAccountID, p.Amount, p.Currency).
		Scan(&it.ID, &it.IdempotencyKey, &it.FromAccountID, &it.ToAccountID, &it.Amount, &it.Currency, &it.Status)
	return it, err
}

func (s *Store) GetIntent(ctx context.Context, id uuid.UUID) (Intent, error) {
	it, err := scanIntent(s.pool.QueryRow(ctx, `SELECT `+intentCols+` FROM transfer_intents WHERE id=$1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Intent{}, ErrIntentNotFound
	}
	return it, err
}

// ListPending devolve intenções pendentes (mais antigas primeiro).
func (s *Store) ListPending(ctx context.Context, limit int) ([]Intent, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT `+intentCols+` FROM transfer_intents WHERE status='pending' ORDER BY created_at LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Intent
	for rows.Next() {
		it, err := scanIntent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (s *Store) ListIntentsByAccount(ctx context.Context, accountID uuid.UUID, limit, offset int32) ([]Intent, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT `+intentCols+` FROM transfer_intents WHERE from_account_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		accountID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Intent
	for rows.Next() {
		it, err := scanIntent(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

// SetRisk grava o score/level/reasons sem mudar o status (observabilidade,
// usado também quando o risco é baixo e a intenção é auto-aprovada).
func (s *Store) SetRisk(ctx context.Context, id uuid.UUID, score int, level string, reasons []byte) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE transfer_intents SET risk_score=$2, risk_level=$3, risk_reasons=$4, updated_at=now() WHERE id=$1`,
		id, score, level, reasons)
	return err
}

func (s *Store) MarkUnderReview(ctx context.Context, id uuid.UUID, score int, level string, reasons []byte) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE transfer_intents SET status='under_review', risk_score=$2, risk_level=$3, risk_reasons=$4, updated_at=now()
		 WHERE id=$1`, id, score, level, reasons)
	return err
}

// CompleteIntent efetiva a intenção no ledger (idempotente) e marca como completed.
// Usa a id da intenção como Idempotency-Key do ledger, então repetir é seguro.
func (s *Store) CompleteIntent(ctx context.Context, id uuid.UUID) error {
	it, err := s.GetIntent(ctx, id)
	if err != nil {
		return err
	}
	if it.Status == "completed" {
		return nil // idempotente
	}
	tr, err := s.Transfer(ctx, TransferParams{
		IdempotencyKey: "intent:" + it.ID.String(),
		FromAccountID:  it.FromAccountID,
		ToAccountID:    it.ToAccountID,
		Amount:         it.Amount,
		Currency:       it.Currency,
	})
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx,
		`UPDATE transfer_intents SET status='completed', transfer_id=$2, updated_at=now() WHERE id=$1`,
		it.ID, tr.ID)
	return err
}

func (s *Store) RejectIntent(ctx context.Context, id uuid.UUID) error {
	ct, err := s.pool.Exec(ctx,
		`UPDATE transfer_intents SET status='rejected', updated_at=now() WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrIntentNotFound
	}
	return nil
}

// RecipientIsNew indica se nunca houve transferência efetivada de from para to.
func (s *Store) RecipientIsNew(ctx context.Context, from, to uuid.UUID) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx,
		`SELECT EXISTS (
		   SELECT 1 FROM entries d
		   JOIN entries c ON c.transfer_id = d.transfer_id
		   WHERE d.account_id = $1 AND d.direction = 'debit'
		     AND c.account_id = $2 AND c.direction = 'credit'
		 )`, from, to).Scan(&exists)
	return !exists, err
}

// DebitHistory devolve os valores de débitos passados da conta (para z-score).
func (s *Store) DebitHistory(ctx context.Context, accountID uuid.UUID, limit int) ([]int64, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT amount FROM entries WHERE account_id=$1 AND direction='debit' ORDER BY seq DESC LIMIT $2`,
		accountID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var a int64
		if err := rows.Scan(&a); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// RecentDebitCount conta débitos recentes da conta (janela simples por contagem).
func (s *Store) RecentDebitCount(ctx context.Context, accountID uuid.UUID, window int) (int, error) {
	var n int
	err := s.pool.QueryRow(ctx,
		`SELECT count(*) FROM (
		   SELECT 1 FROM entries WHERE account_id=$1 AND direction='debit' ORDER BY seq DESC LIMIT $2
		 ) t`, accountID, window).Scan(&n)
	return n, err
}
