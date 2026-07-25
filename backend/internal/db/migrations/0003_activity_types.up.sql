CREATE TABLE activity_types (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name       text        NOT NULL,
    color      text        NOT NULL,
    sort       integer     NOT NULL DEFAULT 500 CHECK (sort >= 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX activity_types_user_id_idx ON activity_types (user_id);
