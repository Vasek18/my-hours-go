CREATE TABLE activities (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    type_id     uuid        REFERENCES activity_types (id) ON DELETE SET NULL,
    description text        NOT NULL,
    start_time  timestamptz NOT NULL,
    end_time    timestamptz NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX activities_user_id_start_time_idx ON activities (user_id, start_time);
