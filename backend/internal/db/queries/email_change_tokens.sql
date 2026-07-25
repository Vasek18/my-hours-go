-- name: CreateEmailChangeToken :one
INSERT INTO email_change_tokens (user_id, new_email, token_hash, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetEmailChangeTokenByHash :one
SELECT * FROM email_change_tokens
WHERE token_hash = $1;

-- name: DeleteEmailChangeToken :exec
DELETE FROM email_change_tokens
WHERE id = $1;

-- name: DeleteEmailChangeTokensForUser :exec
DELETE FROM email_change_tokens
WHERE user_id = $1;
