-- +goose Up
CREATE TABLE accounts (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    currency   CHAR(3) NOT NULL,
    type       TEXT NOT NULL CHECK (type IN ('wallet','external')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE transfers (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'completed',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE SEQUENCE entries_seq;

CREATE TABLE entries (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    seq         BIGINT NOT NULL DEFAULT nextval('entries_seq'),
    transfer_id UUID NOT NULL REFERENCES transfers(id),
    account_id  UUID NOT NULL REFERENCES accounts(id),
    direction   TEXT NOT NULL CHECK (direction IN ('debit','credit')),
    amount      BIGINT NOT NULL CHECK (amount > 0),
    currency    CHAR(3) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_entries_account_seq ON entries (account_id, seq);

-- Append-only: bloqueia UPDATE/DELETE no nível do banco.
CREATE RULE entries_no_update AS ON UPDATE TO entries DO INSTEAD NOTHING;
CREATE RULE entries_no_delete AS ON DELETE TO entries DO INSTEAD NOTHING;

-- +goose Down
DROP TABLE entries;
DROP SEQUENCE entries_seq;
DROP TABLE transfers;
DROP TABLE accounts;
