package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"goscraper/src/handlers"
	"goscraper/src/helpers/databases"
	"goscraper/src/types"
	"goscraper/src/utils"
)

// Vercel requires an exported function to handle requests
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.URL.Path {
	case "/hello":
		json.NewEncoder(w).Encode(map[string]string{"message": "Hello, World!"})
	case "/login":
		handleLogin(w, r)
	case "/user":
		handleUser(w, r)
	default:
		http.Error(w, "404 Not Found", http.StatusNotFound)
	}
}

// Login handler
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

// User data handler
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

// Fetch all data function
func fetchAllData(token string) (map[string]interface{}, error) {
	type result struct {
		key  string
		data interface{}
		err  error
	}

	resultChan := make(chan result, 5)

	go func() { data, err := handlers.GetUser(token); resultChan <- result{"user", data, err} }()
	go func() { data, err := handlers.GetAttendance(token); resultChan <- result{"attendance", data, err} }()
	go func() { data, err := handlers.GetMarks(token); resultChan <- result{"marks", data, err} }()
	go func() { data, err := handlers.GetCourses(token); resultChan <- result{"courses", data, err} }()
	go func() { data, err := handlers.GetTimetable(token); resultChan <- result{"timetable", data, err} }()

	data := make(map[string]interface{})
	for i := 0; i < 5; i++ {
		r := <-resultChan
		if r.err != nil {
			return nil, r.err
		}
		data[r.key] = r.data
	}

	if user, ok := data["user"].(*types.User); ok {
		data["regNumber"] = user.RegNumber
	}

	return data, nil
}
