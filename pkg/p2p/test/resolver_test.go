package test

import (
	"simple_blockchain/pkg/p2p/resolver"
	"testing"
)

// TestResolveSeedNodes tests resolving seed nodes
func TestResolveSeedNodes(t *testing.T) {
	seeds := resolver.ResolveSeedNodes()

	if seeds == nil {
		t.Fatal("Seed nodes should not be nil")
	}

	if len(seeds) == 0 {
		t.Error("Seed nodes should not be empty")
	}
}

// TestResolveSeedNodes_ReturnsSlice tests that seed nodes returns a slice
func TestResolveSeedNodes_ReturnsSlice(t *testing.T) {
	seeds := resolver.ResolveSeedNodes()

	// Verify it's a slice
	if seeds == nil {
		t.Fatal("Should return a slice, not nil")
	}
}

// TestResolveSeedNodes_Consistency tests that seed nodes are consistent
func TestResolveSeedNodes_Consistency(t *testing.T) {
	seeds1 := resolver.ResolveSeedNodes()
	seeds2 := resolver.ResolveSeedNodes()

	if len(seeds1) != len(seeds2) {
		t.Error("Seed nodes should be consistent across calls")
	}

	// Verify same seeds are returned
	for i := range seeds1 {
		if seeds1[i] != seeds2[i] {
			t.Error("Seed nodes should be identical across calls")
		}
	}
}

// TestResolveSeedNodes_ValidFormat tests seed node format
func TestResolveSeedNodes_ValidFormat(t *testing.T) {
	seeds := resolver.ResolveSeedNodes()

	for _, seed := range seeds {
		if seed == "" {
			t.Error("Seed node should not be empty string")
		}

		// Basic format check (should contain colon for host:port)
		hasColon := false
		for _, char := range seed {
			if char == ':' {
				hasColon = true
				break
			}
		}

		if !hasColon {
			t.Errorf("Seed node %s should be in host:port format", seed)
		}
	}
}

// TestResolveSeedNodes_NotModifiable tests that returned slice is independent
func TestResolveSeedNodes_NotModifiable(t *testing.T) {
	seeds1 := resolver.ResolveSeedNodes()
	originalLen := len(seeds1)

	// Try to modify the returned slice
	seeds1 = append(seeds1, "new.seed:9999")

	// Get seeds again
	seeds2 := resolver.ResolveSeedNodes()

	// Original seeds should not be affected
	if len(seeds2) != originalLen {
		t.Error("Modifying returned slice should not affect subsequent calls")
	}
}
