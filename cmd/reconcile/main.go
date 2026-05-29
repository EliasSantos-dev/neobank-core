// Comando reconcile roda a reconciliação e imprime o report (exit≠0 se não-saudável).
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/EliasSantos-dev/neobank-core/internal/recon"
	"github.com/EliasSantos-dev/neobank-core/internal/store"
)

func main() {
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		dsn = "postgres://neobank:neobank@localhost:5432/neobank?sslmode=disable"
	}
	pool, err := store.NewPool(context.Background(), dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, "conexão:", err)
		os.Exit(2)
	}
	defer pool.Close()

	rep, err := recon.Run(context.Background(), store.New(pool))
	if err != nil {
		fmt.Fprintln(os.Stderr, "reconcile:", err)
		os.Exit(2)
	}
	out, _ := json.MarshalIndent(rep, "", "  ")
	fmt.Println(string(out))
	if !rep.Healthy {
		os.Exit(1)
	}
}
