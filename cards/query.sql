-- name: QueryBanks :many
SELECT id, name, slug, logo_url, created_at FROM banks ORDER BY name;

-- name: UpsertBank :one
INSERT INTO banks (name, slug, logo_url)
VALUES ($1, $2, $3)
ON CONFLICT (slug) DO UPDATE
    SET name     = EXCLUDED.name,
        logo_url = EXCLUDED.logo_url
RETURNING id, name, slug, logo_url, created_at;

-- name: BankExists :one
SELECT EXISTS(SELECT 1 FROM banks WHERE id = $1);

-- name: InsertCard :one
INSERT INTO cards (bank_id, label, purpose, last4)
VALUES ($1, $2, $3, $4)
RETURNING id, bank_id, label, purpose, last4, created_at;

-- name: QueryCards :many
SELECT c.id, c.bank_id,
       b.name AS bank_name, b.slug AS bank_slug, b.logo_url AS bank_logo_url, b.created_at AS bank_created_at,
       c.label, c.purpose, c.last4, c.created_at
FROM cards c
JOIN banks b ON b.id = c.bank_id
ORDER BY c.created_at;

-- name: GetCardByID :one
SELECT c.id, c.bank_id,
       b.name AS bank_name, b.slug AS bank_slug, b.logo_url AS bank_logo_url, b.created_at AS bank_created_at,
       c.label, c.purpose, c.last4, c.created_at
FROM cards c
JOIN banks b ON b.id = c.bank_id
WHERE c.id = $1;

-- name: GetCardByLast4 :one
SELECT c.id, c.bank_id,
       b.name AS bank_name, b.slug AS bank_slug, b.logo_url AS bank_logo_url, b.created_at AS bank_created_at,
       c.label, c.purpose, c.last4, c.created_at
FROM cards c
JOIN banks b ON b.id = c.bank_id
WHERE c.last4 = $1
LIMIT 1;

-- name: GetCardByLast4AndBank :one
SELECT c.id, c.bank_id,
       b.name AS bank_name, b.slug AS bank_slug, b.logo_url AS bank_logo_url, b.created_at AS bank_created_at,
       c.label, c.purpose, c.last4, c.created_at
FROM cards c
JOIN banks b ON b.id = c.bank_id
WHERE c.last4 = $1
  AND b.slug  = $2
LIMIT 1;
