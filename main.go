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

	"github.com/joho/godotenv"
)

type feedback struct {
	ID        int64  `json:"id"`
	Feedback  string `json:"feedback"`
	IPAddress string `json:"ip_address"`
}

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func getFeedbackItems(w http.ResponseWriter, r *http.Request) {
	db := connectDB()

	defer db.Close()

	var feedbackData []feedback
	err := db.Model(&feedbackData).Select()
	if err != nil {
		panic(err)
	}
	json.NewEncoder(w).Encode(feedbackData)
}

func connectDB() pg.DB {
	password, existsPassword := os.LookupEnv("PG_PASSWORD")
	database, existsDatabase := os.LookupEnv("PG_DATABASE")
	user, existsUser := os.LookupEnv("PG_USER")
	address, existsAddress := os.LookupEnv("PG_ADDRESS")

	if !existsPassword || !existsDatabase || !existsUser || !existsAddress {
		panic("Database credentials not found")
	}

	db := pg.Connect(&pg.Options{
		User:     user,
		Database: database,
		Password: password,
		Addr:     address,
	})

	return *db
}

func createFeedback(w http.ResponseWriter, r *http.Request) {
	db := connectDB()
	defer db.Close()

	var newFeedback feedback
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "API Error: incorrect data format")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(reqBody, &newFeedback)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if newFeedback.Feedback == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = db.Insert(&feedback{
		Feedback:  newFeedback.Feedback,
		IPAddress: r.RemoteAddr,
	})

	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(newFeedback)
}

func optionsRequest(w http.ResponseWriter, r *http.Request) {
	// Just return
}

func main() {
	fmt.Println("Initializing db...")
	fmt.Println("starting server...")
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("*", optionsRequest).Methods("OPTIONS")
	router.HandleFunc("/feedback", getFeedbackItems).Methods("GET")
	router.HandleFunc("/feedback", createFeedback).Methods("POST")
	withHandlers := handlers.CORS(
		handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"}),
		handlers.AllowedMethods([]string{"GET", "POST", "HEAD", "OPTIONS"}),
		handlers.AllowedOrigins([]string{"*"}))(router)
	log.Fatal(http.ListenAndServe(":8080", withHandlers))
}
