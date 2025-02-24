package app

import (
	"html/template"
	"rsvbackend/internal/config"
	"rsvbackend/internal/database"

	"github.com/gorilla/sessions"
)

// AppState holds the application-wide state and dependencies
type AppState struct {
	AppConfig *config.Config            // Application configuration
	DB        database.QueriesInterface // Database queries
	Store     *sessions.CookieStore     // Session store for authentication
	Templates *template.Template        // Loaded templates
}
