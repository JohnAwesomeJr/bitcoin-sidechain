package main

import (
	"encoding/json"
	"net/http"
	"path/filepath"
)

func main() {

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/hashData", hashDatabase)

	if err := http.ListenAndServe(":80", nil); err != nil {
		panic(err)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	// Open the index.html file and serve it
	http.ServeFile(w, r, "index.html")
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	// Get the name parameter from the form
	name := r.FormValue("name")

	// Prepare the JSON response
	response := map[string]string{
		"test":                                  "false",
		"name":                                  name,
		"this is the template for json returns": "true",
	}

	// Set the Content-Type header to JSON before writing to ResponseWriter
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func hashDatabase(w http.ResponseWriter, r *http.Request) {
	dbFilename := filepath.Join(".", "nodeList1.db")
	hash := computeDatabaseHash(dbFilename)
	response := map[string]string{"hash": hash}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
