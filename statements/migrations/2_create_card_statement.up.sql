-- Junction table: one row per card detected in a statement.
-- card_id is nullable — NULL when the card was not found in our system at parse time.
-- status: 1 = parsed (transactions stored), 2 = skipped (card not in system)
CREATE TABLE card_statement (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    statement_id  UUID        NOT NULL REFERENCES statements(id),
    card_last4    TEXT        NOT NULL,
    card_id       UUID,
    status        SMALLINT    NOT NULL DEFAULT 1,
    statement_bal NUMERIC(12, 2),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (statement_id, card_last4)
);

CREATE INDEX card_statement_card_id_idx ON card_statement (card_id);
