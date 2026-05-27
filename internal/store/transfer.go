package store

import (
	"context"
	"errors"

	"github.com/EliasSantos-dev/neobank-core/internal/ledger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type TransferParams struct {
	IdempotencyKey string
	FromAccountID  uuid.UUID
	ToAccountID    uuid.UUID
	Amount         int64
	Currency       string
}

// Transfer move `Amount` de FromAccountID para ToAccountID de forma atômica,
// idempotente e concorrente-safe (lock de âncora ordenado nas contas).
func (s *Store) Transfer(ctx context.Context, p TransferParams) (ledger.Transfer, error) {
	if p.Amount <= 0 {
		return ledger.Transfer{}, ledger.ErrInvalidAmount
	}
	if p.FromAccountID == p.ToAccountID {
		return ledger.Transfer{}, ledger.ErrSameAccount
	}

	// Fast-path de idempotência (fora da transação).
	if tr, err := getTransferByKey(ctx, s.pool, p.IdempotencyKey); err == nil {
		return tr, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return ledger.Transfer{}, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ledger.Transfer{}, err
	}
	defer tx.Rollback(ctx)

	// 1. Lock de âncora ordenado (ORDER BY id) -> sem deadlock, sem race.
	accs, err := lockAccounts(ctx, tx, p.FromAccountID, p.ToAccountID)
	if err != nil {
		return ledger.Transfer{}, err
	}
	from, okFrom := accs[p.FromAccountID]
	to, okTo := accs[p.ToAccountID]
	if !okFrom || !okTo {
		return ledger.Transfer{}, ledger.ErrAccountNotFound
	}

	// 2. Validação de moeda.
	if from.Currency != p.Currency || to.Currency != p.Currency {
		return ledger.Transfer{}, ledger.ErrCurrencyMismatch
	}

	// 3. Saldo derivado SOB LOCK -> checagem de fundos confiável.
	if !from.Type.AllowsNegative() {
		bal, err := balance(ctx, tx, from.ID)
		if err != nil {
			return ledger.Transfer{}, err
		}
		if bal < p.Amount {
			return ledger.Transfer{}, ledger.ErrInsufficientFunds
		}
	}

	// 4. Cria o transfer (UNIQUE(idempotency_key) protege duplicata concorrente).
	var tr ledger.Transfer
	err = tx.QueryRow(ctx,
		`INSERT INTO transfers (idempotency_key) VALUES ($1)
		 RETURNING id, idempotency_key, status, created_at`,
		p.IdempotencyKey).Scan(&tr.ID, &tr.IdempotencyKey, &tr.Status, &tr.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			_ = tx.Rollback(ctx)
			return getTransferByKey(ctx, s.pool, p.IdempotencyKey)
		}
		return ledger.Transfer{}, err
	}

	// 5. Lança as duas pernas balanceadas (débito na origem, crédito no destino).
	entries, err := ledger.BuildEntries(tr.ID, from, to, p.Amount)
	if err != nil {
		return ledger.Transfer{}, err
	}
	for _, e := range entries {
		if _, err := tx.Exec(ctx,
			`INSERT INTO entries (transfer_id, account_id, direction, amount, currency)
			 VALUES ($1, $2, $3, $4, $5)`,
			e.TransferID, e.AccountID, string(e.Direction), e.Amount, e.Currency); err != nil {
			return ledger.Transfer{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return ledger.Transfer{}, err
	}
	return tr, nil
}

func lockAccounts(ctx context.Context, q querier, a, b uuid.UUID) (map[uuid.UUID]ledger.Account, error) {
	rows, err := q.Query(ctx,
		`SELECT id, currency, type, created_at FROM accounts
		 WHERE id IN ($1, $2) ORDER BY id FOR UPDATE`, a, b)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[uuid.UUID]ledger.Account, 2)
	for rows.Next() {
		var acc ledger.Account
		var typ string
		if err := rows.Scan(&acc.ID, &acc.Currency, &typ, &acc.CreatedAt); err != nil {
			return nil, err
		}
		acc.Type = ledger.AccountType(typ)
		out[acc.ID] = acc
	}
	return out, rows.Err()
}
