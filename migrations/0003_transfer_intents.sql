-- +goose Up
CREATE TABLE transfer_intents (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    idempotency_key TEXT NOT NULL UNIQUE,
    from_account_id UUID NOT NULL REFERENCES accounts(id),
    to_account_id   UUID NOT NULL REFERENCES accounts(id),
    amount          BIGINT NOT NULL CHECK (amount > 0),
    currency        CHAR(3) NOT NULL,
    status          TEXT NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','under_review','completed','rejected')),
    risk_score      INT,
    risk_level      TEXT,
    risk_reasons    JSONB,
    transfer_id     UUID REFERENCES transfers(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_intents_pending ON transfer_intents (created_at) WHERE status = 'pending';

-- +goose Down
DROP TABLE transfer_intents;
