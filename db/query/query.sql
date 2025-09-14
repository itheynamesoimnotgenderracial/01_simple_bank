-- name: CreateAccount :one
INSERT INTO accounts (
  owner, balance, currency
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetAccount :one
SELECT * FROM accounts
WHERE id = $1 LIMIT 1;

-- name: GetAccountForUpdate :one
SELECT * FROM accounts
WHERE id = $1 LIMIT 1 
FOR NO KEY UPDATE;

-- name: ListAccounts :many
SELECT * FROM accounts
WHERE owner = $1
ORDER BY id
LIMIT $2
OFFSET $3;

-- name: UpdateAccount :one
UPDATE accounts
SET balance = $2
WHERE id = $1
RETURNING *;

-- name: DeleteAccount :exec
DELETE FROM accounts
WHERE id = $1;

-- name: CreateEntry :one
INSERT INTO entries (
  account_id, amount
) VALUES (
  $1, $2
)
RETURNING *;

-- name: GetEntry :one
SELECT * FROM entries
WHERE id = $1 LIMIT 1;

-- name: ListEntries :many
SELECT * FROM entries
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: CreateTransfer :one
INSERT INTO transfers (
  form_account_id, to_account_id, amount
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: GetTransfer :one
SELECT * FROM transfers
WHERE id = $1 LIMIT 1;

-- name: ListTransfers :many
SELECT * FROM transfers
ORDER BY id
LIMIT $1
OFFSET $2;

-- name: AddAccountBalance :one
UPDATE accounts
SET balance = balance + sqlc.arg(amount)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: CreateUser :one
INSERT INTO users (
  username,
  hashed_password,
  full_name,
  email
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE username = $1 LIMIT 1;


-- name: CreateSession :one
INSERT INTO sessions (
  id,
  username,
  refresh_token,
  user_agent,
  client_ip,
  is_blocked,
  expires_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetSession :one
SELECT * FROM sessions
WHERE username = $1 LIMIT 1;

-- name: UpdateUser :one
UPDATE users
SET
  hashed_password = coalesce(sqlc.narg(hashed_password), hashed_password),
  password_changed_at = coalesce(sqlc.narg(password_changed_at), password_changed_at),
  full_name = coalesce(sqlc.narg(full_name), full_name),
  email = coalesce(sqlc.narg(email), email),
  is_email_verified = coalesce(sqlc.narg(is_email_verified), is_email_verified)
WHERE username = sqlc.arg(username)
RETURNING *;

-- name: CreateVerifyEmail :one
INSERT INTO verify_emails (
  username,
  email,
  secret_code
) VALUES (
  $1, $2, $3
)
RETURNING *;


-- name: UpdateVerifyEmail :one
UPDATE verify_emails
SET
  is_used = TRUE
WHERE 
  id = @id
  AND secret_code = @secret_code
  AND is_used = FALSE
  AND expires_at > now()
RETURNING *;