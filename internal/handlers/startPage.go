package handlers

import (
	"net/http"
	"rsvbackend/internal/app"
)

func HandleHome(appState *app.AppState, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World"))
}
