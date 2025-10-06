package test

import (
	"encoding/json"
	"testing"

	"github.com/Nikolat27/simple_blockchain/pkg/p2p/types"
)

// TestNewMessage tests creating a new message
func TestNewMessage(t *testing.T) {
	msgType := types.RequestHeadersMsg
	senderAddr := "127.0.0.1:8080"
	payload := types.Payload([]byte(`{"test":"data"}`))

	msg := types.NewMessage(msgType, senderAddr, payload)

	if msg == nil {
		t.Fatal("Message should not be nil")
	}

	if msg.Type != msgType {
		t.Errorf("Expected type %s, got %s", msgType, msg.Type)
	}

	if msg.SenderAddress != senderAddr {
		t.Errorf("Expected sender address %s, got %s", senderAddr, msg.SenderAddress)
	}

	if string(msg.Payload) != string(payload) {
		t.Error("Payload mismatch")
	}
}

// TestMessage_Marshal tests marshaling a message
func TestMessage_Marshal(t *testing.T) {
	msg := types.NewMessage(types.RequestHeadersMsg, "127.0.0.1:8080", types.Payload{})

	data := msg.Marshal()

	if len(data) == 0 {
		t.Error("Marshaled data should not be empty")
	}

	// Verify it's valid JSON
	var unmarshaled types.Message
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Errorf("Marshaled data should be valid JSON: %v", err)
	}

	if unmarshaled.Type != msg.Type {
		t.Error("Unmarshaled type doesn't match original")
	}

	if unmarshaled.SenderAddress != msg.SenderAddress {
		t.Error("Unmarshaled sender address doesn't match original")
	}
}

// TestMessage_MarshalUnmarshal tests round-trip marshaling
func TestMessage_MarshalUnmarshal(t *testing.T) {
	original := types.NewMessage(
		types.BlockBroadcastMsg,
		"localhost:9000",
		types.Payload([]byte(`{"block_id":123}`)),
	)

	// Marshal
	data := original.Marshal()

	// Unmarshal
	var decoded types.Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify
	if decoded.Type != original.Type {
		t.Errorf("Type mismatch: expected %s, got %s", original.Type, decoded.Type)
	}

	if decoded.SenderAddress != original.SenderAddress {
		t.Errorf("Sender address mismatch: expected %s, got %s", original.SenderAddress, decoded.SenderAddress)
	}

	if string(decoded.Payload) != string(original.Payload) {
		t.Error("Payload mismatch after round-trip")
	}
}

// TestPayload_Unmarshal tests unmarshaling payload
func TestPayload_Unmarshal(t *testing.T) {
	type TestData struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	original := TestData{Name: "test", Value: 42}
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	payload := types.Payload(data)

	var decoded TestData
	if err := payload.Unmarshal(&decoded); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	if decoded.Name != original.Name {
		t.Errorf("Expected name %s, got %s", original.Name, decoded.Name)
	}

	if decoded.Value != original.Value {
		t.Errorf("Expected value %d, got %d", original.Value, decoded.Value)
	}
}

// TestPayload_UnmarshalInvalidJSON tests unmarshaling invalid JSON
func TestPayload_UnmarshalInvalidJSON(t *testing.T) {
	payload := types.Payload([]byte(`{invalid json}`))

	var result map[string]any
	if err := payload.Unmarshal(&result); err == nil {
		t.Error("Should return error for invalid JSON")
	}
}

// TestPayload_UnmarshalEmpty tests unmarshaling empty payload
func TestPayload_UnmarshalEmpty(t *testing.T) {
	payload := types.Payload([]byte{})

	var result map[string]any
	if err := payload.Unmarshal(&result); err == nil {
		t.Error("Should return error for empty payload")
	}
}

// TestMessageTypes_Constants tests that message type constants are defined
func TestMessageTypes_Constants(t *testing.T) {
	constants := map[string]string{
		"RequestHeadersMsg":   types.RequestHeadersMsg,
		"SendBlockHeadersMsg": types.SendBlockHeadersMsg,
		"RequestBlockMsg":     types.RequestBlockMsg,
		"SendBlockMsg":        types.SendBlockMsg,
		"MempoolBroadcastMsg": types.MempoolBroadcastMsg,
		"BlockBroadcastMsg":   types.BlockBroadcastMsg,
		"CancelMiningMsg":     types.CancelMiningMsg,
	}

	for name, value := range constants {
		if value == "" {
			t.Errorf("Constant %s should not be empty", name)
		}
	}

	// Verify uniqueness
	seen := make(map[string]bool)
	for _, value := range constants {
		if seen[value] {
			t.Errorf("Duplicate message type constant: %s", value)
		}
		seen[value] = true
	}
}

