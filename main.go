package main

import (
	"bitcoin-sidechain/cryptoUtils"
	"encoding/json"
	"net/http"
	"path/filepath"
)

func main() {

	// Front End Pages
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/sendTransaction", sendTransactionHandler)
	http.HandleFunc("/keysGen", keyGenHandler)

	// API Endpoints
	http.HandleFunc("/verifysignature", VerifySignatureHandler)

	// Work In Progress
	http.HandleFunc("/hashData", hashDatabaseHandler)

	if err := http.ListenAndServe(":80", nil); err != nil {
		panic(err)
	}
}
func rootHandler(w http.ResponseWriter, r *http.Request) {

	http.ServeFile(w, r, "pages/index.html")
}

func sendTransactionHandler(w http.ResponseWriter, r *http.Request) {

	http.ServeFile(w, r, "pages/signTransaction.html")
}

func keyGenHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "pages/keyGen.html")
}

func VerifySignatureHandler(w http.ResponseWriter, r *http.Request) {
	type Transaction struct {
		From   string `json:"from"`
		To     string `json:"to"`
		Amount string `json:"amount"`
		Nonce  string `json:"nonce"`
	}

	type KeySignRequest struct {
		Signature   string      `json:"signature"`
		Transaction Transaction `json:"transaction"`
	}

	// Decode the incoming JSON request
	var req KeySignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close() // Close the request body

	// Convert the Transaction back to a JSON string
	transactionJSON, err := json.Marshal(req.Transaction)
	if err != nil {
		http.Error(w, "Error converting transaction to JSON", http.StatusInternalServerError)
		return
	}

	// Assuming cryptoUtils.VerifySignature is your verification function
	publicKey := req.Transaction.From // Use the "from" field as the public key
	signature := req.Signature
	message := string(transactionJSON) // The JSON string of the transaction

	// Check if the signature is valid
	verificationResult, _ := cryptoUtils.VerifySignature(signature, publicKey, message)

	// Create the response object
	response := map[string]string{}
	if verificationResult == "valid" {
		response["message"] = "Valid"
	} else {
		response["message"] = "Invalid"
	}

	// Set the content type and encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func hashDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	dbFilename := filepath.Join(".", "nodeList1.db")
	hash := computeDatabaseHash(dbFilename)
	response := map[string]string{"hash": hash}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
