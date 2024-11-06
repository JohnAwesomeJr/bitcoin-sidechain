package cryptoUtils

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/asn1"
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"sort"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"

	"fmt"
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

func ComputeDatabaseHash(dbFilename string) string {
	// Open the SQLite database
	db, err := sql.Open("sqlite", dbFilename)
	if err != nil {
		return fmt.Sprintf("failed to open database: %v", err)
	}
	defer db.Close()

	// Query all table names
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		return fmt.Sprintf("failed to query table names: %v", err)
	}
	defer rows.Close()

	// Buffer to hold CSV data
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)

	// Iterate through tables and write their data to buffer
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return fmt.Sprintf("failed to scan table name: %v", err)
		}

		// Query table rows
		tableRows, err := db.Query(fmt.Sprintf("SELECT * FROM %s", table))
		if err != nil {
			return fmt.Sprintf("failed to query rows from table %s: %v", table, err)
		}
		defer tableRows.Close()

		// Get column names and write as header
		columns, err := tableRows.Columns()
		if err != nil {
			return fmt.Sprintf("failed to get columns from table %s: %v", table, err)
		}
		writer.Write(columns)

		// Write rows to buffer
		for tableRows.Next() {
			columnPointers := make([]interface{}, len(columns))
			for i := range columnPointers {
				columnPointers[i] = new(interface{}) // Allocate new interface{}
			}

			if err := tableRows.Scan(columnPointers...); err != nil {
				return fmt.Sprintf("failed to scan row from table %s: %v", table, err)
			}

			rowData := make([]string, len(columns))
			for i, col := range columnPointers {
				if col != nil {
					rowData[i] = fmt.Sprintf("%v", *(col.(*interface{}))) // Dereference the interface
				}
			}
			writer.Write(rowData)
		}
	}

	writer.Flush()

	// Compute SHA-256 hash of the buffer content
	hash := sha256.Sum256(buffer.Bytes())
	return hex.EncodeToString(hash[:])
}

// NewWallet function takes a wallet address, checks if it exists, and creates it with a balance of 0 if it doesn't.
func NewWallet(walletAddress string) error {
	// Connect to the SQLite database (assuming database file is wallet.db)
	db, err := sql.Open("sqlite3", "nodeList1.db")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Check if wallet already exists
	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM wallet_balances WHERE wallet = ?)", walletAddress).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check wallet existence: %w", err)
	}

	// If wallet already exists, do nothing
	if exists {
		fmt.Println("Wallet already exists.")
		return nil
	}

	// If wallet does not exist, create it with a balance of 0
	_, err = db.Exec("INSERT INTO wallet_balances (wallet, balance) VALUES (?, 0)", walletAddress)
	if err != nil {
		return fmt.Errorf("failed to create new wallet: %w", err)
	}

	fmt.Println("New wallet created with balance 0.")
	return nil
}
