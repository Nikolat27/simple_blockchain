package tests

import (
	"testing"

	"github.com/Nikolat27/simple_blockchain/pkg/CryptoGraphy"
	"github.com/Nikolat27/simple_blockchain/pkg/blockchain"
)

func TestTx_Hash(t *testing.T) {
	tx := &blockchain.Transaction{
		From:      "Alice",
		To:        "Bob",
		Amount:    100,
		Timestamp: 1234567890,
		Status:    "pending",
	}

	hash := tx.Hash()
	if len(hash) == 0 {
		t.Error("Hash should not be empty")
	}
}

func TestTx_SignAndVerify(t *testing.T) {
	keyPair, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}
	tx := &blockchain.Transaction{
		From:      "Alice",
		To:        "Bob",
		Amount:    50,
		Timestamp: 1234567890,
		Status:    "pending",
	}

	if err = tx.Sign(keyPair); err != nil {
		t.Fatalf("Sign failed: %v", err)
	}

	if !tx.Verify() {
		t.Error("Signature verification failed")
	}
}

func TestTx_Verify_InvalidSignature(t *testing.T) {
	tx := &blockchain.Transaction{
		From:      "Alice",
		To:        "Bob",
		Amount:    50,
		Timestamp: 1234567890,
		Status:    "pending",
		PublicKey: "invalid",
		Signature: []byte("bad"),
	}

	if tx.Verify() {
		t.Error("Verify should fail for invalid signature")
	}
}

func TestCreateCoinbaseTx(t *testing.T) {
	miner := "Miner1"
	reward := uint64(100)

	tx := blockchain.CreateCoinbaseTx(miner, reward)
	if !tx.IsCoinbase {
		t.Error("Coinbase transaction should have IsCoinbase=true")
	}

	if tx.Status != "confirmed" {
		t.Error("Coinbase transaction should have Status='confirmed'")
	}

	if tx.To != miner || tx.Amount != reward {
		t.Error("Coinbase transaction fields not set correctly")
	}
}

func TestTxSize(t *testing.T) {
	tx := &blockchain.Transaction{
		From:      "Alice",
		To:        "Bob",
		Amount:    50,
		Timestamp: 1234567890,
		Status:    "pending",
		PublicKey: "invalid",
		Signature: []byte("bad"),
	}

	if size := tx.Size(); size != 70 {
		t.Errorf("the size function is not working properly, got %d, want 70", size)
	}
}

func TestTx_HashConsistency(t *testing.T) {
	tx := &blockchain.Transaction{
		From:      "Alice",
		To:        "Bob",
		Amount:    100,
		Timestamp: 1234567890,
		Status:    "pending",
	}

	// Hash twice and verify consistency
	hash1 := tx.Hash()
	hash2 := tx.Hash()

	if string(hash1) != string(hash2) {
		t.Error("Transaction hash should be consistent")
	}
}

func TestTx_HashExcludesSignature(t *testing.T) {
	keyPair, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	tx := &blockchain.Transaction{
		From:      "Alice",
		To:        "Bob",
		Amount:    100,
		Timestamp: 1234567890,
		Status:    "pending",
		PublicKey: keyPair.GetPublicKeyHex(), // Set public key before hashing
	}

	// Hash before signing
	hashBefore := tx.Hash()

	// Sign transaction (this sets signature but public key is already set)
	err = tx.Sign(keyPair)
	if err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}

	// Hash after signing (should be the same since signature is excluded from hash)
	hashAfter := tx.Hash()

	if string(hashBefore) != string(hashAfter) {
		t.Error("Transaction hash should not change after signing (signature is excluded from hash)")
	}
}

func TestTx_SignWithHexKeys(t *testing.T) {
	keyPair, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	privateKeyHex := keyPair.GetPrivateKeyHex()
	publicKeyHex := keyPair.GetPublicKeyHex()

	tx := &blockchain.Transaction{
		From:      "Alice",
		To:        "Bob",
		Amount:    100,
		Timestamp: 1234567890,
		Status:    "pending",
	}

	err = tx.SignWithHexKeys(privateKeyHex, publicKeyHex)
	if err != nil {
		t.Fatalf("Failed to sign with hex keys: %v", err)
	}

	if tx.Signature == nil {
		t.Error("Signature should not be nil after signing")
	}

	if tx.PublicKey == "" {
		t.Error("PublicKey should not be empty after signing")
	}

	// Verify the signature
	if !tx.Verify() {
		t.Error("Signature verification should pass")
	}
}

func TestTx_SignWithHexKeys_InvalidKeys(t *testing.T) {
	tx := &blockchain.Transaction{
		From:      "Alice",
		To:        "Bob",
		Amount:    100,
		Timestamp: 1234567890,
		Status:    "pending",
	}

	err := tx.SignWithHexKeys("invalid_private_key", "invalid_public_key")
	if err == nil {
		t.Error("Should return error for invalid hex keys")
	}
}

