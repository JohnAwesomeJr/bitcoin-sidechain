package cryptoUtils

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/asn1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"log"
	"math/big"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"

	"fmt"

	_ "github.com/mattn/go-sqlite3"

	_ "github.com/go-sql-driver/mysql" // This imports the MySQL driver
)

func KeyGenRSAOLDSTYLE() {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println("Error generating private key:", err)
		return
	}

	// Save the private key to a file
	privateFile, err := os.Create("private_key.pem")
	if err != nil {
		fmt.Println("Error creating private key file:", err)
		return
	}
	defer privateFile.Close()

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privatePem := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	if err := pem.Encode(privateFile, privatePem); err != nil {
		fmt.Println("Error encoding private key:", err)
		return
	}

	fmt.Println("Private key saved to private_key.pem")

	// Extract the public key
	publicKey := &privateKey.PublicKey

	// Save the public key to a file
	publicFile, err := os.Create("public_key.pem")
	if err != nil {
		fmt.Println("Error creating public key file:", err)
		return
	}
	defer publicFile.Close()

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		fmt.Println("Error marshaling public key:", err)
		return
	}

	publicPem := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	if err := pem.Encode(publicFile, publicPem); err != nil {
		fmt.Println("Error encoding public key:", err)
		return
	}

	fmt.Println("Public key saved to public_key.pem")
}
func KeyGenOldECStyle() {
	// Generate a secp256k1 key pair
	privateKey, err := btcec.NewPrivateKey()
	if err != nil {
		fmt.Println("Error generating private key:", err)
		return
	}

	// Save the private key to a file
	privateFile, err := os.Create("private_key.pem")
	if err != nil {
		fmt.Println("Error creating private key file:", err)
		return
	}
	defer privateFile.Close()

	privateKeyBytes := privateKey.Serialize()
	privatePem := &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	if err := pem.Encode(privateFile, privatePem); err != nil {
		fmt.Println("Error encoding private key:", err)
		return
	}

	fmt.Println("Private key saved to private_key.pem")

	// Save the public key to a file
	publicFile, err := os.Create("public_key.pem")
	if err != nil {
		fmt.Println("Error creating public key file:", err)
		return
	}
	defer publicFile.Close()

	publicKeyBytes := privateKey.PubKey().SerializeCompressed()
	publicPem := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	if err := pem.Encode(publicFile, publicPem); err != nil {
		fmt.Println("Error encoding public key:", err)
		return
	}

	fmt.Println("Public key saved to public_key.pem")
}

func KeyGen() {
	// Generate a new private key
	privateKey, err := btcec.NewPrivateKey()
	if err != nil {
		fmt.Println("Error generating private key:", err)
		return
	}

	// Get the public key from the private key
	publicKey := privateKey.PubKey()

	// Convert private key to bytes
	privateKeyBytes := privateKey.Serialize()

	// Convert public key to bytes (uncompressed format)
	pubKeyBytes := publicKey.SerializeUncompressed()

	// Print the keys in Base64 format
	fmt.Printf("Private Key: %s\n", base64.StdEncoding.EncodeToString(privateKeyBytes))
	fmt.Printf("Public Key: %s\n", base64.StdEncoding.EncodeToString(pubKeyBytes))
}

func FormatPEMPublicKey(key string) string {
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

func VerifySignatureOLD(signatureBase64 string, publicKeyPEM string, message string) (string, error) {
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
	if err := rsa.VerifyPKCS1v15(rsaPubKey, crypto.SHA256, HashMessage(message), signature); err != nil {
		return "not valid", nil
	}

	return "verified", nil
}

func VerifySignatureOLD2(publicKeyBase64 string, message string, signatureBase64 string) (string, error) {
	// Decode the public key from Base64
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return "Not Valid", err
	}

	// Parse the public key
	publicKey, err := x509.ParsePKIXPublicKey(publicKeyBytes)
	if err != nil {
		return "Not Valid", err
	}

	ecdsaPublicKey, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "Not Valid", errors.New("not an ECDSA public key")
	}

	// Decode the signature from Base64
	signatureBytes, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return "Not Valid", err
	}

	// Hash the message
	hash := sha256.Sum256([]byte(message))

	// Split the signature into r and s values
	half := len(signatureBytes) / 2
	r := big.NewInt(0).SetBytes(signatureBytes[:half])
	s := big.NewInt(0).SetBytes(signatureBytes[half:])

	// Verify the signature
	valid := ecdsa.Verify(ecdsaPublicKey, hash[:], r, s)
	if valid {
		return "Valid", nil
	}

	return "Not Valid", nil
}

