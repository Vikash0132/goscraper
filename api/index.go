package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"goscraper/src/handlers"
	"goscraper/src/types"
)

// ✅ Vercel requires an exported function
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.URL.Path {
	case "/api/hello": // Always use "/api/" prefix in Vercel
		json.NewEncoder(w).Encode(map[string]string{"message": "Hello, World!"})
	case "/api/login":
		handleLogin(w, r)
	case "/api/user":
		handleUser(w, r)
	default:
		http.Error(w, `{"error": "404 Not Found"}`, http.StatusNotFound)
	}
}

// ✅ Login handler
func handleLogin(w http.ResponseWriter, r *http.Request) {
	var creds struct {
		Username string `json:"account"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil || creds.Username == "" || creds.Password == "" {
		http.Error(w, `{"error": "Invalid JSON or missing credentials"}`, http.StatusBadRequest)
		return
	}

	lf := &handlers.LoginFetcher{}
	session, err := lf.Login(creds.Username, creds.Password)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(session)
}

// ✅ User data handler
func handleUser(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("X-CSRF-Token")
	if token == "" {
		http.Error(w, `{"error": "Missing X-CSRF-Token header"}`, http.StatusUnauthorized)
		return
	}

	user, err := handlers.GetUser(token)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%v"}`, err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(user)
}
