-- +goose Up
CREATE TABLE users (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email             TEXT NOT NULL UNIQUE,
    password_hash     TEXT NOT NULL,
    wallet_account_id UUID NOT NULL REFERENCES accounts(id),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Tesouraria BRL: conta external de UUID fixo (fonte/sumidouro do funding simulado).
INSERT INTO accounts (id, currency, type)
VALUES ('00000000-0000-0000-0000-000000000001', 'BRL', 'external');

-- +goose Down
DELETE FROM accounts WHERE id = '00000000-0000-0000-0000-000000000001';
DROP TABLE users;
