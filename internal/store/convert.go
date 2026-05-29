package store

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/EliasSantos-dev/neobank-core/internal/fx"
	"github.com/EliasSantos-dev/neobank-core/internal/ledger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type ConvertParams struct {
	IdempotencyKey string
	UserID         uuid.UUID
	From           string
	To             string
	Amount         int64
	RateE8         int64
}

type ConvertResult struct {
	Converted int64
	RateE8    int64
}

// ConvertCurrency converte Amount de From->To entre as wallets do usuário,
// atômico: trava as 4 contas, posta as duas pernas mono-moeda e emite o evento.
func (s *Store) ConvertCurrency(ctx context.Context, p ConvertParams) (ConvertResult, error) {
	if p.From == p.To {
		return ConvertResult{}, ledger.ErrSameAccount
	}
	if p.Amount <= 0 {
		return ConvertResult{}, ledger.ErrInvalidAmount
	}
	fxFrom, ok1 := ledger.FxAccount(p.From)
	fxTo, ok2 := ledger.FxAccount(p.To)
	if !ok1 || !ok2 {
		return ConvertResult{}, ledger.ErrCurrencyMismatch
	}
	walletFrom, err := s.GetOrCreateWallet(ctx, p.UserID, p.From)
	if err != nil {
		return ConvertResult{}, err
	}
	walletTo, err := s.GetOrCreateWallet(ctx, p.UserID, p.To)
	if err != nil {
		return ConvertResult{}, err
	}
	converted := fx.Convert(p.Amount, p.RateE8)
	baseKey := "conv:" + p.IdempotencyKey

	// idempotência: se a perna base já existe, devolve o resultado anterior.
	if _, err := getTransferByKey(ctx, s.pool, baseKey); err == nil {
		return ConvertResult{Converted: converted, RateE8: p.RateE8}, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return ConvertResult{}, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return ConvertResult{}, err
	}
	defer tx.Rollback(ctx)

	// trava as 4 contas distintas em ordem determinística (sem deadlock).
	accs, err := lockAccounts4(ctx, tx, walletFrom, walletTo, fxFrom, fxTo)
	if err != nil {
		return ConvertResult{}, err
	}

	bal, err := balance(ctx, tx, walletFrom)
	if err != nil {
		return ConvertResult{}, err
	}
	if bal < p.Amount {
		return ConvertResult{}, ledger.ErrInsufficientFunds
	}

	// perna base: walletFrom -> fxFrom (moeda From)
	trBase, err := createTransferTx(ctx, tx, baseKey)
	if err != nil {
		return ConvertResult{}, err
	}
	if err := postLegs(ctx, tx, trBase, accs[walletFrom], accs[fxFrom], p.Amount); err != nil {
		return ConvertResult{}, err
	}
	// perna quote: fxTo -> walletTo (moeda To)
	trQuote, err := createTransferTx(ctx, tx, "conv-quote:"+p.IdempotencyKey)
	if err != nil {
		return ConvertResult{}, err
	}
	if err := postLegs(ctx, tx, trQuote, accs[fxTo], accs[walletTo], converted); err != nil {
		return ConvertResult{}, err
	}

	payload, _ := json.Marshal(map[string]any{
		"user_id": p.UserID, "from": p.From, "to": p.To, "amount": p.Amount, "converted": converted, "rate_e8": p.RateE8,
	})
	if _, err := writeEventTx(ctx, tx, "currency.converted", payload); err != nil {
		return ConvertResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return ConvertResult{}, err
	}
	return ConvertResult{Converted: converted, RateE8: p.RateE8}, nil
}

func createTransferTx(ctx context.Context, tx pgx.Tx, key string) (uuid.UUID, error) {
	var id uuid.UUID
	err := tx.QueryRow(ctx,
		`INSERT INTO transfers (idempotency_key) VALUES ($1) RETURNING id`, key).Scan(&id)
	return id, err
}

// lockAccounts4 trava as contas distintas ORDER BY id FOR UPDATE e devolve id->Account.
func lockAccounts4(ctx context.Context, tx pgx.Tx, ids ...uuid.UUID) (map[uuid.UUID]ledger.Account, error) {
	rows, err := tx.Query(ctx,
		`SELECT id, currency, type, created_at FROM accounts WHERE id = ANY($1) ORDER BY id FOR UPDATE`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[uuid.UUID]ledger.Account, len(ids))
	for rows.Next() {
		var a ledger.Account
		var typ string
		if err := rows.Scan(&a.ID, &a.Currency, &typ, &a.CreatedAt); err != nil {
			return nil, err
		}
		a.Type = ledger.AccountType(typ)
		out[a.ID] = a
	}
	return out, rows.Err()
}
