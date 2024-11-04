package cryptoUtils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
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
func KeyGen() {
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

func VerifySignature(signatureBase64 string, publicKeyPEM string, message string) (string, error) {
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
