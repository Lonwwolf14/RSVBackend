-- +goose Up
CREATE TABLE
    trains (
        id UUID PRIMARY KEY,
        name TEXT NOT NULL,
        total_seats INT NOT NULL
    );

CREATE TABLE
    tickets (
        id UUID PRIMARY KEY,
        train_id UUID NOT NULL,
        user_id UUID NOT NULL,
        seat_number INT NOT NULL,
        booked_at TIMESTAMP
        WITH
            TIME ZONE DEFAULT NOW (),
            FOREIGN KEY (train_id) REFERENCES trains (id),
            FOREIGN KEY (user_id) REFERENCES users (id),
            UNIQUE (train_id, seat_number) -- Ensures a seat can only be booked once
    );

-- Seed some sample trains
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