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
	if !ok || userIDStr == "" {
		log.Println("User not authenticated")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Println("Invalid user UUID:", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		trains, err := appState.DB.GetAvailableTickets(r.Context())
		if err != nil {
			log.Println("Error fetching available tickets:", err)
			http.Error(w, "Failed to fetch available tickets", http.StatusInternalServerError)
			return
		}
		err = appState.Templates.ExecuteTemplate(w, "book.html", map[string]interface{}{
			"Trains": trains,
			"Error":  nil, // Explicitly set to nil for GET
		})
		if err != nil {
			log.Println("Error rendering template:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err = r.ParseForm()
	if err != nil {
		log.Println("Error parsing form:", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	trainIDStr := r.FormValue("train_id")
	seatNumberStr := r.FormValue("seat_number")

	trainID, err := uuid.Parse(trainIDStr)
	if err != nil {
		log.Println("Invalid train UUID:", err)
		err = appState.Templates.ExecuteTemplate(w, "book.html", map[string]string{
			"Error": "Invalid train ID",
		})
		if err != nil {
			log.Println("Error rendering template:", err)
		}
		return
	}

	seatNumber, err := strconv.Atoi(seatNumberStr)
	if err != nil || seatNumber < 1 {
		log.Println("Invalid seat number:", seatNumberStr)
		err = appState.Templates.ExecuteTemplate(w, "book.html", map[string]string{
			"Error": "Invalid seat number",
		})
		if err != nil {
			log.Println("Error rendering template:", err)
		}
		return
	}

	ticketID := uuid.New()
	_, err = appState.DB.CreateTicket(r.Context(), database.CreateTicketParams{
		ID:         ticketID.String(),
		TrainID:    trainID.String(),
		UserID:     userID.String(),
		SeatNumber: int64(seatNumber),
	})
	if err != nil {
		log.Println("Error booking ticket:", err)
		err = appState.Templates.ExecuteTemplate(w, "book.html", map[string]string{
			"Error": "Seat already booked or invalid",
		})
		if err != nil {
			log.Println("Error rendering template:", err)
		}
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
	if !ok || userIDStr == "" {
		log.Println("User not authenticated")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Println("Invalid user UUID:", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	ticketIDStr := r.URL.Query().Get("ticket_id")
	ticketID, err := uuid.Parse(ticketIDStr)
	if err != nil {
		log.Println("Invalid ticket UUID:", err)
		http.Redirect(w, r, "/tickets", http.StatusSeeOther)
		return
	}

	err = appState.DB.DeleteTicket(r.Context(), database.DeleteTicketParams{
		ID:     ticketID.String(),
		UserID: userID.String(),
	})
	if err != nil {
		log.Println("Error cancelling ticket:", err)
		http.Error(w, "Failed to cancel ticket", http.StatusInternalServerError)
		return
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
	if !ok || userIDStr == "" {
		log.Println("User not authenticated")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Println("Invalid user UUID:", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tickets, err := appState.DB.GetUserTickets(r.Context(), userID.String())
	if err != nil {
		log.Println("Error fetching user tickets:", err)
		http.Error(w, "Failed to fetch user tickets", http.StatusInternalServerError)
		return
	}

	err = appState.Templates.ExecuteTemplate(w, "tickets.html", map[string]interface{}{
		"Tickets": tickets,
	})
	if err != nil {
		log.Println("Error rendering template:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func HandleViewAvailableTickets(appState *app.AppState, w http.ResponseWriter, r *http.Request) {
	availableTickets, err := appState.DB.GetAvailableTickets(r.Context())
	if err != nil {
		log.Println("Error fetching available tickets:", err)
		http.Error(w, "Failed to fetch available tickets", http.StatusInternalServerError)
		return
	}
	log.Printf("Fetched available tickets: %+v", availableTickets) // Debug log

	err = appState.Templates.ExecuteTemplate(w, "available.html", map[string]interface{}{
		"Tickets": availableTickets,
	})
	if err != nil {
		log.Println("Error rendering available tickets template:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
