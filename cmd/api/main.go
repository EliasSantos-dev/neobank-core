package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/EliasSantos-dev/neobank-core/internal/api"
	"github.com/EliasSantos-dev/neobank-core/internal/store"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		dsn = "postgres://neobank:neobank@localhost:5432/neobank?sslmode=disable"
	}

	pool, err := store.NewPool(context.Background(), dsn)
	if err != nil {
		log.Error("conexão com o banco", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	srv := api.NewServer(store.New(pool))
	addr := ":8080"
	log.Info("ledger ouvindo", "addr", addr)
	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Error("servidor", "err", err)
		os.Exit(1)
	}
}
