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

	http.HandleFunc("/keyGenUi", keyGenUi)
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

func VerifySignatureOLD(w http.ResponseWriter, r *http.Request) {

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
	publicKey := req.Transaction.From

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

	result, err := cryptoUtils.VerifySignature(publicKey, reorderedCleanedJson, signature)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println(result)
	}

	w.Write([]byte(publicKey))
	fmt.Println(reorderedCleanedJson)
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
	fmt.Println(cryptoUtils.VerifySignature(signature, publicKey, message))
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
	// bemFile := "public_key.pem"
	// binaryData, _ := cryptoUtils.ImportPEMFile(bemFile)
	// binaryTobase58 := cryptoUtils.BinaryToBase58Check(binaryData)
	// binaryToBase64, _ := cryptoUtils.BinaryToBase64(binaryData)
	// Base58CheckToBinary, _ := cryptoUtils.Base58CheckToBinary(binaryTobase58)
	// binarytopem, _ := cryptoUtils.BinaryToPEM(Base58CheckToBinary, "public")
	// fmt.Println(binaryToBase64)
	cryptoUtils.KeyGen()

}

func keyGenUi(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "keyGenUi.html")

}
