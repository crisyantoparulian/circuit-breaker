package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

var (
	counter = 0
	mtx     sync.Mutex
)

func main() {

	http.HandleFunc("/data", handler)

	// Start the HTTP server on port 8080
	fmt.Println("Server listening on port 8081")
	http.ListenAndServe(":8081", nil)

}
func handler(w http.ResponseWriter, r *http.Request) {
	mtx.Lock()
	counter++
	mtx.Unlock()
	defer fmt.Println("TOTAL REQUEST == ", counter)
	data := Response{
		Success: true,
		Message: "success",
	}

	if counter > 1 {
		data.Success = false
		data.Message = "failed"
	}

	responseJson(w, data)
}

func responseJson(w http.ResponseWriter, data Response) {
	// Encode the data as JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if data.Success {
		// Set the Content-Type header to application/json
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Write the JSON response to the client
		w.Write(jsonData)
	} else {
		// Set the Content-Type header to application/json
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		// Write the JSON response to the client
		w.Write(jsonData)
	}
}
