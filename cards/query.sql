-- name: QueryBanks :many
SELECT id, name, slug, created_at FROM banks ORDER BY name;

-- name: BankExists :one
SELECT EXISTS(SELECT 1 FROM banks WHERE id = $1);

-- name: InsertCard :one
INSERT INTO cards (bank_id, label, purpose, last4)
VALUES ($1, $2, $3, $4)
RETURNING id, bank_id, label, purpose, last4, created_at;

-- name: QueryCards :many
SELECT c.id, c.bank_id, b.name AS bank_name, c.label, c.purpose, c.last4, c.created_at
FROM cards c
JOIN banks b ON b.id = c.bank_id
ORDER BY c.created_at;

-- name: GetCardByID :one
SELECT c.id, c.bank_id, b.name AS bank_name, c.label, c.purpose, c.last4, c.created_at
FROM cards c
JOIN banks b ON b.id = c.bank_id
WHERE c.id = $1;
