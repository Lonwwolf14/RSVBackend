-- name: CreateTicket :one
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
RETURNING id, train_id, user_id, seat_number, booked_at;

-- name: DeleteTicket :exec
DELETE FROM tickets
WHERE id = ? AND user_id = ?;

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
WHERE tk.user_id = ?;