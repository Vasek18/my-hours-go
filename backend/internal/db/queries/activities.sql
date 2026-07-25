-- name: ListActivitiesInRange :many
SELECT * FROM activities
WHERE user_id = @user_id
  AND start_time >= @from_time
  AND start_time < @to_time
ORDER BY start_time;

-- name: GetActivity :one
SELECT * FROM activities
WHERE id = $1 AND user_id = $2;

-- name: CreateActivity :one
INSERT INTO activities (user_id, type_id, description, start_time, end_time)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateActivity :one
UPDATE activities
SET type_id = $3,
    description = $4,
    start_time = $5,
    end_time = $6,
    updated_at = now()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteActivity :execrows
DELETE FROM activities
WHERE id = $1 AND user_id = $2;
