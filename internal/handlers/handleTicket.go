package handlers

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"rsvbackend/internal/app"
	"rsvbackend/internal/database"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// HandleBookTicket handles ticket booking requests, enforcing distributed mutual exclusion
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

	// Check if any node is in critical section
	appState.Node.Mutex.Lock()
	if appState.Node.AnyCS && !appState.Node.InCS {
		// Add to queue and wait
		appState.Node.Clock++
		req := app.Request{NodeID: appState.Node.ID, Timestamp: appState.Node.Clock, UserID: userIDStr}
		appState.Node.Requests = append(appState.Node.Requests, req)
		appState.Node.Mutex.Unlock()
		http.Redirect(w, r, "/", http.StatusSeeOther) // Redirect back to home
		return
	}
	appState.Node.Mutex.Unlock()

	if r.Method == http.MethodGet {
		if !requestCriticalSection(appState, userIDStr) {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		defer releaseCriticalSection(appState)

		trains, err := appState.DB.GetAvailableTickets(r.Context())
		if err != nil {
			log.Println("Error fetching available tickets:", err)
			http.Error(w, "Failed to fetch available tickets", http.StatusInternalServerError)
			return
		}
		err = appState.Templates.ExecuteTemplate(w, "book.html", map[string]interface{}{
			"Trains": trains,
			"Error":  nil,
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

	if !requestCriticalSection(appState, userIDStr) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	defer releaseCriticalSection(appState)

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

// HandleCancelTicket cancels a user's ticket
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

// HandleViewTickets displays the user's booked tickets
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

// HandleViewAvailableTickets displays all available tickets
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

// requestCriticalSection requests entry into the critical section for booking
func requestCriticalSection(appState *app.AppState, userID string) bool {
	appState.Node.Mutex.Lock()
	defer appState.Node.Mutex.Unlock()

	if appState.Node.AnyCS && !appState.Node.InCS {
		appState.Node.Clock++
		req := app.Request{NodeID: appState.Node.ID, Timestamp: appState.Node.Clock, UserID: userID}
		appState.Node.Requests = append(appState.Node.Requests, req)
		return false
	}

	appState.Node.Clock++
	req := app.Request{NodeID: appState.Node.ID, Timestamp: appState.Node.Clock, UserID: userID}
	appState.Node.InCS = true
	appState.Node.AnyCS = true

	client := &http.Client{Timeout: 5 * time.Second}
	for _, peer := range appState.Node.Peers {
		data, _ := json.Marshal(req)
		go func(peer string) {
			resp, err := client.Post(peer+"/request", "application/json", bytes.NewBuffer(data))
			if err != nil || resp.StatusCode != http.StatusOK {
				log.Printf("Failed to notify %s: %v", peer, err)
			}
		}(peer)
	}
	return true
}

// releaseCriticalSection releases the critical section and processes the queue
func releaseCriticalSection(appState *app.AppState) {
	appState.Node.Mutex.Lock()
	defer appState.Node.Mutex.Unlock()

	appState.Node.InCS = false
	if len(appState.Node.Requests) == 0 {
		appState.Node.AnyCS = false
	} else {
		// Find earliest request (FCFS)
		earliest := appState.Node.Requests[0]
		earliestIdx := 0
		for i, req := range appState.Node.Requests[1:] {
			if req.Timestamp < earliest.Timestamp {
				earliest = req
				earliestIdx = i + 1
			}
		}
		appState.Node.Requests = append(appState.Node.Requests[:earliestIdx], appState.Node.Requests[earliestIdx+1:]...)

		client := &http.Client{Timeout: 5 * time.Second}
		data, _ := json.Marshal(map[string]string{"NodeID": appState.Node.ID, "RequesterID": earliest.NodeID})
		go func(peer string) {
			resp, err := client.Post(peer+"/reply", "application/json", bytes.NewBuffer(data))
			if err != nil || resp.StatusCode != http.StatusOK {
				log.Printf("Failed to reply to %s: %v", peer, err)
			}
		}(earliest.NodeID)
	}

	// Notify peers of release
	client := &http.Client{Timeout: 5 * time.Second}
	for _, peer := range appState.Node.Peers {
		go func(peer string) {
			_, err := client.Post(peer+"/release", "application/json", nil)
			if err != nil {
				log.Printf("Failed to notify %s of release: %v", peer, err)
			}
		}(peer)
	}
}

// HandleRequest handles incoming critical section requests from other nodes
func HandleRequest(appState *app.AppState, w http.ResponseWriter, r *http.Request) {
	var req app.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	appState.Node.Mutex.Lock()
	defer appState.Node.Mutex.Unlock()

	appState.Node.Clock = max(appState.Node.Clock, req.Timestamp) + 1
	if appState.Node.InCS || appState.Node.AnyCS {
		appState.Node.Requests = append(appState.Node.Requests, req)
		w.WriteHeader(http.StatusAccepted)
		return
	}
	appState.Node.AnyCS = true
	w.WriteHeader(http.StatusOK)
}

// HandleReply handles replies from other nodes granting critical section access
func HandleReply(appState *app.AppState, w http.ResponseWriter, r *http.Request) {
	var reply struct {
		NodeID      string `json:"NodeID"`
		RequesterID string `json:"RequesterID"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reply); err != nil {
		http.Error(w, "Invalid reply", http.StatusBadRequest)
		return
	}

	appState.Node.Mutex.Lock()
	defer appState.Node.Mutex.Unlock()

	if reply.RequesterID == appState.Node.ID && !appState.Node.InCS {
		appState.Node.InCS = true
	}
	w.WriteHeader(http.StatusOK)
}

// HandleRelease handles notifications of critical section release from other nodes
func HandleRelease(appState *app.AppState, w http.ResponseWriter, r *http.Request) {
	appState.Node.Mutex.Lock()
	defer appState.Node.Mutex.Unlock()

	if len(appState.Node.Requests) == 0 {
		appState.Node.AnyCS = false
	}
	w.WriteHeader(http.StatusOK)
}

// max returns the maximum of two int64 values
func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}
