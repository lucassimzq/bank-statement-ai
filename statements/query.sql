-- name: InsertStatement :one
INSERT INTO statements (card_id, file_hash, file_path, status)
VALUES ($1, $2, $3, 0)
RETURNING id, card_id, year, month, statement_bal, file_path, parsed_at, created_at, status, message, file_hash;

-- name: GetStatementByID :one
SELECT id, card_id, year, month, statement_bal, file_path, parsed_at, created_at, status, message, file_hash
FROM statements
WHERE id = $1;

-- name: QueryStatementsByCard :many
SELECT id, card_id, year, month, statement_bal, file_path, parsed_at, created_at, status, message, file_hash
FROM statements
WHERE card_id = $1
ORDER BY created_at DESC;

-- name: UpdateStatementParsed :one
UPDATE statements
SET year          = $2,
    month         = $3,
    statement_bal = $4,
    status        = 1,
    parsed_at     = NOW()
WHERE id = $1
RETURNING id, card_id, year, month, statement_bal, file_path, parsed_at, created_at, status, message, file_hash;

-- name: UpdateStatementError :one
UPDATE statements
SET status  = 2,
    message = $2
WHERE id = $1
RETURNING id, card_id, year, month, statement_bal, file_path, parsed_at, created_at, status, message, file_hash;

-- name: UpdateStatementBalance :one
UPDATE statements
SET statement_bal = $2,
    parsed_at     = NOW()
WHERE id = $1
RETURNING id, card_id, year, month, statement_bal, file_path, parsed_at, created_at, status, message, file_hash;

-- name: StatementExistsByHash :one
SELECT EXISTS(SELECT 1 FROM statements WHERE file_hash = $1 AND status = 1);

-- name: StatementExistsByCardPeriod :one
SELECT EXISTS(
    SELECT 1 FROM statements
    WHERE card_id = $1
      AND year    = $2
      AND month   = $3
      AND status != 2
);
