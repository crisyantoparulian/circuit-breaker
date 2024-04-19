package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

var (
	counter = 0
)

func main() {

	// Register the handler function for the /hello route
	http.HandleFunc("/data", handler)

	// Start the HTTP server on port 8080
	fmt.Println("Server listening on port 8081")
	http.ListenAndServe(":8081", nil)

}
func handler(w http.ResponseWriter, r *http.Request) {
	counter++
	defer fmt.Println("TOTAL REQUEST == ", counter)
	data := Response{
		Success: true,
		Message: "success",
	}

	if counter > 30 {
		data.Success = false
		data.Message = "failed"
	}
	// time.Sleep(7 * time.Second)

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
