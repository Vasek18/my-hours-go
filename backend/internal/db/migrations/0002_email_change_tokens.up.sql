CREATE TABLE email_change_tokens (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    new_email  text        NOT NULL,
    token_hash text        NOT NULL,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX email_change_tokens_token_hash_key ON email_change_tokens (token_hash);
CREATE INDEX email_change_tokens_user_id_idx ON email_change_tokens (user_id);
