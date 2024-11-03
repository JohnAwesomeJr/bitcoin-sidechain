package main

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
)

func main() {

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/hello", helloHandler)
	http.HandleFunc("/hashData", hashDatabase)
	http.HandleFunc("/transaction", transaction)
	http.HandleFunc("/keygen", keyGenApi)
	http.HandleFunc("/keysign", keySign)

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

func keyGenApi(w http.ResponseWriter, r *http.Request) {
	keyGen()
}
func formatPEMPublicKey(key string) string {
	// Define the PEM header and footer for a public key
	header := "-----BEGIN PUBLIC KEY-----\n"
	footer := "-----END PUBLIC KEY-----"

	// Split the key into lines of 64 characters
	formattedKey := ""
	for i := 0; i < len(key); i += 64 {
		end := i + 64
		if end > len(key) {
			end = len(key)
		}
		formattedKey += key[i:end] + "\n"
	}

	// Concatenate the header, formatted key, and footer
	return header + formattedKey + footer
}

// verifySignature verifies the signature against the public key and the message.
func verifySignature(signatureBase64 string, publicKeyPEM string, message string) (string, error) {
	// Decode the base64 encoded signature
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return "not valid", fmt.Errorf("failed to decode signature: %w", err)
	}

	// Decode the PEM formatted public key
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil || block.Type != "PUBLIC KEY" {
		return "not valid", errors.New("failed to parse PEM block containing public key")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "not valid", fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return "not valid", errors.New("not an RSA public key")
	}

	// Verify the signature
	if err := rsa.VerifyPKCS1v15(rsaPubKey, crypto.SHA256, hashMessage(message), signature); err != nil {
		return "not valid", nil
	}

	return "verified", nil
}

// hashMessage hashes the message using SHA256.
func hashMessage(message string) []byte {
	h := sha256.New()
	h.Write([]byte(message))
	return h.Sum(nil)
}
func removeWhitespace(input string) string {
	// Replace line breaks with empty string and trim spaces
	return strings.ReplaceAll(strings.TrimSpace(input), "\n", "")
}
func reorderJSON(jsonStr string) string {
	// Unmarshal the input JSON string into a map
	var data map[string]interface{}

	// Create a sorted slice of keys
	keys := make([]string, 0, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Create a new map to hold the ordered JSON
	orderedData := make(map[string]interface{})
	for _, key := range keys {
		orderedData[key] = data[key]
	}

	// Marshal the ordered map back into a JSON string
	reorderedJSON, err := json.Marshal(orderedData)
	if err != nil {
		return "", fmt.Errorf("error marshalling JSON: %w", err)
	}

	return string(reorderedJSON), nil
}

func keySign(w http.ResponseWriter, r *http.Request) {

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
	publicKey := formatPEMPublicKey(from)

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
	cleanedJson := removeWhitespace(string(jsonData))
	reorderedCleanedJson := reorderJSON(cleanedJson)

	// Print the JSON string
	fmt.Println(string(jsonData))

	result, err := verifySignature(signature, publicKey, reorderedCleanedJson)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println(result)
	}

	w.Write([]byte(publicKey))

}
