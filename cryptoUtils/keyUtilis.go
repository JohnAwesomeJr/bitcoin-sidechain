package cryptoUtils

import (
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
func BinaryToPEM(data []byte) (string, error) {
	block := &pem.Block{
		Type:  "PUBLIC KEY",
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
