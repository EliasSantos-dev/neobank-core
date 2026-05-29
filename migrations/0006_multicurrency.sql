-- +goose Up
INSERT INTO accounts (id, currency, type) VALUES
  ('00000000-0000-0000-0000-000000000010','BRL','external'),
  ('00000000-0000-0000-0000-000000000011','USD','external'),
  ('00000000-0000-0000-0000-000000000012','EUR','external');

CREATE TABLE user_wallets (
    user_id    UUID NOT NULL REFERENCES users(id),
    currency   CHAR(3) NOT NULL,
    account_id UUID NOT NULL REFERENCES accounts(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, currency)
);

INSERT INTO user_wallets (user_id, currency, account_id)
SELECT id, 'BRL', wallet_account_id FROM users;

-- +goose Down
DROP TABLE user_wallets;
DELETE FROM accounts WHERE id IN (
  '00000000-0000-0000-0000-000000000010',
  '00000000-0000-0000-0000-000000000011',
  '00000000-0000-0000-0000-000000000012'
);
