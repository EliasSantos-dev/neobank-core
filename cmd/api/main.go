package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/EliasSantos-dev/neobank-core/internal/api"
	"github.com/EliasSantos-dev/neobank-core/internal/audit"
	"github.com/EliasSantos-dev/neobank-core/internal/events"
	"github.com/EliasSantos-dev/neobank-core/internal/gateway"
	"github.com/EliasSantos-dev/neobank-core/internal/relay"
	"github.com/EliasSantos-dev/neobank-core/internal/risk"
	"github.com/EliasSantos-dev/neobank-core/internal/store"
	"github.com/EliasSantos-dev/neobank-core/internal/worker"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		dsn = "postgres://neobank:neobank@localhost:5432/neobank?sslmode=disable"
	}

	secret := []byte(os.Getenv("JWT_SECRET"))
	if len(secret) == 0 {
		log.Warn("JWT_SECRET vazio — usando segredo de desenvolvimento (NÃO use em produção)")
		secret = []byte("dev-secret-change-me")
	}

	adminToken := os.Getenv("ADMIN_TOKEN")
	if adminToken == "" {
		log.Warn("ADMIN_TOKEN vazio — endpoints administrativos ficarão inacessíveis")
	}

	pool, err := store.NewPool(context.Background(), dsn)
	if err != nil {
		log.Error("conexão com o banco", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	st := store.New(pool)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Worker de risco em background: pontua intenções pendentes.
	eng := risk.NewEngine(risk.RuleReasoningAdvisor{}, 50)
	w := worker.New(st, eng)
	go w.Run(ctx, 2*time.Second)

	// Barramento de eventos: NATS embutido + auditoria + relay do outbox.
	bus, err := events.NewBus()
	if err != nil {
		log.Error("nats embutido", "err", err)
		os.Exit(1)
	}
	defer bus.Close()
	if err := audit.Start(ctx, bus, st); err != nil {
		log.Error("audit", "err", err)
		os.Exit(1)
	}
	go relay.New(st, bus).Run(ctx, 1*time.Second)

	// Gateway de pagamento (provedor fake; trocável por HTTP real).
	gw := gateway.NewService(st, gateway.NewFakeProvider())

	srv := api.NewServer(st, secret, adminToken, gw)
	addr := ":8080"
	log.Info("neobank ouvindo", "addr", addr)
	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Error("servidor", "err", err)
		os.Exit(1)
	}
}
