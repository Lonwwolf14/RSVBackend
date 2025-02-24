-- +goose Up
CREATE TABLE
    trains (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        total_seats INTEGER NOT NULL
    );

CREATE TABLE
    tickets (
        id TEXT PRIMARY KEY,
        train_id TEXT NOT NULL,
        user_id TEXT NOT NULL,
        seat_number INTEGER NOT NULL,
        booked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (train_id) REFERENCES trains (id),
        FOREIGN KEY (user_id) REFERENCES users (id),
        UNIQUE (train_id, seat_number)
    );

INSERT INTO
    trains (id, name, total_seats)
VALUES
    (
        '550e8400-e29b-41d4-a716-446655440000',
        'Express 101',
        50
    ),
    (
        '550e8400-e29b-41d4-a716-446655440001',
        'Night Rider',
        60
    );

-- +goose Down
DROP TABLE tickets;

DROP TABLE trains;