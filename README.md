# neobank-core

![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16+-4169E1?logo=postgresql&logoColor=white)
![Tests](https://img.shields.io/badge/tests-race--checked-success)

Núcleo de **ledger de partidas dobradas** para uma carteira digital, em Go + PostgreSQL. Correto, concorrente-safe e auditável por construção.

Este repositório é o **M1 (Ledger Core)** de uma plataforma maior (ver [Roadmap](#roadmap)). O foco aqui é o que mais importa em fintech: **integridade do dinheiro sob concorrência**.

## Princípio: history-as-truth

O saldo **nunca é mutado**. A única fonte da verdade é o histórico imutável de lançamentos (`entries`, append-only — `UPDATE`/`DELETE` bloqueados no próprio banco). O saldo é **derivado**:

```sql
saldo = SUM(créditos) − SUM(débitos)
```

Consequência: é impossível o saldo divergir da história, porque ele **é** a história. Inspirado no design de ledgers como o [TigerBeetle](https://tigerbeetle.com/).

### Invariantes garantidas

- **Partidas dobradas:** toda transferência gera um débito e um crédito de igual valor (Σ débitos = Σ créditos).
- **Conservação:** a soma dos saldos é constante — dinheiro não nasce nem some.
- **Não-negatividade:** contas `wallet` não ficam negativas; contas `external` (funding) podem.
- **Imutabilidade:** lançamentos nunca mudam.
- **Exactly-once:** `Idempotency-Key` única — retries não duplicam efeito.
- **Inteiros:** valores em unidades mínimas (`int64`), nunca `float`.

## Concorrência

Sem saldo mutável para travar, a serialização por conta é feita travando a **âncora estável** (`accounts`) dentro de uma transação:

```sql
SELECT ... FROM accounts WHERE id IN ($1,$2) ORDER BY id FOR UPDATE;
```

- A trava em `accounts` serializa transferências concorrentes na mesma conta → **zero double-spend**.
- `ORDER BY id` garante ordem determinística de aquisição → **sem deadlock**.
- Decisão consciente: **não** se trava "o último lançamento" (problema de phantom com inserts concorrentes).

> **Alternativa:** isolamento `SERIALIZABLE` + retry no erro `40001`, sem lock explícito. O lock de âncora foi escolhido por previsibilidade sob contenção; a alternativa é um próximo passo natural de benchmark.

O teste `TestTransfer_Concurrent_NoDoubleSpend` dispara 50 transferências paralelas e prova, **com `go test -race`**, que nenhum saldo fica negativo e que a conservação é mantida.

## API

| Método | Rota | Descrição |
|--------|------|-----------|
| `POST` | `/accounts` | Cria conta (`{"currency":"BRL","type":"wallet"}`). |
| `POST` | `/transfers` | Move dinheiro. Header `Idempotency-Key` obrigatório. |
| `GET`  | `/accounts/{id}/balance` | Saldo derivado atual. |

Erros de domínio mapeados para HTTP: saldo insuficiente → `422`, conflito de idempotência → `409`, valor inválido → `400`.

## Como rodar

### Com Docker

```bash
make up          # sobe Postgres
make migrate     # aplica migrations
make run         # API em :8080
```

### Com um Postgres local (sem Docker)

```bash
createdb neobank
DB_URL="postgres://$USER@/neobank?host=/var/run/postgresql&sslmode=disable" make migrate
DB_URL="postgres://$USER@/neobank?host=/var/run/postgresql&sslmode=disable" make run
```

### Testes

```bash
make test    # todos os testes (cada um cria um banco efêmero)
make race    # com detector de corrida — inclui o teste de concorrência
```

Os testes de integração criam/derrubam um banco efêmero por teste. Por padrão usam o Postgres local via socket; ajuste com `NEOBANK_TEST_DSN`.

## Arquitetura

```
cmd/api        # entrypoint HTTP
cmd/migrate    # aplica migrations embutidas (goose)
internal/
  money/       # value object de dinheiro (puro)
  ledger/      # domínio puro: tipos, invariantes, BuildEntries
  store/       # adapter Postgres (pgx): transações, lock, idempotência
  api/         # transporte HTTP
migrations/    # schema append-only (embutido via go:embed)
```

Fronteiras isoladas: o domínio (`ledger`) não conhece banco nem HTTP e é testável sozinho; o `store` concentra a correção transacional; a `api` é só transporte.

## Roadmap

Este é o M1. Próximas fatias (cada uma com seu próprio ciclo):

- **M2** — Camada carteira/neobank: contas de usuário, depósitos/saques, auth, extrato.
- **M3** — IA de segurança de transações: risk scoring em tempo real (regras → ML), na frente do ledger.
- **M4** — Event-driven + reconciliação (outbox, eventos, replay).
- **M5** — Gateway de pagamento + multi-moeda/FX.
- **M6** — Credit scoring + dashboard de observabilidade.

## Fora de escopo do M1

Auth, multi-moeda, eventos, IA, gateway externo e snapshots de saldo são milestones posteriores. M1 é, deliberadamente, **só um ledger correto + API mínima + testes que provam a correção**.