// TestMessage_WithDifferentTypes tests messages with different types
func TestMessage_WithDifferentTypes(t *testing.T) {
	messageTypes := []string{
		types.RequestHeadersMsg,
		types.SendBlockHeadersMsg,
		types.RequestBlockMsg,
		types.SendBlockMsg,
		types.MempoolBroadcastMsg,
		types.BlockBroadcastMsg,
		types.CancelMiningMsg,
	}

	for _, msgType := range messageTypes {
		t.Run(msgType, func(t *testing.T) {
			msg := types.NewMessage(msgType, "127.0.0.1:8080", types.Payload{})

			if msg.Type != msgType {
				t.Errorf("Expected type %s, got %s", msgType, msg.Type)
			}

			// Verify it can be marshaled
			data := msg.Marshal()
			if len(data) == 0 {
				t.Error("Marshaled data should not be empty")
			}

			// Verify it can be unmarshaled
			var decoded types.Message
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Errorf("Failed to unmarshal: %v", err)
			}

			if decoded.Type != msgType {
				t.Errorf("Type mismatch after unmarshal: expected %s, got %s", msgType, decoded.Type)
			}
		})
	}
}

// TestMessage_WithComplexPayload tests message with complex payload
func TestMessage_WithComplexPayload(t *testing.T) {
	type ComplexData struct {
		ID     int64    `json:"id"`
		Name   string   `json:"name"`
		Tags   []string `json:"tags"`
		Nested struct {
			Value int `json:"value"`
		} `json:"nested"`
	}

	complexData := ComplexData{
		ID:   123,
		Name: "test",
		Tags: []string{"tag1", "tag2"},
	}
	complexData.Nested.Value = 456

	payload, err := json.Marshal(complexData)
	if err != nil {
		t.Fatalf("Failed to marshal complex data: %v", err)
	}

	msg := types.NewMessage(types.BlockBroadcastMsg, "127.0.0.1:8080", types.Payload(payload))

	// Marshal message
	data := msg.Marshal()

	// Unmarshal message
	var decoded types.Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	// Unmarshal payload
	var decodedData ComplexData
	if err := decoded.Payload.Unmarshal(&decodedData); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}

	// Verify
	if decodedData.ID != complexData.ID {
		t.Error("ID mismatch")
	}

	if decodedData.Name != complexData.Name {
		t.Error("Name mismatch")
	}

	if len(decodedData.Tags) != len(complexData.Tags) {
		t.Error("Tags length mismatch")
	}

	if decodedData.Nested.Value != complexData.Nested.Value {
		t.Error("Nested value mismatch")
	}
}

// TestMessage_EmptyPayload tests message with empty payload
func TestMessage_EmptyPayload(t *testing.T) {
	msg := types.NewMessage(types.CancelMiningMsg, "127.0.0.1:8080", types.Payload{})

	data := msg.Marshal()

	var decoded types.Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(decoded.Payload) != 0 {
		t.Error("Empty payload should remain empty after round-trip")
	}
}

// TestMessage_LargePayload tests message with large payload
func TestMessage_LargePayload(t *testing.T) {
	// Create a large payload (1MB)
	largeData := make([]byte, 1024*1024)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	msg := types.NewMessage(types.BlockBroadcastMsg, "127.0.0.1:8080", types.Payload(largeData))

	// Should be able to marshal large payload
	data := msg.Marshal()
	if len(data) == 0 {
		t.Error("Should be able to marshal large payload")
	}

	// Should be able to unmarshal
	var decoded types.Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Errorf("Should be able to unmarshal large payload: %v", err)
	}
}

// TestPayload_UnmarshalToStruct tests unmarshaling to struct
func TestPayload_UnmarshalToStruct(t *testing.T) {
	type Block struct {
		ID     int64  `json:"id"`
		Hash   string `json:"hash"`
		Height int    `json:"height"`
	}

	original := Block{
		ID:     1,
		Hash:   "abc123",
		Height: 100,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	payload := types.Payload(data)

	var decoded Block
	if err := payload.Unmarshal(&decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Error("ID mismatch")
	}

	if decoded.Hash != original.Hash {
		t.Error("Hash mismatch")
	}

	if decoded.Height != original.Height {
		t.Error("Height mismatch")
	}
}

// TestPayload_UnmarshalToSlice tests unmarshaling to slice
func TestPayload_UnmarshalToSlice(t *testing.T) {
	original := []int{1, 2, 3, 4, 5}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	payload := types.Payload(data)

	var decoded []int
	if err := payload.Unmarshal(&decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(decoded) != len(original) {
		t.Errorf("Expected length %d, got %d", len(original), len(decoded))
	}

	for i, v := range original {
		if decoded[i] != v {
			t.Errorf("Element %d mismatch: expected %d, got %d", i, v, decoded[i])
		}
	}
}

// TestPayload_UnmarshalToMap tests unmarshaling to map
func TestPayload_UnmarshalToMap(t *testing.T) {
	original := map[string]any{
		"key1": "value1",
		"key2": 123,
		"key3": true,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	payload := types.Payload(data)

	var decoded map[string]any
	if err := payload.Unmarshal(&decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded["key1"] != original["key1"] {
		t.Error("key1 mismatch")
	}

	// JSON numbers are decoded as float64
	if decoded["key2"].(float64) != float64(123) {
		t.Error("key2 mismatch")
	}

	if decoded["key3"] != original["key3"] {
		t.Error("key3 mismatch")
	}
}
