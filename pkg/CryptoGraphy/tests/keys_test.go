package tests

import (
	"crypto/ed25519"
	"encoding/hex"
	"testing"

	"github.com/Nikolat27/simple_blockchain/pkg/CryptoGraphy"
)

// TestGenerateKeyPair tests the generation of a new key pair
func TestGenerateKeyPair(t *testing.T) {
	kp, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	if kp == nil {
		t.Fatal("KeyPair should not be nil")
	}

	// Verify private key length (Ed25519 private key is 64 bytes)
	if len(kp.PrivateKey) != ed25519.PrivateKeySize {
		t.Errorf("Expected private key size %d, got %d", ed25519.PrivateKeySize, len(kp.PrivateKey))
	}

	// Verify public key length (Ed25519 public key is 32 bytes)
	if len(kp.PublicKey) != ed25519.PublicKeySize {
		t.Errorf("Expected public key size %d, got %d", ed25519.PublicKeySize, len(kp.PublicKey))
	}

	// Verify address is generated (should be 40 hex characters for 20 bytes)
	if len(kp.Address) != 40 {
		t.Errorf("Expected address length 40, got %d", len(kp.Address))
	}

	// Verify address is valid hex
	_, err = hex.DecodeString(kp.Address)
	if err != nil {
		t.Errorf("Address should be valid hex: %v", err)
	}
}

// TestGenerateKeyPair_Uniqueness tests that generated key pairs are unique
func TestGenerateKeyPair_Uniqueness(t *testing.T) {
	kp1, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate first key pair: %v", err)
	}

	kp2, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate second key pair: %v", err)
	}

	// Verify keys are different
	if string(kp1.PrivateKey) == string(kp2.PrivateKey) {
		t.Error("Generated private keys should be unique")
	}

	if string(kp1.PublicKey) == string(kp2.PublicKey) {
		t.Error("Generated public keys should be unique")
	}

	if kp1.Address == kp2.Address {
		t.Error("Generated addresses should be unique")
	}
}

// TestKeyPair_Sign tests signing a message
func TestKeyPair_Sign(t *testing.T) {
	kp, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	message := []byte("Hello, Blockchain!")
	signature := kp.Sign(message)

	if signature == nil {
		t.Fatal("Signature should not be nil")
	}

	// Ed25519 signature is 64 bytes
	if len(signature) != ed25519.SignatureSize {
		t.Errorf("Expected signature size %d, got %d", ed25519.SignatureSize, len(signature))
	}
}

// TestKeyPair_SignAndVerify tests signing and verifying a message
func TestKeyPair_SignAndVerify(t *testing.T) {
	kp, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	message := []byte("Test message for signing")
	signature := kp.Sign(message)

	// Verify with correct key pair
	if !kp.Verify(message, signature) {
		t.Error("Signature verification should succeed with correct key pair")
	}
}

// TestKeyPair_Verify_WrongMessage tests verification with wrong message
func TestKeyPair_Verify_WrongMessage(t *testing.T) {
	kp, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	message := []byte("Original message")
	signature := kp.Sign(message)

	// Try to verify with different message
	wrongMessage := []byte("Different message")
	if kp.Verify(wrongMessage, signature) {
		t.Error("Signature verification should fail with wrong message")
	}
}

// TestKeyPair_Verify_WrongSignature tests verification with wrong signature
func TestKeyPair_Verify_WrongSignature(t *testing.T) {
	kp, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	message := []byte("Test message")
	wrongSignature := make([]byte, ed25519.SignatureSize)

	// Try to verify with wrong signature
	if kp.Verify(message, wrongSignature) {
		t.Error("Signature verification should fail with wrong signature")
	}
}

// TestKeyPair_Verify_WrongKeyPair tests verification with different key pair
func TestKeyPair_Verify_WrongKeyPair(t *testing.T) {
	kp1, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate first key pair: %v", err)
	}

	kp2, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate second key pair: %v", err)
	}

	message := []byte("Test message")
	signature := kp1.Sign(message)

	// Try to verify with different key pair
	if kp2.Verify(message, signature) {
		t.Error("Signature verification should fail with different key pair")
	}
}

