-- name: CreateTicket :one
INSERT INTO tickets (id, train_id, user_id, seat_number)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: DeleteTicket :exec
DELETE FROM tickets
WHERE id = $1 AND user_id = $2;

-- name: GetAvailableTickets :many
SELECT t.id, t.name, t.total_seats, 
       (t.total_seats - COUNT(tk.id)) AS available_seats
FROM trains t
LEFT JOIN tickets tk ON t.id = tk.train_id
GROUP BY t.id, t.name, t.total_seats;

-- name: GetUserTickets :many
SELECT tk.id, t.name, tk.seat_number, tk.booked_at
FROM tickets tk
JOIN trains t ON tk.train_id = t.id
WHERE tk.user_id = $1;