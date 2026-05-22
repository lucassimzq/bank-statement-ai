CREATE TABLE statements (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    card_id       UUID NOT NULL,
    year          INT NOT NULL,
    month         INT NOT NULL,
    statement_bal NUMERIC(12,2),
    file_path     TEXT,
    parsed_at     TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (card_id, year, month)
);
