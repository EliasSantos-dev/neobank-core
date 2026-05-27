// Package test fornece infraestrutura compartilhada para testes de integração:
// cria um banco efêmero no Postgres local, aplica as migrations e o derruba no fim.
// Não usa Docker.
package test

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/url"
	"os"
	"testing"

	"github.com/EliasSantos-dev/neobank-core/internal/store"
	"github.com/EliasSantos-dev/neobank-core/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib" // registra o driver "pgx" para database/sql
	"github.com/pressly/goose/v3"
)

// adminDSN aponta para um banco existente usado só para criar/derrubar os efêmeros.
func adminDSN() string {
	if v := os.Getenv("NEOBANK_TEST_DSN"); v != "" {
		return v
	}
	return "postgres://elias-santos@/neobank_dev?host=/var/run/postgresql&sslmode=disable"
}

func dsnWithDB(base, db string) string {
	u, err := url.Parse(base)
	if err != nil {
		return base
	}
	u.Path = "/" + db
	return u.String()
}

func randDBName(t *testing.T) string {
	t.Helper()
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("rand: %v", err)
	}
	return "neobank_test_" + hex.EncodeToString(b)
}

// NewPostgres sobe um banco efêmero migrado e devolve um pool pronto.
func NewPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()
	base := adminDSN()
	name := randDBName(t)

	admin, err := sql.Open("pgx", base)
	if err != nil {
		t.Fatalf("abrir admin: %v", err)
	}
	if _, err := admin.ExecContext(ctx, "CREATE DATABASE "+name); err != nil {
		_ = admin.Close()
		t.Fatalf("criar database: %v", err)
	}
	_ = admin.Close()

	dsn := dsnWithDB(base, name)

	// Aplica migrations via goose (FS embutido -> independe do diretório).
	mdb, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("abrir db de migração: %v", err)
	}
	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("dialect: %v", err)
	}
	if err := goose.Up(mdb, "."); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	_ = mdb.Close()

	pool, err := store.NewPool(ctx, dsn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
		if adminCleanup, err := sql.Open("pgx", base); err == nil {
			_, _ = adminCleanup.ExecContext(context.Background(), "DROP DATABASE IF EXISTS "+name+" WITH (FORCE)")
			_ = adminCleanup.Close()
		}
	})
	return pool
}
