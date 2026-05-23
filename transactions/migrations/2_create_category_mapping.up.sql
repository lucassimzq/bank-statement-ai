CREATE TABLE category_mapping (
    id               UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_pattern TEXT        NOT NULL UNIQUE,
    category_id      UUID        NOT NULL REFERENCES categories(id),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
