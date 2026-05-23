CREATE TABLE categories (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL UNIQUE,
    slug       TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO categories (name, slug) VALUES
    ('Dining',          'dining'),
    ('Groceries',       'groceries'),
    ('Online Shopping', 'online-shopping'),
    ('Transport',       'transport'),
    ('Insurance',       'insurance'),
    ('Entertainment',   'entertainment'),
    ('Health',          'health'),
    ('Utilities',       'utilities'),
    ('Travel',          'travel'),
    ('EWallet Topup',   'ewallet-topup'),
    ('Pet',             'pet'),
    ('Others',          'others');

CREATE TABLE transactions (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    statement_id  UUID NOT NULL,
    card_id       UUID NOT NULL,
    txn_date      DATE NOT NULL,
    merchant_raw  TEXT NOT NULL,
    merchant      TEXT NOT NULL,
    amount        NUMERIC(12,2) NOT NULL,
    category_id   UUID NOT NULL REFERENCES categories(id),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_statement ON transactions (statement_id);
CREATE INDEX idx_transactions_card_date ON transactions (card_id, txn_date);
