package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"log"
)

func sign(data string, privateKey []byte) (string, error) {
	// Convert the private key to PEM format
	// Parse the private key
	block, _ := pem.Decode(privateKey)
	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block containing the private key")
	}

	privateKeyParsed, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %v", err)
	}

	// Hash the UUID using SHA-256
	hash := sha256.New()
	hash.Write([]byte(data))
	hashed := hash.Sum(nil)

	// Sign the hashed UUID using the private key and SHA-256
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKeyParsed, crypto.SHA256, hashed)
	if err != nil {
		log.Fatalf("Failed to sign UUID: %v", err)
		return "", err
	}
	return base64.StdEncoding.EncodeToString(signature), nil
}
