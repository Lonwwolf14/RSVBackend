package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"rsvbackend/internal/app"
	"rsvbackend/internal/config"
	"rsvbackend/internal/database"
	"rsvbackend/internal/handlers"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var store = sessions.NewCookieStore([]byte("super-secret-key"))

func initDB(DBURL string) (*sql.DB, error) {
	DB, err := sql.Open("postgres", DBURL)
	if err != nil {
		return nil, err
	}
	return DB, nil
}

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

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using default port 8080")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	config, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	DB, err := initDB(config.DBURL)
	if err != nil {
		log.Fatal(err)
	}
	defer DB.Close()

	queries := database.New(DB)
	templates, err := template.ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}

	appState := &app.AppState{
		AppConfig: config,
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

	fmt.Printf("Server running on port %s\n", port)
	err = http.ListenAndServe(":"+port, router)
	if err != nil {
		log.Print(err)
	}
}

func wrapHandler(appState *app.AppState, handlerFunc func(*app.AppState, http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handlerFunc(appState, w, r)
	}
}
