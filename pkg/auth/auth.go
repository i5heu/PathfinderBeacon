package auth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
)

type Key struct {
	PrivateKey *rsa.PrivateKey
}

func GenerateKey() (*Key, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("Failed to generate private key: %v", err)
	}
	return &Key{PrivateKey: privateKey}, nil
}

func FromPem(pemBytes []byte) (*Key, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("Failed to decode pem block")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse private key: %v", err)
	}

	return &Key{PrivateKey: privateKey}, nil
}

func (a *Key) PrivateKeyToPem() string {
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(a.PrivateKey),
	})

	return string(privateKeyPEM)
}

func (a *Key) PublicKeyToPem() []byte {
	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&a.PrivateKey.PublicKey),
	})

	return publicKeyPEM
}

func (a *Key) PublicKeyToPemBase64() string {
	return base64.StdEncoding.EncodeToString(a.PublicKeyToPem())
}

func (a *Key) GetRoomName() string {
	// get public key and hash it with sha224
	publicKey := a.PublicKeyToPem()
	hash := sha256.Sum224(publicKey)
	return hex.EncodeToString(hash[:])
}

func (a *Key) GetRoomSignature() ([]byte, error) {
	roomName := a.GetRoomName()
	roomHash := sha512.Sum512([]byte(roomName))

	signature, err := rsa.SignPKCS1v15(rand.Reader, a.PrivateKey, crypto.SHA512, roomHash[:])
	if err != nil {
		return nil, fmt.Errorf("Failed to sign room name: %v", err)
	}
	return signature, nil
}

func VerifyRoomSignature(roomName string, signatureBase64 string, publicKey string) (bool, error) {
	roomHash := sha512.Sum512([]byte(roomName))

	// Decode the base64 signature
	signature, err := base64.StdEncoding.DecodeString(signatureBase64)
	if err != nil {
		return false, fmt.Errorf("Failed to decode signature: %v", err)
	}

	// Decode the public key
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return false, fmt.Errorf("Failed to decode public key: %v", err)
	}

	// Parse the public key
	block, _ := pem.Decode(publicKeyBytes)
	if block == nil {
		return false, fmt.Errorf("Failed to decode pem block")
	}

	publicKeyParsed, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return false, fmt.Errorf("Failed to parse public key: %v", err)
	}

	// Verify the signature using the public key
	err = rsa.VerifyPKCS1v15(publicKeyParsed, crypto.SHA512, roomHash[:], signature)
	if err != nil {
		return false, fmt.Errorf("Signature verification failed: %v", err)
	}
	return true, nil
}
