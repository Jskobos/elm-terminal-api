package main

import (
	"fmt"
	"os"

	"github.com/go-pg/pg/v9"
)

// Feedback represents a single piece of user feedback
type Feedback struct {
	ID        int64  `json:"id"`
	Feedback  string `json:"feedback"`
	IPAddress string `json:"ip_address"`
	Created   string `json:"created"`
}

// Book represents a single book
type Book struct {
	ID        int64  `json:"id"`
	Title  		string `json:"title"`
	Author 		string `json:"author"`
	YearRead	int64  `json:"year_read"`
	Created   string `json:"created"`			
}

// ConnectDB uses environment variables to initialize a
// connection to the database.
func ConnectDB() (*pg.DB)  {
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

	return db
}

func getFeedbackItems(db *pg.DB) ([]Feedback, error) {
	var feedbackData []Feedback
	err := db.Model(&feedbackData).Select()
	if (err != nil) { 
		fmt.Println(err)
		return nil, err
	}
	if (feedbackData == nil) {
		return make([]Feedback, 0), nil
	}
	return feedbackData, nil
}

func (f *Feedback) createFeedbackItem(db *pg.DB) (error) {
	err := db.Insert(f)
	if (err != nil) {
		return err
	}
	return nil
}

func getBooks(db *pg.DB) ([]Book, error) {
	var bookData []Book
	err := db.Model(&bookData).Select()
	if (err != nil) { 
		return nil, err
	}
	if (bookData == nil) {
		return make([]Book, 0), nil
	}
	return bookData, nil
}

