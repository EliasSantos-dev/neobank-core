package store_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/EliasSantos-dev/neobank-core/internal/ledger"
	"github.com/EliasSantos-dev/neobank-core/internal/store"
	itest "github.com/EliasSantos-dev/neobank-core/test"
	"github.com/stretchr/testify/require"
)

// TestTransfer_Concurrent_NoDoubleSpend dispara muitas transferências em paralelo
// na mesma conta de origem e prova: nenhum double-spend, saldo nunca negativo,
// e conservação do dinheiro (Σ saldos = 0).
func TestTransfer_Concurrent_NoDoubleSpend(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	ctx := context.Background()

	ext, err := s.CreateAccount(ctx, "BRL", ledger.External)
	require.NoError(t, err)
	a, err := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	require.NoError(t, err)
	b, err := s.CreateAccount(ctx, "BRL", ledger.Wallet)
	require.NoError(t, err)

	// A recebe 100 (cabem exatamente 10 transferências de 10).
	_, err = s.Transfer(ctx, store.TransferParams{
		IdempotencyKey: "fund", FromAccountID: ext.ID, ToAccountID: a.ID,
		Amount: 100, Currency: "BRL",
	})
	require.NoError(t, err)

	const n = 50
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := s.Transfer(ctx, store.TransferParams{
				IdempotencyKey: fmt.Sprintf("c-%d", i),
				FromAccountID:  a.ID, ToAccountID: b.ID,
				Amount: 10, Currency: "BRL",
			})
			errs <- err
		}(i)
	}
	wg.Wait()
	close(errs)

	success := 0
	for err := range errs {
		if err == nil {
			success++
		} else {
			require.ErrorIs(t, err, ledger.ErrInsufficientFunds)
		}
	}

	balA, _ := s.Balance(ctx, a.ID)
	balB, _ := s.Balance(ctx, b.ID)
	balExt, _ := s.Balance(ctx, ext.ID)

	require.Equal(t, 10, success, "exatamente 10 transferências de 10 cabem em 100")
	require.EqualValues(t, 0, balA, "A nunca fica negativo")
	require.EqualValues(t, 100, balB)
	require.EqualValues(t, -100, balExt)
	require.EqualValues(t, 0, balA+balB+balExt, "Σ saldos = 0 (conservação)")
}
