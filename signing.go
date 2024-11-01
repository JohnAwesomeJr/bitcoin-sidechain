package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strings"
)

// Sign a message using the private key
func signMessage(message string) (string, error) {
	privKey, err := loadPrivateKey("private_key.pem")
	if err != nil {
		return "", err
	}

	hash := []byte(message) // For simplicity, using the message directly as the hash
	r, s, err := ecdsa.Sign(rand.Reader, privKey, hash)
	if err != nil {
		return "", fmt.Errorf("failed to sign message: %w", err)
	}

	signature := r.String() + "," + s.String()
	return signature, nil
}

// Verify a message using the public key and its signature
func verifyMessage(message, signature string) (bool, error) {
	pubKey, err := loadPublicKey("public_key.pem")
	if err != nil {
		return false, err
	}

	hash := []byte(message) // For simplicity, using the message directly as the hash
	parts := splitSignature(signature)
	if len(parts) != 2 {
		return false, errors.New("invalid signature format")
	}

	r := new(big.Int)
	s := new(big.Int)
	r.SetString(parts[0], 10)
	s.SetString(parts[1], 10)

	valid := ecdsa.Verify(pubKey, hash, r, s)
	if valid {
		fmt.Println("The signature is valid.")
	} else {
		fmt.Println("The signature is invalid.")
	}

	return valid, nil
}

// Load the private key from a PEM file
func loadPrivateKey(filename string) (*ecdsa.PrivateKey, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open private key file: %w", err)
	}
	defer file.Close()

	// Read the entire file
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	block, _ := pem.Decode(fileBytes)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	return x509.ParseECPrivateKey(block.Bytes)
}

// Load the public key from a PEM file
func loadPublicKey(filename string) (*ecdsa.PublicKey, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open public key file: %w", err)
	}
	defer file.Close()

	// Read the entire file
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	block, _ := pem.Decode(fileBytes)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	return publicKey.(*ecdsa.PublicKey), nil
}

// Split the signature string into its components
func splitSignature(signature string) []string {
	return strings.Split(signature, ",")
}
