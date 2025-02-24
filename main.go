package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"rsvbackend/internal/app"
	"rsvbackend/internal/config"
	"rsvbackend/internal/database"
	"rsvbackend/internal/handlers"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func initDB(DBURL string) (*sql.DB, error) {
	DB, err := sql.Open("postgres", DBURL)
	if err != nil {
		return nil, err
	}
	return DB, nil
}

func main() {

	//Getting port from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file, using default port 8080")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	//Getting the config data
	config, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	//Database
	DB, err := initDB(config.DBURL)
	if err != nil {
		log.Fatal(err)
	}
	defer DB.Close()

	//Initialising queries
	queries := database.New(DB)
	appState := &app.AppState{
		AppConfig: config,
		DB:        queries,
	}

	//Mux
	router := mux.NewRouter()

	router.HandleFunc("/", wrapHandler(appState, handlers.HandleHome))
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
