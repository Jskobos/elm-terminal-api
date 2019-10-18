package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"
)

type feedback struct {
	ID        int64  `json:"id"`
	Feedback  string `json:"feedback"`
	IPAddress string `json:"ip_address"`
}

type allFeedback []feedback

var feedbackData = allFeedback{
	{
		ID:       1,
		Feedback: "This is some sample feedback",
	},
}

func createSchema(db *pg.DB) error {
	for _, model := range []interface{}{(*feedback)(nil)} {
		err := db.CreateTable(model, &orm.CreateTableOptions{
			Temp: true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func dbConnect() {
	db := pg.Connect(&pg.Options{
		User: "postgres",
	})
	defer db.Close()

	err := createSchema(db)
	if err != nil {
		panic(err)
	}
}

func getFeedbackItems(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(feedbackData)
}

func createFeedback(w http.ResponseWriter, r *http.Request) {
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
	feedbackData = append(feedbackData, newFeedback)
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(newFeedback)
}

func optionsRequest(w http.ResponseWriter, r *http.Request) {
	// Just return
}

func main() {
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
