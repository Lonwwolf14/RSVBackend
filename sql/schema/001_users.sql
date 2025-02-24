-- +goose Up
CREATE TABLE
    users (id UUID PRIMARY KEY, email TEXT UNIQUE NOT NULL);

-- +goose Down
DROP TABLE users;