package handlers

import (
	"log"
	"net/http"
	"rsvbackend/internal/app"
	"rsvbackend/internal/auth"
	"rsvbackend/internal/database"

	"github.com/google/uuid"
)

func HandleRegister(appState *app.AppState, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		appState.Templates.ExecuteTemplate(w, "register.html", nil)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	if email == "" || password == "" {
		appState.Templates.ExecuteTemplate(w, "register.html", map[string]string{"Error": "Email and password are required"})
		return
	}

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		log.Println("Error hashing password:", err)
		appState.Templates.ExecuteTemplate(w, "register.html", map[string]string{"Error": "Internal server error"})
		return
	}

	// Generate UUID as string to match sqlc's CreateUserParams
	userID := uuid.New().String() // Convert to string immediately
	log.Printf("Creating user: id=%s, email=%s, password=%s", userID, email, hashedPassword)
	params := database.CreateUserParams{
		ID:       userID, // Now a string
		Email:    email,
		Password: hashedPassword,
	}

	_, err = appState.DB.CreateUser(r.Context(), params)
	if err != nil {
		// Check if it's a duplicate email error (SQLite error code 19 or similar)
		if err.Error() == "UNIQUE constraint failed: users.email" { // Adjust based on actual error string
			appState.Templates.ExecuteTemplate(w, "register.html", map[string]string{"Error": "Email already exists"})
		} else {
			log.Printf("Error creating user: params=%+v, err=%v", params, err)
			appState.Templates.ExecuteTemplate(w, "register.html", map[string]string{"Error": "Server error"})
		}
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func HandleLogin(appState *app.AppState, w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		appState.Templates.ExecuteTemplate(w, "login.html", nil)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	user, err := appState.DB.GetUserByEmail(r.Context(), email)
	if err != nil {
		log.Println("Error fetching user:", err)
		appState.Templates.ExecuteTemplate(w, "login.html", map[string]string{"Error": "Invalid email or password"})
		return
	}

	err = auth.CheckPassword(user.Password, password)
	if err != nil {
		appState.Templates.ExecuteTemplate(w, "login.html", map[string]string{"Error": "Invalid email or password"})
		return
	}

	session, err := appState.Store.Get(r, "session-name")
	if err != nil {
		log.Println("Error getting session:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	session.Values["authenticated"] = true
	session.Values["userID"] = user.ID // ID is a string from sqlc
	log.Printf("Setting session: authenticated=%v, userID=%s", session.Values["authenticated"], user.ID)
	err = session.Save(r, w)
	if err != nil {
		log.Println("Error saving session:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func HandleHome(appState *app.AppState, w http.ResponseWriter, r *http.Request) {
	session, err := appState.Store.Get(r, "session-name")
	if err != nil {
		log.Println("Error getting session:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	userID, ok := session.Values["userID"].(string)
	if !ok || userID == "" {
		log.Println("userID not found in session or not a string")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	appState.Node.Mutex.Lock()
	bookingBlocked := appState.Node.AnyCS
	appState.Node.Mutex.Unlock()

	err = appState.Templates.ExecuteTemplate(w, "home.html", map[string]interface{}{
		"UserID":         userID,
		"BookingBlocked": bookingBlocked,
	})
	if err != nil {
		log.Println("Error rendering home template:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func HandleLogout(appState *app.AppState, w http.ResponseWriter, r *http.Request) {
	session, err := appState.Store.Get(r, "session-name")
	if err != nil {
		log.Println("Error getting session:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	session.Values["authenticated"] = false
	delete(session.Values, "userID")
	err = session.Save(r, w)
	if err != nil {
		log.Println("Error saving session:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
