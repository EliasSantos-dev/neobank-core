-- +goose Up
INSERT INTO accounts (id, currency, type)
VALUES ('00000000-0000-0000-0000-000000000002', 'BRL', 'external');

CREATE TABLE payments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id      UUID NOT NULL REFERENCES accounts(id),
    kind            TEXT NOT NULL CHECK (kind IN ('deposit','withdrawal')),
    amount          BIGINT NOT NULL CHECK (amount > 0),
    currency        CHAR(3) NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','completed','failed')),
    provider_ref    TEXT,
    idempotency_key TEXT NOT NULL UNIQUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE webhook_events (
    event_id    UUID PRIMARY KEY,
    payment_id  UUID NOT NULL,
    consumed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE webhook_events;
DROP TABLE payments;
DELETE FROM accounts WHERE id = '00000000-0000-0000-0000-000000000002';
