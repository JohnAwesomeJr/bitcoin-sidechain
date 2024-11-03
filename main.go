package main

import (
	"bitcoin-sidechain/cryptoUtils"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
)

func main() {

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/hashData", hashDatabase)
	http.HandleFunc("/transaction", transaction)
	http.HandleFunc("/keygen", KeyGenApi)
	http.HandleFunc("/verifysignature", VerifySignature)
	http.HandleFunc("/bem", bemTest)

	if err := http.ListenAndServe(":80", nil); err != nil {
		panic(err)
	}
}

func KeyGenApi(w http.ResponseWriter, r *http.Request) {
	cryptoUtils.KeyGen()
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
		"cryptoUtils":                           "false",
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

func VerifySignature(w http.ResponseWriter, r *http.Request) {

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

	// Parse the JSON body
	var req KeySignRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Assigning variables from the request
	signature := req.Signature
	from := req.Transaction.From
	to := req.Transaction.To
	amount := req.Transaction.Amount
	nonce := req.Transaction.Nonce
	publicKey := cryptoUtils.FormatPEMPublicKey(from)

	type justTransactionData struct {
		From   string `json:"from"`
		To     string `json:"to"`
		Amount string `json:"amount"`
		Nonce  string `json:"nonce"`
	}
	transactionToBeVerified := justTransactionData{
		From:   from,
		To:     to,
		Amount: amount,
		Nonce:  nonce,
	}
	jsonData, err := json.MarshalIndent(transactionToBeVerified, "", "  ")
	if err != nil {
		fmt.Println("Error marshaling to JSON:", err)
		return
	}
	cleanedJson := cryptoUtils.RemoveWhitespace(string(jsonData))
	reorderedCleanedJson, err := cryptoUtils.ReorderJSON(cleanedJson)
	if err != nil {
		fmt.Println("There was an error")
	}

	// Print the JSON string
	// fmt.Println(string(jsonData))

	result, err := cryptoUtils.VerifySignature(signature, publicKey, reorderedCleanedJson)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println(result)
	}

	w.Write([]byte(publicKey))
}

func transaction(w http.ResponseWriter, r *http.Request) {
	type TransactionRequest struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	// Ensure the request content type is JSON
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	// Decode the JSON body into the struct
	var req TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Prepare and send a JSON response
	response := map[string]string{"from": req.From, "to": req.To}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func bemTest(w http.ResponseWriter, r *http.Request) {
	bemFile := "public_key.pem"
	binaryData, _ := cryptoUtils.ImportPEMFile(bemFile)
	binaryTobase58 := cryptoUtils.BinaryToBase58Check(binaryData)
	Base58CheckToBinary, _ := cryptoUtils.Base58CheckToBinary(binaryTobase58)
	binarytopem, _ := cryptoUtils.BinaryToPEM(Base58CheckToBinary)
	fmt.Println(binarytopem)
}
