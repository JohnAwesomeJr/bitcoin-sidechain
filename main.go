package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {
	// Define command-line flags
	generateKeysFlag := flag.Bool("generate-keys", false, "Generate a new public/private key pair")
	signMessageFlag := flag.String("sign", "", "Sign a message")
	verifyMessageFlag := flag.String("verify", "", "Verify a message with its signature")
	signatureFlag := flag.String("signature", "", "Signature to verify")

	// Parse the flags
	flag.Parse()

	if *generateKeysFlag {
		err := generateKeys()
		if err != nil {
			log.Fatalf("Error generating keys: %v", err)
		}
		fmt.Println("Keys generated successfully.")
	}

	if *signMessageFlag != "" {
		signature, err := signMessage(*signMessageFlag)
		if err != nil {
			log.Fatalf("Error signing message: %v", err)
		}
		fmt.Printf("Signature: %s\n", signature)
	}

	if *verifyMessageFlag != "" && *signatureFlag != "" {
		isValid, err := verifyMessage(*verifyMessageFlag, *signatureFlag)
		if err != nil {
			log.Fatalf("Error verifying message: %v", err)
		}
		if isValid {
			fmt.Println("Signature is valid.")
		} else {
			fmt.Println("Signature is NOT valid.")
		}
	}
}
