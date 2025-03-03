// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: tickets.sql

package database

import (
	"context"
	"database/sql"
)

const createTicket = `-- name: CreateTicket :one
INSERT INTO tickets (id, train_id, user_id, seat_number)
SELECT ?1, ?2, ?3, ?4
WHERE EXISTS (
    SELECT 1 
    FROM trains t
    WHERE t.id = ?2 
    AND ?4 BETWEEN 1 AND t.total_seats
    AND NOT EXISTS (
        SELECT 1 
        FROM tickets tk 
        WHERE tk.train_id = t.id 
        AND tk.seat_number = ?4
    )
)
RETURNING id, train_id, user_id, seat_number, booked_at
`

type CreateTicketParams struct {
	ID         string
	TrainID    string
	UserID     string
	SeatNumber int64
}

func (q *Queries) CreateTicket(ctx context.Context, arg CreateTicketParams) (Ticket, error) {
	row := q.db.QueryRowContext(ctx, createTicket,
		arg.ID,
		arg.TrainID,
		arg.UserID,
		arg.SeatNumber,
	)
	var i Ticket
	err := row.Scan(
		&i.ID,
		&i.TrainID,
		&i.UserID,
		&i.SeatNumber,
		&i.BookedAt,
	)
	return i, err
}

const deleteTicket = `-- name: DeleteTicket :exec
DELETE FROM tickets
WHERE id = ? AND user_id = ?
`

type DeleteTicketParams struct {
	ID     string
	UserID string
}

func (q *Queries) DeleteTicket(ctx context.Context, arg DeleteTicketParams) error {
	_, err := q.db.ExecContext(ctx, deleteTicket, arg.ID, arg.UserID)
	return err
}

const getAvailableTickets = `-- name: GetAvailableTickets :many
SELECT t.id, t.name, t.total_seats, 
       (t.total_seats - COUNT(tk.id)) AS available_seats
FROM trains t
LEFT JOIN tickets tk ON t.id = tk.train_id
GROUP BY t.id, t.name, t.total_seats
`

type GetAvailableTicketsRow struct {
	ID             string
	Name           string
	TotalSeats     int64
	AvailableSeats interface{}
}

func (q *Queries) GetAvailableTickets(ctx context.Context) ([]GetAvailableTicketsRow, error) {
	rows, err := q.db.QueryContext(ctx, getAvailableTickets)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAvailableTicketsRow
	for rows.Next() {
		var i GetAvailableTicketsRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.TotalSeats,
			&i.AvailableSeats,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUserTickets = `-- name: GetUserTickets :many
SELECT tk.id, t.name, tk.seat_number, tk.booked_at
FROM tickets tk
JOIN trains t ON tk.train_id = t.id
WHERE tk.user_id = ?
`

type GetUserTicketsRow struct {
	ID         string
	Name       string
	SeatNumber int64
	BookedAt   sql.NullTime
}

func (q *Queries) GetUserTickets(ctx context.Context, userID string) ([]GetUserTicketsRow, error) {
	rows, err := q.db.QueryContext(ctx, getUserTickets, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetUserTicketsRow
	for rows.Next() {
		var i GetUserTicketsRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.SeatNumber,
			&i.BookedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
