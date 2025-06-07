package tests

import (
	"testing"
	"time"

	"github.com/gofiber/storage/memory/v2"
)

func TestBasicCache(t *testing.T) {
	// Create a Fiber memory store
	store := memory.New()
	defer store.Close()

	// Test basic set and get
	key := "test_key"
	value := []byte("test_value")

	err := store.Set(key, value, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to set: %v", err)
	}

	retrieved, err := store.Get(key)
	if err != nil {
		t.Fatalf("Failed to get: %v", err)
	}

	if string(retrieved) != string(value) {
		t.Errorf("Expected %s, got %s", value, retrieved)
	}

	t.Log("Basic cache test passed")
}
