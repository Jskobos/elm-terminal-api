package main

import (
	"log"
	"os"
	"testing"
)

var a App


func TestMain(m *testing.M) {
    a.Initialize()

    ensureTableExists()
    code := m.Run()
    clearTable()
    os.Exit(code)
}

func ensureTableExists() {
    if _, err := a.DB.Exec(tableCreationQuery); err != nil {
        log.Fatal(err)
    }
}

func clearTable() {
    a.DB.Exec("DELETE FROM feedbacks")
}

const tableCreationQuery = `CREATE TABLE feedbacks (
    id serial PRIMARY KEY,
    feedback VARCHAR(1024),
    ip_address VARCHAR(50),
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
  )`