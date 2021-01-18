package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

var a App


func TestMain(m *testing.M) {
    readEnv()
    a.Initialize()
    ensureTableExists()
    code := m.Run()
    // clearTable()
    os.Exit(code)
}

func ensureTableExists() {
    if _, err := a.DB.Exec(tableCreationQuery); err != nil {
        log.Fatal(err)
    }
}

func clearTable() {
    a.DB.Exec("DELETE FROM feedbacks")
    a.DB.Exec("ALTER SEQUENCE feedbacks_id_seq RESTART WITH 1")
}

func TestUnauthorized(t *testing.T) {
    clearTable()

    req, _ := http.NewRequest("GET", "/feedback", nil)
    response := executeRequest(req, false)

    checkResponseCode(t, http.StatusUnauthorized, response.Code)

    if body := response.Body.String(); body != "Unauthorized" {
        t.Errorf("Expected Unauthorized. Got %s", body)
    }
}

func TestEmptyTable(t *testing.T) {
    clearTable()

    req, _ := http.NewRequest("GET", "/feedback", nil)
    response := executeRequest(req, true)

    checkResponseCode(t, http.StatusOK, response.Code)
    var body []Feedback
    json.Unmarshal(response.Body.Bytes(), &body)
    if (len(body) != 0) {
        t.Errorf("Expected an empty array. Got %s", response.Body.String())
    }
}

func TestGETSuccess(t *testing.T) {
    clearTable()
    addFeedback(1)

    req, _ := http.NewRequest("GET", "/feedback", nil)
    response := executeRequest(req, true)

    checkResponseCode(t, http.StatusOK, response.Code)
    var body []Feedback
    err := json.Unmarshal(response.Body.Bytes(), &body)
    if (err != nil) {
        t.Errorf("Expected an array of feedback items. Got %s", response.Body.String())
    }
    if (len(body) != 1) {
        t.Errorf("Expected an array of feedback items. Got %s", response.Body.String())
    }
}

func executeRequest(req *http.Request, authorized bool) *httptest.ResponseRecorder {
    if (authorized) {
        secret, _ := os.LookupEnv("SECRET");
        req.Header.Add("Authorization", "Bearer " + secret)
    }
    rr := httptest.NewRecorder()
    a.Router.ServeHTTP(rr, req)

    return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
    if expected != actual {
        t.Errorf("Expected response code %d. Got %d\n", expected, actual)
    }
}

func addFeedback(count int) {
    if count < 1 {
        count = 1
    }

    for i := 0; i < count; i++ {
        newFeedback := Feedback{Feedback: "Feedback " + strconv.Itoa(i), IPAddress: "1.1.1.1"}
        err := a.DB.Insert(&newFeedback)
        if (err != nil) {
            log.Fatal(err)
        }
    }
}


const tableCreationQuery = `CREATE TABLE IF NOT EXISTS feedbacks (
    id serial PRIMARY KEY,
    feedback VARCHAR(1024),
    ip_address VARCHAR(50),
    created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
  )`