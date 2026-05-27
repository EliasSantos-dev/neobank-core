DB_URL ?= postgres://neobank:neobank@localhost:5432/neobank?sslmode=disable

.PHONY: up down migrate run test race vet

up:      ## sobe o Postgres via docker compose
	docker compose up -d

down:    ## derruba o Postgres
	docker compose down

migrate: ## aplica as migrations em DB_URL
	DB_URL="$(DB_URL)" go run ./cmd/migrate

run:     ## roda a API em :8080
	DB_URL="$(DB_URL)" go run ./cmd/api

test:    ## roda todos os testes
	go test ./...

race:    ## roda os testes com detector de corrida
	go test -race ./...

vet:     ## roda o go vet
	go vet ./...