// VerifySignature verifies if the given base64 encoded signature is valid for the message signed with the given public key.
func VerifySignature(signatureB64, publicKeyB64, jsonMessage string) (string, error) {
	type ECDSASignature struct {
		R *big.Int
		S *big.Int
	}
	// Decode the base64 encoded signature
	signatureBytes, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return "not valid", fmt.Errorf("failed to decode signature: %v", err)
	}

	// Decode the base64 encoded public key
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyB64)
	if err != nil {
		return "not valid", fmt.Errorf("failed to decode public key: %v", err)
	}

	// Parse the public key
	pubKey, err := btcec.ParsePubKey(publicKeyBytes)
	if err != nil {
		return "not valid", fmt.Errorf("failed to parse public key: %v", err)
	}

	// Clean the JSON message
	cleanedMessage, err := cleanJSON(jsonMessage)
	if err != nil {
		return "not valid", fmt.Errorf("failed to clean JSON message: %v", err)
	}

	// Hash the cleaned message
	hash := sha256.Sum256([]byte(cleanedMessage))

	// Parse the ECDSA signature (DER format)
	var sig ECDSASignature
	if _, err := asn1.Unmarshal(signatureBytes, &sig); err != nil {
		return "not valid", fmt.Errorf("failed to unmarshal DER signature: %v", err)
	}

	// Verify the signature
	valid := ecdsa.Verify(pubKey.ToECDSA(), hash[:], sig.R, sig.S)

	if valid {
		return "valid", nil
	}
	return "not valid", errors.New("signature verification failed")
}

// cleanJSON cleans the JSON string by sorting the keys and removing whitespace and line breaks.
func cleanJSON(jsonString string) (string, error) {
	var jsonData map[string]interface{}

	// Unmarshal the JSON string into a map
	err := json.Unmarshal([]byte(jsonString), &jsonData)
	if err != nil {
		return "", err
	}

	// Create a sorted list of keys
	var keys []string
	for key := range jsonData {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Build a cleaned JSON string
	var cleanedBuilder strings.Builder
	cleanedBuilder.WriteString("{")
	for i, key := range keys {
		if i > 0 {
			cleanedBuilder.WriteString(",")
		}
		value := jsonData[key]
		valueBytes, _ := json.Marshal(value)
		cleanedBuilder.WriteString("\"" + key + "\":" + string(valueBytes))
	}
	cleanedBuilder.WriteString("}")

	return cleanedBuilder.String(), nil
}

func HashMessage(message string) []byte {
	h := sha256.New()
	h.Write([]byte(message))
	return h.Sum(nil)
}

func RemoveWhitespace(input string) string {
	// Replace line breaks with empty string and trim spaces
	return strings.ReplaceAll(strings.TrimSpace(input), "\n", "")
}

func ReorderJSON(jsonStr string) (string, error) {
	// Unmarshal the input JSON string into a map
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", fmt.Errorf("error unmarshalling JSON: %w", err)
	}
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

func ComputeDatabaseHash() string {
	// MySQL connection string (DSN format)
	dsn := "node:test@tcp(node-1-database:3306)/node" // Modify with your actual MySQL connection string

	// Open the MySQL database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Query sorted data from the "nodes" table
	rows, err := db.Query(`
		SELECT sort_order, computer_id, ip_address, node_group 
		FROM nodes 
		ORDER BY sort_order, computer_id, ip_address, node_group
	`)
	if err != nil {
		log.Fatalf("Failed to query rows: %v", err)
	}
	defer rows.Close()

	// Collect rows as strings
	var rowsData []string
	for rows.Next() {
		var sortOrder int
		var computerID, ipAddress string
		var nodeGroup int
		if err := rows.Scan(&sortOrder, &computerID, &ipAddress, &nodeGroup); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}

		// Format each row consistently
		rowsData = append(rowsData, fmt.Sprintf("%d|%s|%s|%d", sortOrder, computerID, ipAddress, nodeGroup))
	}

	// Join rows into a single string
	normalizedData := strings.Join(rowsData, "\n")

	// Compute SHA-256 hash of the normalized data
	hash := sha256.Sum256([]byte(normalizedData))
	return hex.EncodeToString(hash[:])
}

