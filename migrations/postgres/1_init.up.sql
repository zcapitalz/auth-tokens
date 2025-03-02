CREATE TABLE refresh_tokens (
    id uuid DEFAULT gen_random_uuid() PRIMARY KEY,
    value_hash BYTEA NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL
);