package tests

import (
	"simple_blockchain/pkg/CryptoGraphy"
	"simple_blockchain/pkg/blockchain"
	"testing"
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
