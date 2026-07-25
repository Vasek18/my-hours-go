-- name: ListDayNotesInRange :many
SELECT * FROM day_notes
WHERE user_id = @user_id
  AND note_date >= @from_date
  AND note_date < @to_date
ORDER BY note_date;

-- name: GetDayNote :one
SELECT * FROM day_notes
WHERE user_id = $1 AND note_date = $2;

-- name: UpsertDayNote :one
INSERT INTO day_notes (user_id, note_date, content)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, note_date)
DO UPDATE SET content = EXCLUDED.content, updated_at = now()
RETURNING *;

-- name: DeleteDayNote :execrows
DELETE FROM day_notes
WHERE user_id = $1 AND note_date = $2;