// TestKeyPair_GetPrivateKeyHex tests getting private key as hex string
func TestKeyPair_GetPrivateKeyHex(t *testing.T) {
	kp, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	privKeyHex := kp.GetPrivateKeyHex()

	// Verify it's valid hex
	decoded, err := hex.DecodeString(privKeyHex)
	if err != nil {
		t.Errorf("Private key hex should be valid: %v", err)
	}

	// Verify length (64 bytes = 128 hex characters)
	if len(privKeyHex) != ed25519.PrivateKeySize*2 {
		t.Errorf("Expected private key hex length %d, got %d", ed25519.PrivateKeySize*2, len(privKeyHex))
	}

	// Verify it matches the original
	if string(decoded) != string(kp.PrivateKey) {
		t.Error("Decoded private key should match original")
	}
}

// TestKeyPair_GetPublicKeyHex tests getting public key as hex string
func TestKeyPair_GetPublicKeyHex(t *testing.T) {
	kp, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	pubKeyHex := kp.GetPublicKeyHex()

	// Verify it's valid hex
	decoded, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		t.Errorf("Public key hex should be valid: %v", err)
	}

	// Verify length (32 bytes = 64 hex characters)
	if len(pubKeyHex) != ed25519.PublicKeySize*2 {
		t.Errorf("Expected public key hex length %d, got %d", ed25519.PublicKeySize*2, len(pubKeyHex))
	}

	// Verify it matches the original
	if string(decoded) != string(kp.PublicKey) {
		t.Error("Decoded public key should match original")
	}
}

// TestLoadKeyPairFromHex tests loading a key pair from hex strings
func TestLoadKeyPairFromHex(t *testing.T) {
	// Generate original key pair
	originalKp, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Get hex strings
	privKeyHex := originalKp.GetPrivateKeyHex()
	pubKeyHex := originalKp.GetPublicKeyHex()

	// Load from hex
	loadedKp, err := CryptoGraphy.LoadKeyPairFromHex(privKeyHex, pubKeyHex)
	if err != nil {
		t.Fatalf("Failed to load key pair from hex: %v", err)
	}

	// Verify keys match
	if string(loadedKp.PrivateKey) != string(originalKp.PrivateKey) {
		t.Error("Loaded private key should match original")
	}

	if string(loadedKp.PublicKey) != string(originalKp.PublicKey) {
		t.Error("Loaded public key should match original")
	}

	if loadedKp.Address != originalKp.Address {
		t.Error("Loaded address should match original")
	}
}

// TestLoadKeyPairFromHex_InvalidPrivateKey tests loading with invalid private key
func TestLoadKeyPairFromHex_InvalidPrivateKey(t *testing.T) {
	kp, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	pubKeyHex := kp.GetPublicKeyHex()

	// Try to load with invalid private key hex
	_, err = CryptoGraphy.LoadKeyPairFromHex("invalid_hex", pubKeyHex)
	if err == nil {
		t.Error("Should return error for invalid private key hex")
	}
}

// TestLoadKeyPairFromHex_InvalidPublicKey tests loading with invalid public key
func TestLoadKeyPairFromHex_InvalidPublicKey(t *testing.T) {
	kp, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	privKeyHex := kp.GetPrivateKeyHex()

	// Try to load with invalid public key hex
	_, err = CryptoGraphy.LoadKeyPairFromHex(privKeyHex, "invalid_hex")
	if err == nil {
		t.Error("Should return error for invalid public key hex")
	}
}

// TestLoadKeyPairFromHex_SignAndVerify tests that loaded key pair can sign and verify
func TestLoadKeyPairFromHex_SignAndVerify(t *testing.T) {
	// Generate and save original key pair
	originalKp, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	privKeyHex := originalKp.GetPrivateKeyHex()
	pubKeyHex := originalKp.GetPublicKeyHex()

	// Load key pair
	loadedKp, err := CryptoGraphy.LoadKeyPairFromHex(privKeyHex, pubKeyHex)
	if err != nil {
		t.Fatalf("Failed to load key pair: %v", err)
	}

	// Sign with loaded key pair
	message := []byte("Test message")
	signature := loadedKp.Sign(message)

	// Verify with loaded key pair
	if !loadedKp.Verify(message, signature) {
		t.Error("Loaded key pair should be able to sign and verify")
	}

	// Verify with original key pair
	if !originalKp.Verify(message, signature) {
		t.Error("Original key pair should be able to verify signature from loaded key pair")
	}
}

