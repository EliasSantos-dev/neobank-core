package store_test

import (
	"context"
	"testing"

	"github.com/EliasSantos-dev/neobank-core/internal/store"
	"github.com/EliasSantos-dev/neobank-core/internal/user"
	itest "github.com/EliasSantos-dev/neobank-core/test"
	"github.com/stretchr/testify/require"
)

func TestCreateUser(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	ctx := context.Background()

	u, err := s.CreateUser(ctx, "a@b.com", "hash1")
	require.NoError(t, err)
	require.NotEqual(t, u.ID, u.WalletAccountID)

	bal, err := s.Balance(ctx, u.WalletAccountID)
	require.NoError(t, err)
	require.EqualValues(t, 0, bal)

	got, hash, err := s.GetUserByEmail(ctx, "a@b.com")
	require.NoError(t, err)
	require.Equal(t, u.ID, got.ID)
	require.Equal(t, "hash1", hash)

	byID, err := s.GetUserByID(ctx, u.ID)
	require.NoError(t, err)
	require.Equal(t, "a@b.com", byID.Email)

	_, err = s.CreateUser(ctx, "a@b.com", "x")
	require.ErrorIs(t, err, user.ErrEmailTaken)
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	pool := itest.NewPostgres(t)
	s := store.New(pool)
	_, _, err := s.GetUserByEmail(context.Background(), "ninguem@x.com")
	require.ErrorIs(t, err, user.ErrUserNotFound)
}