func TestTx_VerifyWithoutSignature(t *testing.T) {
	tx := &blockchain.Transaction{
		From:      "Alice",
		To:        "Bob",
		Amount:    100,
		Timestamp: 1234567890,
		Status:    "pending",
	}

	// Verify without signature should fail
	if tx.Verify() {
		t.Error("Verification should fail without signature")
	}
}

func TestTx_VerifyWithoutPublicKey(t *testing.T) {
	tx := &blockchain.Transaction{
		From:      "Alice",
		To:        "Bob",
		Amount:    100,
		Timestamp: 1234567890,
		Status:    "pending",
		Signature: []byte("some_signature"),
	}

	// Verify without public key should fail
	if tx.Verify() {
		t.Error("Verification should fail without public key")
	}
}

func TestTx_HashEncodeToString(t *testing.T) {
	tx := &blockchain.Transaction{
		From:      "Alice",
		To:        "Bob",
		Amount:    100,
		Timestamp: 1234567890,
		Status:    "pending",
	}

	hash := tx.Hash()
	hashStr := hash.EncodeToString()

	if hashStr == "" {
		t.Error("Encoded hash string should not be empty")
	}

	// Should be hex encoded (64 characters for 32 bytes)
	if len(hashStr) != 64 {
		t.Errorf("Expected hash string length 64, got %d", len(hashStr))
	}
}

func TestCreateCoinbaseTx_Properties(t *testing.T) {
	miner := "MinerAddress"
	reward := uint64(10000)

	tx := blockchain.CreateCoinbaseTx(miner, reward)

	// Verify all properties
	if tx.From != "" {
		t.Error("Coinbase transaction should have empty From field")
	}

	if tx.To != miner {
		t.Errorf("Expected To %s, got %s", miner, tx.To)
	}

	if tx.Amount != reward {
		t.Errorf("Expected Amount %d, got %d", reward, tx.Amount)
	}

	if tx.Fee != blockchain.CoinbaseTxFee {
		t.Errorf("Expected Fee %d, got %d", blockchain.CoinbaseTxFee, tx.Fee)
	}

	if tx.Status != "confirmed" {
		t.Errorf("Expected Status 'confirmed', got %s", tx.Status)
	}

	if !tx.IsCoinbase {
		t.Error("IsCoinbase should be true")
	}

	if tx.Timestamp == 0 {
		t.Error("Timestamp should be set")
	}
}

func TestNewTransaction(t *testing.T) {
	from := "Alice"
	to := "Bob"
	amount := uint64(100)
	timestamp := int64(1234567890)

	tx := blockchain.NewTransaction(from, to, amount, timestamp)

	if tx.From != from {
		t.Errorf("Expected From %s, got %s", from, tx.From)
	}

	if tx.To != to {
		t.Errorf("Expected To %s, got %s", to, tx.To)
	}

	if tx.Amount != amount {
		t.Errorf("Expected Amount %d, got %d", amount, tx.Amount)
	}

	if tx.Timestamp != timestamp {
		t.Errorf("Expected Timestamp %d, got %d", timestamp, tx.Timestamp)
	}

	if tx.Status != "pending" {
		t.Errorf("Expected Status 'pending', got %s", tx.Status)
	}

	if tx.IsCoinbase {
		t.Error("Regular transaction should not be coinbase")
	}
}

func TestTx_SizeWithDifferentFields(t *testing.T) {
	// Test with minimal fields
	tx1 := &blockchain.Transaction{
		From:   "",
		To:     "",
		Amount: 0,
	}
	size1 := tx1.Size()

	// Test with populated fields
	tx2 := &blockchain.Transaction{
		From:      "Alice",
		To:        "Bob",
		Amount:    100,
		Fee:       10,
		Timestamp: 1234567890,
		PublicKey: "publickey123",
		Signature: []byte("signature"),
		Status:    "confirmed",
	}
	size2 := tx2.Size()

	// Size should increase with more data
	if size2 <= size1 {
		t.Error("Transaction with more data should have larger size")
	}
}

func TestTx_SignSetsPublicKey(t *testing.T) {
	keyPair, err := CryptoGraphy.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	tx := &blockchain.Transaction{
		From:      "Alice",
		To:        "Bob",
		Amount:    100,
		Timestamp: 1234567890,
	}

	err = tx.Sign(keyPair)
	if err != nil {
		t.Fatalf("Failed to sign transaction: %v", err)
	}

	// Verify public key is set
	if tx.PublicKey == "" {
		t.Error("PublicKey should be set after signing")
	}

	// Verify public key matches the keypair
	expectedPublicKey := keyPair.GetPublicKeyHex()
	if tx.PublicKey != expectedPublicKey {
		t.Error("PublicKey should match the keypair's public key")
	}
}
