-- name: InsertStatement :one
INSERT INTO statements (file_hash, file_path, status)
VALUES ($1, $2, 0)
RETURNING id, year, month, statement_bal, file_path, status, message, file_hash, parsed_at, created_at;

-- name: GetStatementByID :one
SELECT id, year, month, statement_bal, file_path, status, message, file_hash, parsed_at, created_at
FROM statements
WHERE id = $1;

-- name: QueryStatementsByCard :many
SELECT s.id, s.year, s.month, s.statement_bal, s.file_path, s.status, s.message, s.file_hash, s.parsed_at, s.created_at
FROM statements s
JOIN card_statement cs ON cs.statement_id = s.id
WHERE cs.card_id = $1
ORDER BY s.created_at DESC;

-- name: UpdateStatementParsed :one
UPDATE statements
SET year          = $2,
    month         = $3,
    statement_bal = $4,
    status        = 1,
    parsed_at     = NOW()
WHERE id = $1
RETURNING id, year, month, statement_bal, file_path, status, message, file_hash, parsed_at, created_at;

-- name: UpdateStatementError :one
UPDATE statements
SET status  = 2,
    message = $2
WHERE id = $1
RETURNING id, year, month, statement_bal, file_path, status, message, file_hash, parsed_at, created_at;

-- name: UpdateStatementBalance :one
UPDATE statements
SET statement_bal = $2,
    parsed_at     = NOW()
WHERE id = $1
RETURNING id, year, month, statement_bal, file_path, status, message, file_hash, parsed_at, created_at;

-- Reject only when the statement is fully parsed (no skipped card_statements).
-- name: StatementFullyParsedByHash :one
SELECT EXISTS(
    SELECT 1 FROM statements s
    WHERE s.file_hash = $1
      AND s.status = 1
      AND NOT EXISTS (
          SELECT 1 FROM card_statement cs
          WHERE cs.statement_id = s.id AND cs.status = 2
      )
);

-- name: InsertCardStatement :one
INSERT INTO card_statement (statement_id, card_last4, card_id, status, statement_bal)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, statement_id, card_last4, card_id, status, statement_bal, created_at;

-- name: CardStatementExistsForPeriod :one
SELECT EXISTS(
    SELECT 1
    FROM card_statement cs
    JOIN statements s ON s.id = cs.statement_id
    WHERE cs.card_id = $1
      AND s.year     = $2
      AND s.month    = $3
      AND cs.status  = 1
);

-- name: ListAllStatements :many
SELECT id, year, month, statement_bal, file_path, status, message, file_hash, parsed_at, created_at
FROM statements
ORDER BY created_at DESC;

-- name: GetCardStatementsByStatementID :many
SELECT id, statement_id, card_last4, card_id, status, statement_bal, created_at
FROM card_statement
WHERE statement_id = $1
ORDER BY created_at ASC;

-- name: ResetStatementForRetry :one
UPDATE statements
SET status  = 0,
    message = NULL
WHERE id = $1
RETURNING id, year, month, statement_bal, file_path, status, message, file_hash, parsed_at, created_at;

-- name: DeleteCardStatementsByStatementID :exec
DELETE FROM card_statement WHERE statement_id = $1;

-- name: DeleteStatement :exec
DELETE FROM statements WHERE id = $1;