func NewWallet(walletAddress string, dbName string) (bool, error) {
	// Create the MySQL connection string (Data Source Name)
	dsn := "node:test@tcp(node-1-database:3306)/node" // Modify this as per your setup

	// Open a connection to the MySQL database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return false, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Check if wallet already exists
	var exists int
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM wallet_balances WHERE wallet = ?)", walletAddress).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check wallet existence: %w", err)
	}

	// If wallet already exists, do nothing
	if exists == 1 {
		fmt.Println("Wallet already exists.")
		return false, fmt.Errorf("wallet already exists")
	}

	// If wallet does not exist, create it with a balance of 0
	_, err = db.Exec("INSERT INTO wallet_balances (wallet, balance) VALUES (?, 0)", walletAddress)
	if err != nil {
		return false, fmt.Errorf("failed to create new wallet: %w", err)
	}

	fmt.Println("New wallet created with balance 0.")

	return true, nil
}

// MoveSats moves an amount from one wallet to another, checking for sufficient balance.
func MoveSats(fromAddress string, toAddress string, amount string, database string) error {
	// MySQL connection string
	dsn := "node:node@tcp(node-1-database:3306)/" + database // Modify this based on your MySQL setup

	// Open the MySQL database connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	// Convert amount to integer
	amountInt, err := strconv.Atoi(amount)
	if err != nil {
		return fmt.Errorf("invalid amount: %v", err)
	}

	// Begin a transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	// Rollback if something goes wrong
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Check balance of fromAddress
	var fromBalance int
	err = tx.QueryRow("SELECT balance FROM wallet_balances WHERE wallet = ?", fromAddress).Scan(&fromBalance)
	if err != nil {
		return fmt.Errorf("failed to retrieve balance for fromAddress: %v", err)
	}

	// Ensure fromAddress has sufficient funds
	if fromBalance < amountInt {
		return fmt.Errorf("insufficient funds in wallet %s", fromAddress)
	}

	// Subtract amount from fromAddress
	_, err = tx.Exec("UPDATE wallet_balances SET balance = balance - ? WHERE wallet = ?", amountInt, fromAddress)
	if err != nil {
		return fmt.Errorf("failed to deduct amount from fromAddress: %v", err)
	}

	// Add amount to toAddress
	_, err = tx.Exec("UPDATE wallet_balances SET balance = balance + ? WHERE wallet = ?", amountInt, toAddress)
	if err != nil {
		return fmt.Errorf("failed to add amount to toAddress: %v", err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// CheckAndAddNonce opens an SQLite database, checks if a nonce exists in the nonce table,
// and adds it if it does not exist, then returns false.
func CheckNonce(dbFileName string, nonce string) (bool, error) {
	// Create the MySQL connection string (Data Source Name)
	// You should have your DB user, password, host, and database name set as per your configuration
	dsn := "node:test@tcp(node-1-database:3306)/node" // Modify this as per your setup
	// Open a connection to the MySQL database
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return false, fmt.Errorf("could not open database: %w", err)
	}
	defer db.Close()

	// Prepare the query to check if the nonce exists
	query := "SELECT EXISTS(SELECT 1 FROM nonce WHERE nonce = ?)"

	// Create a variable to hold the result of the query
	var exists int

	// Execute the query
	err = db.QueryRow(query, nonce).Scan(&exists)
	if err != nil {
		log.Println("Error executing query:", err)
		return false, fmt.Errorf("query execution failed: %w", err)
	}

	// If the nonce does not exist, insert it into the database
	if exists == 0 {
		insertQuery := "INSERT INTO nonce (nonce) VALUES (?)"
		_, err = db.Exec(insertQuery, nonce)
		if err != nil {
			log.Println("Error inserting nonce:", err)
			return false, fmt.Errorf("failed to insert nonce: %w", err)
		}
		// Return false as per the requirement after adding the nonce
		return false, nil
	}

	// Return true if the nonce already existed
	return true, nil
}
