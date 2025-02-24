package handlers

import (
	"log"
	"net/http"
	"rsvbackend/internal/app"
	"rsvbackend/internal/database"
	"strconv"

	"github.com/google/uuid"
)

func HandleBookTicket(appState *app.AppState, w http.ResponseWriter, r *http.Request) {
	session, err := appState.Store.Get(r, "session-name")
	if err != nil {
		log.Println("Error getting session:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	userIDStr, ok := session.Values["userID"].(string)
	if !ok {
		log.Println("userID not found in session")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	userID, _ := uuid.Parse(userIDStr)

	if r.Method == http.MethodGet {
		trains, err := appState.DB.GetAvailableTickets(r.Context())
		if err != nil {
			log.Println("Error fetching available tickets:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		appState.Templates.ExecuteTemplate(w, "book.html", map[string]interface{}{"Trains": trains})
		return
	}

	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	trainIDStr := r.FormValue("train_id")
	seatNumberStr := r.FormValue("seat_number")
	trainID, err := uuid.Parse(trainIDStr)
	if err != nil {
		appState.Templates.ExecuteTemplate(w, "book.html", map[string]string{"Error": "Invalid train ID"})
		return
	}
	seatNumber, err := strconv.Atoi(seatNumberStr)
	if err != nil || seatNumber < 1 {
		appState.Templates.ExecuteTemplate(w, "book.html", map[string]string{"Error": "Invalid seat number"})
		return
	}

	ticketID := uuid.New()
	_, err = appState.DB.CreateTicket(r.Context(), database.CreateTicketParams{
		ID:         ticketID,
		TrainID:    trainID,
		UserID:     userID,
		SeatNumber: int32(seatNumber),
	})
	if err != nil {
		log.Println("Error booking ticket:", err)
		appState.Templates.ExecuteTemplate(w, "book.html", map[string]string{"Error": "Seat already booked or invalid"})
		return
	}

	http.Redirect(w, r, "/tickets", http.StatusSeeOther)
}

func HandleCancelTicket(appState *app.AppState, w http.ResponseWriter, r *http.Request) {
	session, err := appState.Store.Get(r, "session-name")
	if err != nil {
		log.Println("Error getting session:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	userIDStr, ok := session.Values["userID"].(string)
	if !ok {
		log.Println("userID not found in session")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	userID, _ := uuid.Parse(userIDStr)

	ticketIDStr := r.URL.Query().Get("ticket_id")
	ticketID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		http.Redirect(w, r, "/tickets", http.StatusSeeOther)
		return
	}

	err = appState.DB.DeleteTicket(r.Context(), database.DeleteTicketParams{
		ID:     ticketID,
		UserID: userID,
	})
	if err != nil {
		log.Println("Error cancelling ticket:", err)
	}

	http.Redirect(w, r, "/tickets", http.StatusSeeOther)
}

func HandleViewTickets(appState *app.AppState, w http.ResponseWriter, r *http.Request) {
	session, err := appState.Store.Get(r, "session-name")
	if err != nil {
		log.Println("Error getting session:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	userIDStr, ok := session.Values["userID"].(string)
	if !ok {
		log.Println("userID not found in session")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	userID, _ := uuid.Parse(userIDStr)

	tickets, err := appState.DB.GetUserTickets(r.Context(), userID)
	if err != nil {
		log.Println("Error fetching user tickets:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	appState.Templates.ExecuteTemplate(w, "tickets.html", map[string]interface{}{"Tickets": tickets})
}

func HandleViewAvailableTickets(appState *app.AppState, w http.ResponseWriter, r *http.Request) {
	trains, err := appState.DB.GetAvailableTickets(r.Context())
	if err != nil {
		log.Println("Error fetching available tickets:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	appState.Templates.ExecuteTemplate(w, "available.html", map[string]interface{}{"Trains": trains})
}
