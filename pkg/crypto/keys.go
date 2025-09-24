package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// KeyPair represents an Ed25519 key pair
type KeyPair struct {
	PrivateKey ed25519.PrivateKey
	PublicKey  ed25519.PublicKey
	Address    string
}

func GenerateKeyPair() (*KeyPair, error) {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Derive address from public key (similar to Bitcoin's approach)
	hash := sha256.Sum256(pubKey)
	address := hex.EncodeToString(hash[:20])

	return &KeyPair{
		PrivateKey: privKey,
		PublicKey:  pubKey,
		Address:    address,
	}, nil
}

func LoadKeyPairFromHex(privateKeyHex, publicKeyHex string) (*KeyPair, error) {
	privBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key hex: %w", err)
	}

	pubBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid public key hex: %w", err)
	}

	privKey := ed25519.PrivateKey(privBytes)
	pubKey := ed25519.PublicKey(pubBytes)

	hash := sha256.Sum256(pubKey)
	address := hex.EncodeToString(hash[:20])

	return &KeyPair{
		PrivateKey: privKey,
		PublicKey:  pubKey,
		Address:    address,
	}, nil
}

func (kp *KeyPair) Sign(message []byte) []byte {
	return ed25519.Sign(kp.PrivateKey, message)
}

func (kp *KeyPair) Verify(message []byte, signature []byte) bool {
	return ed25519.Verify(kp.PublicKey, message, signature)
}

func (kp *KeyPair) GetPrivateKeyHex() string {
	return hex.EncodeToString(kp.PrivateKey)
}

func (kp *KeyPair) GetPublicKeyHex() string {
	return hex.EncodeToString(kp.PublicKey)
}

func DeriveAddressFromPublicKey(pubKeyHex string) (string, error) {
	pubKey, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid public key hex: %w", err)
	}

	hash := sha256.Sum256(pubKey)
	return hex.EncodeToString(hash[:20]), nil
}

func VerifySignature(pubKeyHex string, message []byte, signature []byte) bool {
	pubKey, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return false
	}

	return ed25519.Verify(ed25519.PublicKey(pubKey), message, signature)
}
