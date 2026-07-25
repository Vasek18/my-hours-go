CREATE TABLE day_notes (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    note_date  date        NOT NULL,
    content    text        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- One note per user per day.
CREATE UNIQUE INDEX day_notes_user_id_note_date_key ON day_notes (user_id, note_date);
