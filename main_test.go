package main

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func Test_searchHandler(t *testing.T) {
	router := mux.NewRouter()
	router.HandleFunc("/search", searchHandler).Methods("POST")
	server := httptest.NewServer(router)
	defer server.Close()

	words := []string{"golang", "concurrency"}
	expected := WordCountMap{
		"golang": {
			Word:          "golang",
			TF:            8,
			DF:            2,
			LastTF:        8,
			LastDF:        2,
			TotalSearches: 1,
		},
		"concurrency": {
			Word:          "concurrency",
			TF:            3,
			DF:            1,
			LastTF:        3,
			LastDF:        1,
			TotalSearches: 1,
		},
	}

	reqBody, err := json.Marshal(words)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	req, err := http.NewRequest("POST", server.URL+"/search", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var actual WordCountMap
	if err := json.Unmarshal(w.Body.Bytes(), &actual); err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}
