-- name: QueryCategories :many
SELECT id, name, slug, created_at FROM categories ORDER BY name;

-- name: GetCategoryBySlug :one
SELECT id, name, slug, created_at FROM categories WHERE slug = $1;

-- name: InsertTransaction :one
INSERT INTO transactions (statement_id, card_id, txn_date, merchant_raw, merchant, amount, category_id)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, statement_id, card_id, txn_date, merchant_raw, merchant, amount, category_id, created_at;

-- name: GetTransactionByID :one
SELECT t.id, t.statement_id, t.card_id, t.txn_date, t.merchant_raw, t.merchant, t.amount,
       c.id AS category_id, c.name AS category_name, c.slug AS category_slug,
       t.created_at
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE t.id = $1;

-- name: QueryTransactionsByStatement :many
SELECT t.id, t.statement_id, t.card_id, t.txn_date, t.merchant_raw, t.merchant, t.amount,
       c.id AS category_id, c.name AS category_name, c.slug AS category_slug,
       t.created_at
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE t.statement_id = $1
ORDER BY t.txn_date;

-- name: QueryTransactionsByCardAndMonth :many
SELECT t.id, t.statement_id, t.card_id, t.txn_date, t.merchant_raw, t.merchant, t.amount,
       c.id AS category_id, c.name AS category_name, c.slug AS category_slug,
       t.created_at
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE t.card_id = $1
  AND t.txn_date >= make_date($2::int, $3::int, 1)
  AND t.txn_date < make_date($2::int, $3::int, 1) + INTERVAL '1 month'
ORDER BY t.txn_date;

-- name: DeleteTransactionsByStatement :exec
DELETE FROM transactions WHERE statement_id = $1;

-- name: QueryMonthlySpending :many
SELECT
    card_id,
    EXTRACT(YEAR  FROM txn_date)::int AS year,
    EXTRACT(MONTH FROM txn_date)::int AS month,
    SUM(amount::numeric)::text        AS total
FROM transactions
WHERE amount::numeric > 0
GROUP BY card_id, year, month
ORDER BY year, month, card_id;

-- name: QueryCategoryMappings :many
SELECT cm.id, cm.merchant_pattern, c.slug AS category_slug, c.name AS category_name, cm.created_at
FROM category_mapping cm
JOIN categories c ON c.id = cm.category_id
ORDER BY cm.merchant_pattern;

-- name: InsertCategoryMapping :one
INSERT INTO category_mapping (merchant_pattern, category_id)
VALUES ($1, (SELECT id FROM categories WHERE slug = $2))
RETURNING id, merchant_pattern, category_id, created_at;

-- name: DeleteCategoryMapping :exec
DELETE FROM category_mapping WHERE id = $1;
