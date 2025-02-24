package app

import (
	"rsvbackend/internal/config"
	"rsvbackend/internal/database"
)

type AppState struct {
	AppConfig *config.Config
	DB        *database.Queries
}