// TestDeriveAddressFromPublicKey tests deriving address from public key
func TestDeriveAddressFromPublicKey(t *testing.T) {
	kp, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	pubKeyHex := kp.GetPublicKeyHex()
	address, err := CryptoGraphy.DeriveAddressFromPublicKey(pubKeyHex)
	if err != nil {
		t.Fatalf("Failed to derive address: %v", err)
	}

	// Verify address matches the key pair's address
	if address != kp.Address {
		t.Errorf("Derived address %s should match key pair address %s", address, kp.Address)
	}

	// Verify address length (40 hex characters for 20 bytes)
	if len(address) != 40 {
		t.Errorf("Expected address length 40, got %d", len(address))
	}
}

// TestDeriveAddressFromPublicKey_InvalidHex tests deriving address with invalid hex
func TestDeriveAddressFromPublicKey_InvalidHex(t *testing.T) {
	_, err := CryptoGraphy.DeriveAddressFromPublicKey("invalid_hex_string")
	if err == nil {
		t.Error("Should return error for invalid public key hex")
	}
}

// TestDeriveAddressFromPublicKey_Consistency tests address derivation consistency
func TestDeriveAddressFromPublicKey_Consistency(t *testing.T) {
	kp, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	pubKeyHex := kp.GetPublicKeyHex()

	// Derive address multiple times
	address1, err := CryptoGraphy.DeriveAddressFromPublicKey(pubKeyHex)
	if err != nil {
		t.Fatalf("Failed to derive address first time: %v", err)
	}

	address2, err := CryptoGraphy.DeriveAddressFromPublicKey(pubKeyHex)
	if err != nil {
		t.Fatalf("Failed to derive address second time: %v", err)
	}

	// Addresses should be identical
	if address1 != address2 {
		t.Error("Address derivation should be deterministic")
	}
}

// TestVerifySignature tests standalone signature verification
func TestVerifySignature(t *testing.T) {
	kp, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	message := []byte("Test message for verification")
	signature := kp.Sign(message)
	pubKeyHex := kp.GetPublicKeyHex()

	// Verify signature
	if !CryptoGraphy.VerifySignature(pubKeyHex, message, signature) {
		t.Error("Signature verification should succeed")
	}
}

// TestVerifySignature_WrongMessage tests verification with wrong message
func TestVerifySignature_WrongMessage(t *testing.T) {
	kp, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	message := []byte("Original message")
	signature := kp.Sign(message)
	pubKeyHex := kp.GetPublicKeyHex()

	// Try to verify with different message
	wrongMessage := []byte("Different message")
	if CryptoGraphy.VerifySignature(pubKeyHex, wrongMessage, signature) {
		t.Error("Signature verification should fail with wrong message")
	}
}

// TestVerifySignature_WrongPublicKey tests verification with wrong public key
func TestVerifySignature_WrongPublicKey(t *testing.T) {
	kp1, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate first key pair: %v", err)
	}

	kp2, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate second key pair: %v", err)
	}

	message := []byte("Test message")
	signature := kp1.Sign(message)
	wrongPubKeyHex := kp2.GetPublicKeyHex()

	// Try to verify with different public key
	if CryptoGraphy.VerifySignature(wrongPubKeyHex, message, signature) {
		t.Error("Signature verification should fail with wrong public key")
	}
}

// TestVerifySignature_InvalidPublicKeyHex tests verification with invalid hex
func TestVerifySignature_InvalidPublicKeyHex(t *testing.T) {
	kp, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	message := []byte("Test message")
	signature := kp.Sign(message)

	// Try to verify with invalid public key hex
	if CryptoGraphy.VerifySignature("invalid_hex", message, signature) {
		t.Error("Signature verification should fail with invalid public key hex")
	}
}

