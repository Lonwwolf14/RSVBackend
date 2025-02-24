package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"html/template"
	"net/http"
	"net/http/httptest"
	"rsvbackend/internal/app"
	"rsvbackend/internal/auth"
	"rsvbackend/internal/database"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
)

// MockQueries implements database.QueriesInterface
type MockQueries struct{}

func (m *MockQueries) CreateUser(ctx context.Context, params database.CreateUserParams) (database.User, error) {
	return database.User{
		ID:       params.ID,
		Email:    params.Email,
		Password: params.Password,
	}, nil
}

func (m *MockQueries) GetUserByEmail(ctx context.Context, email string) (database.User, error) {
	if email == "test@example.com" {
		hashed, _ := auth.HashPassword("password123")
		return database.User{
			ID:       uuid.New(),
			Email:    email,
			Password: hashed,
		}, nil
	}
	return database.User{}, sql.ErrNoRows
}

func setupAppState(t *testing.T) *app.AppState {
	// Mock templates for testing
	tmpl := template.New("mock")
	tmpl, err := tmpl.Parse(`{{define "register.html"}}Register{{end}}{{define "login.html"}}Login{{end}}{{define "home.html"}}Home {{.UserID}}{{end}}`)
	if err != nil {
		t.Fatalf("Failed to parse mock templates: %v", err)
	}

	return &app.AppState{
		DB:        &MockQueries{},
		Store:     sessions.NewCookieStore([]byte("super-secret-key")),
		Templates: tmpl,
	}
}

func TestHandleRegister(t *testing.T) {
	appState := setupAppState(t)

	// Test GET request
	req, _ := http.NewRequest("GET", "/register", nil)
	rr := httptest.NewRecorder()
	HandleRegister(appState, rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if rr.Body.String() != "Register" {
		t.Errorf("expected 'Register', got %s", rr.Body.String())
	}

	// Test POST request (successful registration)
	form := bytes.NewBufferString("email=newuser@example.com&password=password123")
	req, _ = http.NewRequest("POST", "/register", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	HandleRegister(appState, rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected redirect (303), got %d", rr.Code)
	}
	if rr.Header().Get("Location") != "/login" {
		t.Errorf("expected redirect to /login, got %s", rr.Header().Get("Location"))
	}
}

func TestHandleLogin(t *testing.T) {
	appState := setupAppState(t)

	// Test GET request
	req, _ := http.NewRequest("GET", "/login", nil)
	rr := httptest.NewRecorder()
	HandleLogin(appState, rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if rr.Body.String() != "Login" {
		t.Errorf("expected 'Login', got %s", rr.Body.String())
	}

	// Test POST request (successful login)
	form := bytes.NewBufferString("email=test@example.com&password=password123")
	req, _ = http.NewRequest("POST", "/login", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	HandleLogin(appState, rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected redirect (303), got %d", rr.Code)
	}
	if rr.Header().Get("Location") != "/" {
		t.Errorf("expected redirect to /, got %s", rr.Header().Get("Location"))
	}

	// Test POST request (wrong password)
	form = bytes.NewBufferString("email=test@example.com&password=wrongpass")
	req, _ = http.NewRequest("POST", "/login", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	HandleLogin(appState, rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200 with error, got %d", rr.Code)
	}
}

func TestHandleHome(t *testing.T) {
	appState := setupAppState(t)

	// Test with no session (should redirect)
	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	HandleHome(appState, rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected redirect (303), got %d", rr.Code)
	}
	if rr.Header().Get("Location") != "/login" {
		t.Errorf("expected redirect to /login, got %s", rr.Header().Get("Location"))
	}

	// Test with valid session
	req, _ = http.NewRequest("GET", "/", nil)
	rr = httptest.NewRecorder()
	session, _ := appState.Store.Get(req, "session-name")
	session.Values["authenticated"] = true
	session.Values["userID"] = "test-user-id"
	session.Save(req, rr)
	HandleHome(appState, rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if rr.Body.String() != "Home test-user-id" {
		t.Errorf("expected 'Home test-user-id', got %s", rr.Body.String())
	}
}

func TestHandleLogout(t *testing.T) {
	appState := setupAppState(t)

	// Test logout with session
	req, _ := http.NewRequest("GET", "/logout", nil)
	rr := httptest.NewRecorder()
	session, _ := appState.Store.Get(req, "session-name")
	session.Values["authenticated"] = true
	session.Values["userID"] = "test-user-id"
	session.Save(req, rr)
	HandleLogout(appState, rr, req)
	if rr.Code != http.StatusSeeOther {
		t.Errorf("expected redirect (303), got %d", rr.Code)
	}
	if rr.Header().Get("Location") != "/login" {
		t.Errorf("expected redirect to /login, got %s", rr.Header().Get("Location"))
	}

	// Verify session cleared
	session, _ = appState.Store.Get(req, "session-name")
	if session.Values["authenticated"] == true {
		t.Error("expected authenticated to be false after logout")
	}
}
