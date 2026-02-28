CREATE TABLE IF NOT EXISTS comments(
    id Bigserial PRIMARY KEY,
    post_id Bigserial NOT NULL,
    user_id Bigserial NOT NULL,
    content Text NOT NULL,
    created_at TIMESTAMP(0) WITH TIME ZONE NOT NULL DEFAULT NOW()
)