// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: users.sql

package database

import (
	"context"
)

const createUser = `-- name: CreateUser :one
INSERT INTO
    users (id, email, password)
VALUES
    (?, ?, ?) RETURNING id, email, password
`

type CreateUserParams struct {
	ID       string
	Email    string
	Password string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.db.QueryRowContext(ctx, createUser, arg.ID, arg.Email, arg.Password)
	var i User
	err := row.Scan(&i.ID, &i.Email, &i.Password)
	return i, err
}

const getUserByEmail = `-- name: GetUserByEmail :one
SELECT
    id, email, password
FROM
    users
WHERE
    email = ?
`

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRowContext(ctx, getUserByEmail, email)
	var i User
	err := row.Scan(&i.ID, &i.Email, &i.Password)
	return i, err
}
