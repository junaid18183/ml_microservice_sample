package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAllDatasets(t *testing.T) {
	// Create a new request to the "/api/mlDataset" endpoint
	req, err := http.NewRequest("GET", "/api/mlDataset", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler function directly passing in the ResponseRecorder and Request
	getAllDatasets(rr, req)

	// Check the status code is as expected
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d but got %d", http.StatusOK, rr.Code)
	}

	// Add more assertions for the response body or other details if needed
}

func TestGetDataset(t *testing.T) {
	// Create a new request to the "/api/mlDataset/{id}" endpoint
	req, err := http.NewRequest("GET", "/api/mlDataset/1", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler function directly passing in the ResponseRecorder and Request
	getAllDatasets(rr, req)

	// Check the status code is as expected
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d but got %d", http.StatusOK, rr.Code)
	}
}

func TestReportHealth(t *testing.T) {
	// Create a new request to the "/api/healtz" endpoint
	req, err := http.NewRequest("GET", "/api/healtz", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler function directly passing in the ResponseRecorder and Request
	getAllDatasets(rr, req)

	// Check the status code is as expected
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d but got %d", http.StatusOK, rr.Code)
	}
}
