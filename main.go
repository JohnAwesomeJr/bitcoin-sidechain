package main

import (
	"bitcoin-sidechain/cryptoUtils"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
)

func main() {

	// Front End Pages
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/sendTransaction", sendTransactionHandler)
	http.HandleFunc("/keysGen", keyGenHandler)
	http.HandleFunc("/balance", walletBalance)

	// API Endpoints
	http.HandleFunc("/verifysignature", VerifySignatureHandler)
	http.HandleFunc("/walletbalance", checkWalletBalance)

	// Work In Progress
	http.HandleFunc("/hashData", hashDatabaseHandler)
	http.HandleFunc("/makewallet", insertNewWallet)
	http.HandleFunc("/shuffleDatabase", shuffleDatabase)

	if err := http.ListenAndServe(":80", nil); err != nil {
		panic(err)
	}
}

// Front End Pages
func rootHandler(w http.ResponseWriter, r *http.Request) {

	http.ServeFile(w, r, "pages/index.html")
}

func sendTransactionHandler(w http.ResponseWriter, r *http.Request) {

	http.ServeFile(w, r, "pages/signTransaction.html")
}

func keyGenHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "pages/keyGen.html")
}

func walletBalance(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "pages/walletBalance.html")
}

// API Endpoints
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
	publicKey := req.Transaction.From
	toAddress := req.Transaction.To
	amount := req.Transaction.Amount
	nonce := req.Transaction.Nonce

	signature := req.Signature
	message := string(transactionJSON) // The JSON string of the transaction

	// Check if the signature is valid
	verificationResult, _ := cryptoUtils.VerifySignature(signature, publicKey, message)

	// make wallet if it doesn't exists
	databaseFile := "nodeList1.db"
	nonceValidation, nonceError := cryptoUtils.CheckNonce(databaseFile, nonce)

	cryptoUtils.NewWallet(toAddress, databaseFile)

	cryptoUtils.MoveSats(publicKey, toAddress, amount, databaseFile)

	// Create the response object
	response := map[string]string{}
	if verificationResult == "valid" && !nonceValidation {
		response["message"] = "Valid"
	} else {
		fmt.Println("Verification Result:", verificationResult)
		fmt.Println("Nonce already used", nonceValidation, "nonceError?", nonceError)
		response["message"] = "Invalid"
	}

	// Set the content type and encode the response as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func checkWalletBalance(w http.ResponseWriter, r *http.Request) {
	// WalletBalanceRequest is a struct to parse the incoming JSON request
	type WalletBalanceRequest struct {
		Wallet string `json:"wallet"`
	}

	// WalletBalanceResponse is a struct to form the JSON response
	type WalletBalanceResponse struct {
		Status  string  `json:"status"`
		Balance float64 `json:"balance,omitempty"`
		Message string  `json:"message,omitempty"`
	}
	// Set the content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Parse the JSON request
	var req WalletBalanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Open the SQLite database
	db, err := sql.Open("sqlite3", "nodeList1.db")
	if err != nil {
		http.Error(w, "Failed to connect to database", http.StatusInternalServerError)
		log.Println("Database connection error:", err)
		return
	}
	defer db.Close()

	// Query the wallet balance
	var balance float64
	err = db.QueryRow("SELECT balance FROM wallet_balances WHERE wallet = ?", req.Wallet).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			// Wallet not found
			response := WalletBalanceResponse{
				Status:  "error",
				Message: "Wallet not found",
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		// Internal error
		http.Error(w, "Failed to query database", http.StatusInternalServerError)
		log.Println("Query error:", err)
		return
	}

	// Wallet found, send the balance
	response := WalletBalanceResponse{
		Status:  "success",
		Balance: balance,
	}
	json.NewEncoder(w).Encode(response)
}

// Work In Progress
func hashDatabaseHandler(w http.ResponseWriter, r *http.Request) {
	dbFilename := filepath.Join(".", "nodeList1.db")
	hash := cryptoUtils.ComputeDatabaseHash(dbFilename)
	response := map[string]string{"hash": hash}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func insertNewWallet(w http.ResponseWriter, r *http.Request) {

	databaseFile := "nodeList1.db"
	cryptoUtils.NewWallet("Gasp, can it be!", databaseFile)
}

func shuffleDatabase(w http.ResponseWriter, r *http.Request) {
	shuffled, _ := cryptoUtils.ShuffleRows("nodeList1.db", "computers", 1)

	fmt.Println(len(shuffled))

}
