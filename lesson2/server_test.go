package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var server *httptest.Server
var mux *http.ServeMux

func WithTestSetup(t *testing.T, testFunction func()) {
	// Create a temporary log file for testing
	mux = http.NewServeMux()

	mux.HandleFunc("/replace", handleReplace("test_db.log"))
	mux.HandleFunc("/get", handleGet("test_db.log"))

	server = httptest.NewServer(mux)
	defer server.Close()
	defer os.Remove("test_db.log")

	testFunction()
}

func TestServerGetBeforeReplace(t *testing.T) {
	WithTestSetup(t, func() {

		// Perform a "get" operation before any "replace" operation
		resp, err := http.Get(server.URL + "/get")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		// Check if the response status code is 200 OK
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, but got %d",
				http.StatusOK, resp.StatusCode)
		}

		// Read the response body and compare it with an empty string
		responseData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		if string(responseData) != "" {
			t.Errorf("Expected an empty response body, but got '%s'",
				responseData)
		}

	})
}

func TestServerBasicFlow(t *testing.T) {
	WithTestSetup(t, func() {

		// Perform a "replace" operation
		payload := []byte("Test data for basic flow")
		resp, err := http.Post(server.URL+"/replace",
			"application/octet-stream", bytes.NewReader(payload))
		if err != nil {
			t.Fatal(err)
		}

		// Check if the response status code is 200 OK
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, but got %d",
				http.StatusOK, resp.StatusCode)
		}

		// Perform a "get" operation
		resp, err = http.Get(server.URL + "/get")
		if err != nil {
			t.Fatal(err)
		}

		// Check if the response status code is 200 OK
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, but got %d",
				http.StatusOK, resp.StatusCode)
		}

		defer resp.Body.Close()

		// Read the response body and compare it with the expected data
		responseData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		expectedData := string(payload)
		if string(responseData) != expectedData {
			t.Errorf("Expected response body '%s', but got '%s'",
				expectedData, responseData)
		}

	})
}

func TestServerGetAfterDisconnect(t *testing.T) {
	WithTestSetup(t, func() {

		// Perform a "replace" operation
		payload := []byte("Test data for get after disconnect")
		resp, err := http.Post(server.URL+"/replace",
			"application/octet-stream", bytes.NewReader(payload))
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		// Disconnect the server
		server.Close()

		// Reconnect
		server = httptest.NewServer(mux)

		// Attempt a "get" operation after the server has been disconnected
		resp, err = http.Get(server.URL + "/get")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()

		// Check if the response status code is still 200 OK
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code %d, but got %d",
				http.StatusOK, resp.StatusCode)
		}

		// Read the response body and compare it with the expected data
		responseData, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		expectedData := string(payload)
		if string(responseData) != expectedData {
			t.Errorf("Expected response body '%s', but got '%s'",
				expectedData, responseData)
		}

	})
}
