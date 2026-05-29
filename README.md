# neobank-core

![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16+-4169E1?logo=postgresql&logoColor=white)
![Tests](https://img.shields.io/badge/tests-race--checked-success)

Núcleo de **ledger de partidas dobradas** para uma carteira digital, em Go + PostgreSQL. Correto, concorrente-safe e auditável por construção.

Cobre **M1 (Ledger Core)**, **M2 (Carteira/Auth)**, **M3 (Segurança de transações)**, **M4 (Event-driven + Reconciliação)** e **M5a (Gateway de pagamento)** de uma plataforma maior (ver [Roadmap](#roadmap)). O foco é o que mais importa em fintech: **integridade do dinheiro sob concorrência**, com carteira, risco com revisão humana, eventos confiáveis e cash-in/out via gateway por cima.

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

Carteira digital sobre o ledger. Autenticação por **JWT** (bcrypt nas senhas); rotas privadas exigem `Authorization: Bearer <token>`. Movimentações de dinheiro exigem header `Idempotency-Key`.

| Método | Rota | Auth | Descrição |
|--------|------|------|-----------|
| `POST` | `/auth/register` | — | Cria usuário + carteira (`{"email","password"}`). |
| `POST` | `/auth/login` | — | Devolve `{"access_token"}`. |
| `GET`  | `/me` | ✅ | Dados do usuário + saldo. |
| `POST` | `/me/deposit` | ✅ | `{"amount"}` — cash-in via gateway (`202`, confirma por webhook). |
| `POST` | `/me/withdraw` | ✅ | `{"amount"}` — cash-out via gateway com hold (`202`). |
| `GET`  | `/me/payments` | ✅ | Lista os pagamentos (depósitos/saques) do usuário. |
| `POST` | `/transfers` | ✅ | `{"to_email","amount"}` — cria **intenção** avaliada por risco (`202`). |
| `GET`  | `/me/transfers` | ✅ | Lista as intenções do usuário (status, score, nível). |
| `GET`  | `/me/statement` | ✅ | Extrato paginado (`?limit=&offset=`). |
| `POST` | `/webhooks/gateway` | — | Callback do provedor (idempotente). |
| `POST` | `/admin/transfers/{id}/review` | admin | `{"decision":"approve"\|"reject"}` (header `X-Admin-Token`). |
| `POST` | `/admin/reconcile` | admin | Roda a reconciliação e retorna o report. |

Depósito/saque são síncronos (tesouraria — funding simulado até haver gateway real). Transferência entre usuários passa pela **camada de risco** (ver abaixo). Erros mapeados: não autenticado → `401`, destinatário inexistente → `404`, e-mail já usado / intenção fora de revisão → `409`, saldo insuficiente → `422`, valor inválido → `400`.

## Segurança de transações (M3)

Transferências usuário→usuário não efetivam direto: viram uma **intenção** pontuada por um motor de risco. O dinheiro só toca o ledger quando aprovado — **sem clawback**, a integridade do M1 fica intocada.

```
POST /transfers → intent (pending) ──(worker pontua)──┬─ baixo risco → completed (efetiva)
                                                       └─ alto risco  → under_review ─(revisão)─┬ approve → completed
                                                                                                └ reject  → rejected
```

- **Motor em camadas** (`internal/risk`, domínio puro): regras determinísticas (velocity, valor atípico, novo destinatário, drain) + anomalia estatística (z-score do valor) + um **advisor** que explica o veredito em linguagem natural. O advisor é uma interface: o default é determinístico (`RuleReasoningAdvisor`); um `LLMAdvisor` pode plugar via `LLM_API_KEY` (seam pronto, default não usa rede).
- **Worker** (`internal/worker`): goroutine que pontua as intenções pendentes (`ProcessOnce`, testável). *Nota: roda um único worker in-process e processa sequencialmente; múltiplos workers exigiriam `FOR UPDATE SKIP LOCKED` — fora do escopo atual.*
- **Human-in-the-loop:** alto risco fica retido até um revisor (admin token) aprovar ou rejeitar.

## Event-driven + Reconciliação (M4)

Toda mudança de estado relevante emite um evento de domínio, com **entrega confiável** e um read-model de auditoria. Resolve o dual-write com o **Transactional Outbox**:

```
mudança de estado + evento  (MESMA transação → outbox)
        │
     relay (at-least-once) ──► NATS (embutido) ──► subscriber de auditoria ──► event_log (idempotente)
```

- **Outbox transacional** (`internal/store`): `transfer.completed`, `intent.under_review` e `intent.rejected` são gravados na mesma transação que os origina — se a tx aborta, não há evento órfão.
- **Relay** (`internal/relay`): publica os eventos não-enviados no barramento e marca como publicados. Semântica **at-least-once** — consumidores são idempotentes (não exactly-once; é o trade-off honesto do outbox).
- **Barramento NATS** (`internal/events`): servidor NATS **real, embutido no processo** com conexão in-process (sem Docker/porta). No deploy real, troca-se por um cluster NATS externo só na config de conexão.
- **Auditoria** (`internal/audit`): assina `transfer.>`/`intent.>` e materializa um `event_log` idempotente (dedup pela PK do evento) — um read-model durável.
- **Reconciliação** (`internal/recon`): verifica os invariantes contábeis — **conservação** (Σ saldos = 0), **partidas dobradas** (débito=crédito por transfer), **não-negatividade** das wallets e **ledger × gateway** (saldo da conta gateway bate com a tabela `payments`). Exposta em `POST /admin/reconcile` (admin token) e no job `cmd/reconcile` (exit ≠ 0 se não-saudável).

## Gateway de pagamento (M5a)

Depósito e saque são **cash-in/out reais via gateway**, confirmados de forma **assíncrona por webhook**. Substituem o atalho síncrono da tesouraria do M2.

```
POST /me/deposit  → payment(pending) + charge ─────────────► webhook succeeded → gateway→wallet (credita)
POST /me/withdraw → hold wallet→gateway (reserva) + payout ─► webhook succeeded → completed
                                                            └► webhook failed    → gateway→wallet (estorna)
```

- **Provedor plugável** (`internal/gateway`): interface `PaymentProvider` com um `FakeProvider` determinístico (recusa valores `% 100 == 13`, como cartões de teste). Um `HTTPProvider` (Stripe/AbacatePay) plugaria a mesma interface.
- **Hold no saque:** os fundos são reservados na solicitação (`wallet→gateway`), impedindo gasto duplo enquanto o payout está em trânsito; falha gera **estorno compensatório** (lançamento novo, não edição de histórico).
- **Webhook idempotente** (`POST /webhooks/gateway`): dedup por `event_id` + guarda de estado terminal — reentrega não duplica efeito. Em produção seria validado por assinatura HMAC do provedor (seam documentado).
- **Eventos** `payment.completed`/`payment.failed` no outbox, fluindo pela auditoria do M4.
- **Conservação** preservada em todos os caminhos; a conta `gateway` (external) representa o dinheiro em trânsito.

### Exemplo

```bash
curl -XPOST localhost:8080/auth/register -d '{"email":"a@b.com","password":"segredo123"}'
TOKEN=$(curl -s -XPOST localhost:8080/auth/login -d '{"email":"a@b.com","password":"segredo123"}' | jq -r .access_token)
curl -XPOST localhost:8080/me/deposit -H "Authorization: Bearer $TOKEN" -H "Idempotency-Key: dep-1" -d '{"amount":5000}'
curl localhost:8080/me -H "Authorization: Bearer $TOKEN"
```

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

> Defina `JWT_SECRET` e `ADMIN_TOKEN` ao subir a API (`make run`). Sem `JWT_SECRET`, um segredo de desenvolvimento é usado (com aviso) — não use em produção. Sem `ADMIN_TOKEN`, os endpoints administrativos ficam inacessíveis.

## Arquitetura

```
cmd/api        # entrypoint HTTP (sobe worker, relay, auditoria e gateway)
cmd/migrate    # aplica migrations embutidas (goose)
cmd/reconcile  # job de reconciliação dos invariantes
internal/
  money/       # value object de dinheiro (puro)
  ledger/      # domínio puro: tipos, invariantes, BuildEntries
  store/       # adapter Postgres (pgx): transações, lock, idempotência, outbox
  user/ auth/  # identidade e autenticação (bcrypt + JWT)
  risk/        # motor de risco em camadas (puro)
  worker/      # processa intenções de transferência pendentes
  events/      # barramento NATS embutido
  relay/       # publica o outbox no barramento (at-least-once)
  audit/       # consome eventos e materializa o event_log
  recon/       # verificação pura dos invariantes contábeis
  gateway/     # provedor de pagamento (fake) + service de cash-in/out
  api/         # transporte HTTP
migrations/    # schema append-only (embutido via go:embed)
```

Fronteiras isoladas: o domínio (`ledger`) não conhece banco nem HTTP e é testável sozinho; o `store` concentra a correção transacional; a `api` é só transporte.

## Roadmap

Plataforma construída em fatias, cada uma com seu próprio ciclo:

- ✅ **M1** — Ledger Core: partidas dobradas, saldo derivado, concorrência.
- ✅ **M2** — Carteira/Auth: usuários, JWT, depósito/saque/transferência, extrato.
- ✅ **M3** — Segurança de transações: motor de risco em camadas, worker e revisão humana.
- ✅ **M4** — Event-driven + reconciliação: outbox transacional, relay → NATS, auditoria, invariantes.
- ✅ **M5a** — Gateway de pagamento: cash-in/out via provedor plugável, webhook idempotente, hold no saque, recon×gateway.
- **M5b** — Multi-moeda / FX: contas multi-moeda, conversão com spread.
- **M6** — Credit scoring + dashboard de observabilidade.

## Fora de escopo (até aqui)

Multi-moeda, eventos/outbox, IA, gateway externo, snapshots de saldo, refresh tokens e reset de senha são milestones posteriores.
