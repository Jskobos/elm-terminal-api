package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type feedback struct {
	ID       int64  `json:"id"`
	Feedback string `json:"feedback"`
}

type allFeedback []feedback

var feedbackData = allFeedback{
	{
		ID:       1,
		Feedback: "This is some sample feedback",
	},
}

func getFeedbackItems(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(feedbackData)
}

func createFeedback(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
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

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func main() {
	fmt.Println("starting server...")
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/feedback", getFeedbackItems).Methods("GET")
	router.HandleFunc("/feedback", createFeedback).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}
