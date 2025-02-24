-- +goose Up
CREATE TABLE
    users (
        id TEXT PRIMARY KEY,
        email TEXT UNIQUE NOT NULL,
        password TEXT NOT NULL
    );

-- +goose Down
DROP TABLE users;