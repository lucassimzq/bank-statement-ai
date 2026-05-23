CREATE TABLE banks (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL UNIQUE,
    slug       TEXT NOT NULL UNIQUE,
    logo_url   TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO banks (name, slug) VALUES
    ('Maybank', 'maybank'),
    ('Alliance Bank', 'alliance'),
    ('HSBC', 'hsbc');
