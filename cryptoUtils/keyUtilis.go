package cryptoUtils

import (
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io/ioutil"

	"github.com/btcsuite/btcutil/base58"
)

// ImportPEMFile imports a PEM file and returns the raw binary data.
func ImportPEMFile(filename string) ([]byte, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	return block.Bytes, nil
}

// BinaryToPEM converts binary data to PEM format.
// If private = "private"
// If public = "public"
func BinaryToPEM(data []byte, keyType string) (string, error) {
	var blockType string
	if keyType == "private" {
		blockType = "EC PRIVATE KEY"
	} else if keyType == "public" {
		blockType = "PUBLIC KEY"
	} else {
		return "", fmt.Errorf("invalid key type: %s", keyType)
	}

	block := &pem.Block{
		Type:  blockType,
		Bytes: data,
	}
	pemData := pem.EncodeToMemory(block)
	return string(pemData), nil
}

// SavePEMFile saves the PEM data to a file.
func SavePEMFile(filename, pemData string) error {
	return ioutil.WriteFile(filename, []byte(pemData), 0644)
}

// BinaryToBase58Check encodes the binary data to Base58Check format.
func BinaryToBase58Check(data []byte) string {
	return base58.Encode(data)
}

// Base58CheckToBinary decodes Base58Check back to binary format.
func Base58CheckToBinary(encoded string) ([]byte, error) {
	return base58.Decode(encoded), nil
}

// BinaryToBase64 converts binary data to Base64 format.
func BinaryToBase64(data []byte) (string, error) {
	if data == nil {
		return "", fmt.Errorf("input data cannot be nil")
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	return encoded, nil
}