// TestVerifySignature_EmptySignature tests verification with empty signature
func TestVerifySignature_EmptySignature(t *testing.T) {
	kp, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	message := []byte("Test message")
	pubKeyHex := kp.GetPublicKeyHex()

	// Try to verify with empty signature
	if CryptoGraphy.VerifySignature(pubKeyHex, message, []byte{}) {
		t.Error("Signature verification should fail with empty signature")
	}
}

// TestSignature_Deterministic tests that signatures are deterministic
func TestSignature_Deterministic(t *testing.T) {
	kp, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	message := []byte("Test message")

	// Sign the same message twice
	signature1 := kp.Sign(message)
	signature2 := kp.Sign(message)

	// Signatures should be identical (Ed25519 is deterministic)
	if string(signature1) != string(signature2) {
		t.Error("Ed25519 signatures should be deterministic")
	}
}

// TestKeyPair_SignEmptyMessage tests signing an empty message
func TestKeyPair_SignEmptyMessage(t *testing.T) {
	kp, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	emptyMessage := []byte{}
	signature := kp.Sign(emptyMessage)

	// Should be able to sign empty message
	if signature == nil {
		t.Fatal("Should be able to sign empty message")
	}

	// Should be able to verify empty message
	if !kp.Verify(emptyMessage, signature) {
		t.Error("Should be able to verify signature of empty message")
	}
}

// TestKeyPair_SignLargeMessage tests signing a large message
func TestKeyPair_SignLargeMessage(t *testing.T) {
	kp, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Create a large message (1MB)
	largeMessage := make([]byte, 1024*1024)
	for i := range largeMessage {
		largeMessage[i] = byte(i % 256)
	}

	signature := kp.Sign(largeMessage)

	// Should be able to sign large message
	if signature == nil {
		t.Fatal("Should be able to sign large message")
	}

	// Should be able to verify large message
	if !kp.Verify(largeMessage, signature) {
		t.Error("Should be able to verify signature of large message")
	}
}

// TestAddress_Format tests that addresses are properly formatted
func TestAddress_Format(t *testing.T) {
	kp, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Address should be lowercase hex
	for _, char := range kp.Address {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			t.Errorf("Address should only contain lowercase hex characters, found: %c", char)
		}
	}
}

// TestConcurrentKeyGeneration tests concurrent key pair generation
func TestConcurrentKeyGeneration(t *testing.T) {
	const numGoroutines = 100
	done := make(chan *CryptoGraphy.KeyPair, numGoroutines)
	errors := make(chan error, numGoroutines)

	// Generate keys concurrently
	for i := 0; i < numGoroutines; i++ {
		go func() {
			kp, err := CryptoGraphy.GenerateKeyPair()
			if err != nil {
				errors <- err
				return
			}
			done <- kp
		}()
	}

	// Collect results
	keyPairs := make([]*CryptoGraphy.KeyPair, 0, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		select {
		case kp := <-done:
			keyPairs = append(keyPairs, kp)
		case err := <-errors:
			t.Fatalf("Error generating key pair: %v", err)
		}
	}

	// Verify all key pairs are unique
	addresses := make(map[string]bool)
	for _, kp := range keyPairs {
		if addresses[kp.Address] {
			t.Error("Found duplicate address in concurrent generation")
		}
		addresses[kp.Address] = true
	}

	if len(addresses) != numGoroutines {
		t.Errorf("Expected %d unique addresses, got %d", numGoroutines, len(addresses))
	}
}

// TestConcurrentSignAndVerify tests concurrent signing and verification
func TestConcurrentSignAndVerify(t *testing.T) {
	kp, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	const numGoroutines = 100
	done := make(chan bool, numGoroutines)

	// Sign and verify concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			message := []byte("Test message " + string(rune(id)))
			signature := kp.Sign(message)
			if !kp.Verify(message, signature) {
				done <- false
				return
			}
			done <- true
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		if !<-done {
			t.Error("Concurrent sign and verify failed")
		}
	}
}
