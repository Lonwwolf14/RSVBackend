-- +goose Up
CREATE TABLE
    users (
        id UUID PRIMARY KEY,
        email TEXT UNIQUE NOT NULL,
        password TEXT NOT NULL -- Store hashed password
    );

-- +goose Down
DROP TABLE users;