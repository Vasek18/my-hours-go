-- name: ListActivityTypes :many
SELECT * FROM activity_types
WHERE user_id = $1
ORDER BY sort ASC, created_at ASC;

-- name: GetActivityType :one
SELECT * FROM activity_types
WHERE id = $1 AND user_id = $2;

-- name: CreateActivityType :one
INSERT INTO activity_types (user_id, name, color, sort)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateActivityType :one
UPDATE activity_types
SET name = $3,
    color = $4,
    sort = $5,
    updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteActivityType :execrows
DELETE FROM activity_types
WHERE id = $1 AND user_id = $2;
