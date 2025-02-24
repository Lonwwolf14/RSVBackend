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

	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		log.Println("Error hashing password:", err)
		appState.Templates.ExecuteTemplate(w, "register.html", map[string]string{"Error": "Internal server error"})
		return
	}

	userID := uuid.New()
	_, err = appState.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:       userID,
		Email:    email,
		Password: hashedPassword,
	})
	if err != nil {
		log.Println("Error creating user:", err)
		appState.Templates.ExecuteTemplate(w, "register.html", map[string]string{"Error": "Email already exists"})
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
	session.Values["userID"] = user.ID.String()
	log.Printf("Setting session: authenticated=%v, userID=%v", session.Values["authenticated"], session.Values["userID"])
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

	var userID string
	if id, ok := session.Values["userID"].(string); ok {
		userID = id
	} else {
		log.Println("userID not found in session or not a string")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	err = appState.Templates.ExecuteTemplate(w, "home.html", map[string]string{"UserID": userID})
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
