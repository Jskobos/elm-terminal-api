package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

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
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(err.Error()))
		return
	}
	fmt.Printf("GET %s: %s %s\n", r.RequestURI, r.RemoteAddr, time.Now())
	json.NewEncoder(w).Encode(feedbackData)
}

func (a *App) createFeedback(w http.ResponseWriter, r *http.Request) {
	var newFeedback Feedback
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "API Error: incorrect data format")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Incorrect data format"))
		return
	}

	err = json.Unmarshal(reqBody, &newFeedback)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte("JSON parse error"))
		return
	}

	if newFeedback.Feedback == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "API Error: invalid payload")
		w.Write([]byte("feedback is required"))
		return
	}

	if len(newFeedback.Feedback) > 1000 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "API Error: invalid payload")
		w.Write([]byte("Feedback text must be 1000 characters or less"))
		return
	}

	feedback := Feedback{
		Feedback:  newFeedback.Feedback,
		IPAddress: r.RemoteAddr,
	}
	err = feedback.createFeedbackItem(a.DB)
		
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Database write error"))
		return
	}

	fmt.Printf("POST %s: %s %s\n", r.RequestURI, r.RemoteAddr, time.Now())

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(newFeedback)
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("*", optionsRequest).Methods("OPTIONS")
	a.Router.HandleFunc("/feedback", a.getFeedbackItems).Methods("GET")
	a.Router.HandleFunc("/feedback", a.createFeedback).Methods("POST")
	
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