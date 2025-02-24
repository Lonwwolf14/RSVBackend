package main

import (
	"net/http"
	"net/http/httptest"
	"rsvbackend/internal/app"
	"testing"
)

func TestAuthMiddleware(t *testing.T) {
	appState := &app.AppState{
		Store: store,
	}

	// Handler to test middleware
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Protected content"))
	})

	// Test without session (should redirect)
	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler := AuthMiddleware(testHandler)
	handler.ServeHTTP(rr, req)
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
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
	if rr.Body.String() != "Protected content" {
		t.Errorf("expected 'Protected content', got %s", rr.Body.String())
	}
}
