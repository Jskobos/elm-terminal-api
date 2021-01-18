package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/go-pg/pg/v9"
)

// App interface for the full application
type App struct {
	Router *mux.Router
	DB *pg.DB
}

func optionsRequest(w http.ResponseWriter, r *http.Request) {
	// Just return
}

func (a *App) getFeedbackItems(w http.ResponseWriter, r *http.Request) {
	// This endpoint is not public
	// @todo: better/more flexible auth
	secret, existsSecret := os.LookupEnv("SECRET");
	authHeader := r.Header["Authorization"];
	if (!existsSecret || len(authHeader) < 1 || authHeader[0] != "Bearer " + secret) {
		fmt.Printf("API Error: Unauthorized\n")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))		
		return
	}

	feedbackData, err := getFeedbackItems(a.DB)
	if err != nil {
		respondWithError(w, r, http.StatusServiceUnavailable, "An unexpected error occurred")
		return
	}
	respondWithJSON(w, r, http.StatusOK, feedbackData)
}

func (a *App) createFeedback(w http.ResponseWriter, r *http.Request) {
	var newFeedback Feedback
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, r, http.StatusBadRequest, "Incorrect data format")
		return
	}

	err = json.Unmarshal(reqBody, &newFeedback)
	if err != nil {
		respondWithError(w, r, http.StatusUnprocessableEntity, "JSON parse error")
		return
	}

	if newFeedback.Feedback == "" {
		respondWithError(w, r, http.StatusBadRequest, "Feedback is required")
		return
	}

	if len(newFeedback.Feedback) > 1000 {
		respondWithError(w, r, http.StatusBadRequest, "Feedback text must be 1000 characters or fewer")
		return
	}

	feedback := Feedback{
		Feedback:  newFeedback.Feedback,
		IPAddress: r.RemoteAddr,
	}
	err = feedback.createFeedbackItem(a.DB)
		
	if err != nil {
		respondWithError(w, r, http.StatusServiceUnavailable, "Database write error")
		return
	}

	respondWithJSON(w, r, http.StatusCreated, newFeedback)
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("*", optionsRequest).Methods("OPTIONS")
	a.Router.HandleFunc("/feedback", a.getFeedbackItems).Methods("GET")
	a.Router.HandleFunc("/feedback", a.createFeedback).Methods("POST")
	
}

func respondWithError(w http.ResponseWriter, r *http.Request, code int, message string) {
	respondWithJSON(w, r, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	log.Printf("%s %s %d %s", r.Method, r.RequestURI, code, r.RemoteAddr)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// Initialize the server and routes.
func (a *App) Initialize() {
	fmt.Println("Starting server on port 8080")
	a.Router = mux.NewRouter().StrictSlash(true)
	a.initializeRoutes()
	a.DB = ConnectDB()
}

// Run starts up the server
func (a *App) Run() {
	withHandlers := handlers.CORS(
		handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
		handlers.AllowedMethods([]string{"GET", "POST", "HEAD", "OPTIONS"}),
		handlers.AllowedOrigins([]string{"*"}))(a.Router)
	log.Fatal(http.ListenAndServe(":8080", withHandlers))
}