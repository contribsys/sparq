package util

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

// generate an RSA public and private keypair
func GenerateKeys() ([]byte, []byte) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(fmt.Errorf("GenerateKey %w", err))
	}

	err = privateKey.Validate()
	if err != nil {
		panic(fmt.Errorf("Validate %w", err))
	}

	privDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privBlock := pem.Block{Type: "RSA PRIVATE KEY", Headers: nil, Bytes: privDER}
	privatePem := pem.EncodeToMemory(&privBlock)

	pubDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		panic(fmt.Errorf("Marshal %w", err))
	}
	pubBlock := pem.Block{Type: "PUBLIC KEY", Headers: nil, Bytes: pubDER}
	publicPem := pem.EncodeToMemory(&pubBlock)

	return publicPem, privatePem
}

func parsePrivateKey(der []byte) (crypto.PrivateKey, error) {
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		switch key := key.(type) {
		case *rsa.PrivateKey:
			return key, nil
		default:
			return nil, fmt.Errorf("found unknown private key type in PKCS#8 wrapping")
		}
	}
	if key, err := x509.ParseECPrivateKey(der); err == nil {
		return key, nil
	}

	return nil, fmt.Errorf("failed to parse private key")
}

func parsePublicKey(der []byte) (crypto.PublicKey, error) {
	if key, err := x509.ParsePKCS1PublicKey(der); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKIXPublicKey(der); err == nil {
		switch key := key.(type) {
		case *rsa.PublicKey:
			return key, nil
		default:
			return nil, fmt.Errorf("found unknown public key type in PKIX wrapping")
		}
	}

	return nil, fmt.Errorf("failed to parse public key")
}

func DecodePrivateKey(k []byte) (crypto.PrivateKey, error) {
	block, _ := pem.Decode(k)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}

	return parsePrivateKey(block.Bytes)
}

func DecodePublicKey(k []byte) (crypto.PublicKey, error) {
	block, _ := pem.Decode(k)
	if block == nil || block.Type != "PUBLIC KEY" {
		if block != nil {
			return nil, fmt.Errorf("failed to decode PEM block containing public key. type: %v", block.Type)
		} else {
			return nil, fmt.Errorf("failed to decode PEM block containing public key.")
		}
	}

	return parsePublicKey(block.Bytes)
}
