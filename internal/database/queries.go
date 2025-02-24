package database

import (
	"context"

	"github.com/google/uuid"
)

type QueriesInterface interface {
	CreateUser(ctx context.Context, params CreateUserParams) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetAvailableTickets(ctx context.Context) ([]GetAvailableTicketsRow, error)   // Match sqlc return type
	CreateTicket(ctx context.Context, params CreateTicketParams) (Ticket, error) // Match sqlc return type
	DeleteTicket(ctx context.Context, params DeleteTicketParams) error
	GetUserTickets(ctx context.Context, userID uuid.UUID) ([]GetUserTicketsRow, error) // Match sqlc return type
}
