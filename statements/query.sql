-- name: InsertStatement :one
INSERT INTO statements (card_id, year, month, file_path)
VALUES ($1, $2, $3, $4)
RETURNING id, card_id, year, month, statement_bal, file_path, parsed_at, created_at;

-- name: GetStatementByID :one
SELECT id, card_id, year, month, statement_bal, file_path, parsed_at, created_at
FROM statements
WHERE id = $1;

-- name: QueryStatementsByCard :many
SELECT id, card_id, year, month, statement_bal, file_path, parsed_at, created_at
FROM statements
WHERE card_id = $1
ORDER BY year DESC, month DESC;

-- name: UpdateStatementBalance :one
UPDATE statements
SET statement_bal = $2, parsed_at = NOW()
WHERE id = $1
RETURNING id, card_id, year, month, statement_bal, file_path, parsed_at, created_at;
