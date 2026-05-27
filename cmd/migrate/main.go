// Comando migrate aplica as migrations embutidas (goose) ao banco em DB_URL.
package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/EliasSantos-dev/neobank-core/migrations"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		dsn = "postgres://neobank:neobank@localhost:5432/neobank?sslmode=disable"
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("abrir banco: %v", err)
	}
	defer db.Close()

	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("dialect: %v", err)
	}
	if err := goose.Up(db, "."); err != nil {
		log.Fatalf("migrate: %v", err)
	}
}
