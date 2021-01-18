package main

import (
	"bytes"
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

    var m map[string]interface{}
    json.Unmarshal(response.Body.Bytes(), &m)
    
    if (m["error"] != "Unauthorized") {
        t.Errorf("Expected Unauthorized error, got %s", m["error"])
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

func TestGETFeedbackSuccess(t *testing.T) {
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

func TestPOSTFeedbackSuccess(t *testing.T) {
    clearTable()

    var jsonStr = []byte(`{"feedback":"test new feedback", "ip_address": "2.2.2.2"}`)

    req, _ := http.NewRequest("POST", "/feedback", bytes.NewBuffer(jsonStr))
    response := executeRequest(req, false)

    checkResponseCode(t, http.StatusCreated, response.Code)
    var body Feedback
    err := json.Unmarshal(response.Body.Bytes(), &body)
    if (err != nil) {
        t.Errorf("Expected feedback to be 'test new feedback'. Got %s", response.Body.String())
    }
    if (body.Feedback != "test new feedback") {
        t.Errorf("Expected feedback to be 'test new feedback'. Got %s", body.Feedback)
    }
    if (body.IPAddress != "2.2.2.2") {
        t.Errorf("Expected ip_address to be '2.2.2.2'. Got %s", body.IPAddress)
    }
}

func TestPOSTFeedbackBadPayload(t *testing.T) {
    clearTable()

    var jsonStr = []byte(`{}`)

    req, _ := http.NewRequest("POST", "/feedback", bytes.NewBuffer(jsonStr))
    response := executeRequest(req, false)

    checkResponseCode(t, http.StatusBadRequest, response.Code)
    var m map[string]interface{}
    json.Unmarshal(response.Body.Bytes(), &m)
    
    if (m["error"] != "Feedback is required") {
        t.Errorf("Expected error message 'Feedback is required', got %s", m["error"])
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