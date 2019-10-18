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
	password, exists := os.LookupEnv("PG_PASSWORD")
	if !exists {
		panic("DB password not found")
	}
	db := pg.Connect(&pg.Options{
		User:     "postgres",
		Database: "postgres",
		Password: password,
		Addr:     "localhost:5432",
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
	}
	err = json.Unmarshal(reqBody, &newFeedback)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(newFeedback.Feedback)
	err = db.Insert(&feedback{
		Feedback:  newFeedback.Feedback,
		IPAddress: newFeedback.IPAddress,
	})

	if err != nil {
		panic(err)
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
