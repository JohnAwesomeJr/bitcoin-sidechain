package main

import (
	"encoding/json"
	"net/http"
)

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/greet", greetHandler) // New endpoint

	if err := http.ListenAndServe(":80", nil); err != nil {
		panic(err)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"ip": "192.168.1.1"}
	json.NewEncoder(w).Encode(response)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"message": "hello, world"}
	json.NewEncoder(w).Encode(response)
}

// New handler for greeting with a query parameter
func greetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	name := r.URL.Query().Get("name") // Get the 'name' query parameter

	if name == "" {
		name = "stranger" // Default value if 'name' is not provided
	}

	response := map[string]string{"greeting": "Hello, " + name + "!"}
	json.NewEncoder(w).Encode(response)
}
