CREATE TABLE cards (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    bank_id    UUID NOT NULL REFERENCES banks(id),
    label      TEXT NOT NULL,
    purpose    TEXT NOT NULL,
    last4      CHAR(4) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
