CREATE TABLE IF NOT EXISTS user_invitations (
    token bytea primary key,
    user_id bigint not null,

    FOREIGN KEY(user_id) REFERENCES users (id) ON DELETE CASCADE
);