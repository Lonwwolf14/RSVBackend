-- name: CreateUser :one
INSERT INTO
    users (id, email, password)
VALUES
    (?, ?, ?) RETURNING *;

-- name: GetUserByEmail :one
SELECT
    *
FROM
    users
WHERE
    email = ?;