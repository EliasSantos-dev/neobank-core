package store_test

import (
	"context"
	"testing"

	"github.com/EliasSantos-dev/neobank-core/internal/ledger"
	"github.com/EliasSantos-dev/neobank-core/internal/store"
	itest "github.com/EliasSantos-dev/neobank-core/test"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func fund(t *testing.T, s *store.Store, from, to uuid.UUID, amount int64, key string) {
	t.Helper()
	_, err := s.Transfer(context.Background(), store.TransferParams{
		IdempotencyKey: key, FromAccountID: from, ToAccountID: to,
		Amount: amount, Currency: "BRL",
	})
	require.NoError(t, err)
}

func TestCreateAccountAndBalance(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	ctx := context.Background()

	acc, err := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	require.NoError(t, err)
	require.Equal(t, "BRL", acc.Currency)
	require.Equal(t, ledger.Wallet, acc.Type)

	bal, err := s.Balance(ctx, acc.ID)
	require.NoError(t, err)
	require.EqualValues(t, 0, bal) // conta nova: saldo derivado = 0
}

func TestTransfer_HappyPath(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	ctx := context.Background()

	ext, _ := s.CreateAccount(ctx, "BRL", ledger.External)
	a, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	b, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	fund(t, s, ext.ID, a.ID, 10000, "fund-a")

	_, err := s.Transfer(ctx, store.TransferParams{
		IdempotencyKey: "t1", FromAccountID: a.ID, ToAccountID: b.ID,
		Amount: 3000, Currency: "BRL",
	})
	require.NoError(t, err)

	balA, _ := s.Balance(ctx, a.ID)
	balB, _ := s.Balance(ctx, b.ID)
	require.EqualValues(t, 7000, balA)
	require.EqualValues(t, 3000, balB)
}

func TestTransfer_InsufficientFunds(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	ctx := context.Background()

	a, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	b, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)

	_, err := s.Transfer(ctx, store.TransferParams{
		IdempotencyKey: "t2", FromAccountID: a.ID, ToAccountID: b.ID,
		Amount: 100, Currency: "BRL",
	})
	require.ErrorIs(t, err, ledger.ErrInsufficientFunds)
}

func TestTransfer_Idempotent(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	ctx := context.Background()

	ext, _ := s.CreateAccount(ctx, "BRL", ledger.External)
	a, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	b, _ := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	fund(t, s, ext.ID, a.ID, 10000, "fund-a")

	p := store.TransferParams{IdempotencyKey: "dup", FromAccountID: a.ID, ToAccountID: b.ID, Amount: 2000, Currency: "BRL"}
	t1, err := s.Transfer(ctx, p)
	require.NoError(t, err)
	t2, err := s.Transfer(ctx, p) // mesma chave
	require.NoError(t, err)
	require.Equal(t, t1.ID, t2.ID) // mesmo transfer, sem efeito novo

	balB, _ := s.Balance(ctx, b.ID)
	require.EqualValues(t, 2000, balB) // creditou só uma vez
}
