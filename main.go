package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"rsvbackend/internal/app"
	"rsvbackend/internal/database"
	"rsvbackend/internal/handlers"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "session-name")
		if err != nil || session.Values["authenticated"] != true || session.Values["userID"] == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

var store *sessions.CookieStore // Declare as a global variable, initialize in main

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using defaults")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		log.Println("SESSION_KEY not set, using insecure default")
		sessionKey = "31392cf13a7c6b24431a653adb18842cd5230e9a9b3c0ba6cfade6ec072773d8" // Use your .env key as fallback for testing
	}
	store = sessions.NewCookieStore([]byte(sessionKey))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,      // Prevent JavaScript access
		// Secure: true,     // Uncomment in production with HTTPS
	}

	dbURL := os.Getenv("DATABASE_URL")
	var queries *database.Queries
	if dbURL != "" {
		db, err := sql.Open("libsql", dbURL)
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		if err = db.Ping(); err != nil {
			log.Fatalf("Failed to ping database: %v", err)
		}
		queries = database.New(db)
	} else {
		log.Println("DATABASE_URL not set, running without DB endpoints")
	}

	templates, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	appState := &app.AppState{
		DB:        queries,
		Store:     store,
		Templates: templates,
	}

	router := mux.NewRouter()
	router.HandleFunc("/register", wrapHandler(appState, handlers.HandleRegister)).Methods("GET", "POST")
	router.HandleFunc("/login", wrapHandler(appState, handlers.HandleLogin)).Methods("GET", "POST")

	protected := router.PathPrefix("/").Subrouter()
	protected.Use(AuthMiddleware)
	protected.HandleFunc("/", wrapHandler(appState, handlers.HandleHome))
	protected.HandleFunc("/logout", wrapHandler(appState, handlers.HandleLogout)).Methods("GET")
	protected.HandleFunc("/book", wrapHandler(appState, handlers.HandleBookTicket)).Methods("GET", "POST")
	protected.HandleFunc("/cancel", wrapHandler(appState, handlers.HandleCancelTicket)).Methods("GET")
	protected.HandleFunc("/tickets", wrapHandler(appState, handlers.HandleViewTickets)).Methods("GET")
	protected.HandleFunc("/available", wrapHandler(appState, handlers.HandleViewAvailableTickets)).Methods("GET")

	fmt.Printf("Server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}

func wrapHandler(appState *app.AppState, handlerFunc func(*app.AppState, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handlerFunc(appState, w, r)
	}
}
