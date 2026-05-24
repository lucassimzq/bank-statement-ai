CREATE TABLE statements (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    year          INT,
    month         INT,
    statement_bal NUMERIC(12,2),
    file_path     TEXT,
    status        SMALLINT NOT NULL DEFAULT 0,
    message       TEXT,
    file_hash     TEXT,
    parsed_at     TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
