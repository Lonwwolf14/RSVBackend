package database

import (
	"context"
)

// QueriesInterface defines the required database methods.
type QueriesInterface interface {
	CreateUser(ctx context.Context, params CreateUserParams) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetAvailableTickets(ctx context.Context) ([]GetAvailableTicketsRow, error)
	CreateTicket(ctx context.Context, params CreateTicketParams) (Ticket, error)
	DeleteTicket(ctx context.Context, params DeleteTicketParams) error // Updated to use DeleteTicketParams
	GetUserTickets(ctx context.Context, userID string) ([]GetUserTicketsRow, error)
}
