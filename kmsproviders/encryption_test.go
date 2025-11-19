package kmsproviders

import (
	"testing"
	"time"
)

func TestMaskKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "standard API key",
			key:      "sk-1234567890abcdefghijklmnopqrstuvwxyz",
			expected: "sk-1...wxyz",
		},
		{
			name:     "short key",
			key:      "short",
			expected: "***",
		},
		{
			name:     "exactly 8 chars",
			key:      "12345678",
			expected: "1234...5678",
		},
		{
			name:     "empty key",
			key:      "",
			expected: "***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskKey(tt.key)
			if result != tt.expected {
				t.Errorf("MaskKey(%q) = %q, want %q", tt.key, result, tt.expected)
			}
		})
	}
}

func TestProviderKeyCache(t *testing.T) {
	cache := &providerKeyCache{
		entries: make(map[string]*cacheEntry),
	}

	// Test adding and retrieving from cache
	t.Run("add and retrieve", func(t *testing.T) {
		key := "test-cache-key"
		value := "test-decrypted-value"

		cache.mu.Lock()
		cache.entries[key] = &cacheEntry{
			decryptedKey: value,
			expiresAt:    time.Now().Add(1 * time.Hour),
		}
		cache.mu.Unlock()

		cache.mu.RLock()
		entry, exists := cache.entries[key]
		cache.mu.RUnlock()

		if !exists {
			t.Fatal("Expected entry to exist in cache")
		}
		if entry.decryptedKey != value {
			t.Errorf("Expected cached value %q, got %q", value, entry.decryptedKey)
		}
	})

	// Test expiration
	t.Run("expiration", func(t *testing.T) {
		key := "expired-key"
		value := "expired-value"

		cache.mu.Lock()
		cache.entries[key] = &cacheEntry{
			decryptedKey: value,
			expiresAt:    time.Now().Add(-1 * time.Hour), // Already expired
		}
		cache.mu.Unlock()

		cache.mu.RLock()
		entry, exists := cache.entries[key]
		cache.mu.RUnlock()

		if !exists {
			t.Fatal("Expected entry to exist in cache")
		}

		// Check if expired
		if !time.Now().After(entry.expiresAt) {
			t.Error("Expected entry to be expired")
		}
	})
}

func TestMakeCacheKey(t *testing.T) {
	m := &ProviderKeyManager{
		cache: &providerKeyCache{
			entries: make(map[string]*cacheEntry),
		},
	}

	// Test that same inputs produce same key
	key1 := m.makeCacheKey("user123", "openai")
	key2 := m.makeCacheKey("user123", "openai")
	if key1 != key2 {
		t.Error("Expected same cache key for identical inputs")
	}

	// Test that different inputs produce different keys
	key3 := m.makeCacheKey("user123", "anthropic")
	if key1 == key3 {
		t.Error("Expected different cache keys for different providers")
	}

	key4 := m.makeCacheKey("user456", "openai")
	if key1 == key4 {
		t.Error("Expected different cache keys for different users")
	}
}

func TestInvalidateCache(t *testing.T) {
	m := &ProviderKeyManager{
		cache: &providerKeyCache{
			entries: make(map[string]*cacheEntry),
		},
	}

	// Add an entry
	userID := "test-user"
	provider := "openai"
	cacheKey := m.makeCacheKey(userID, provider)

	m.cache.mu.Lock()
	m.cache.entries[cacheKey] = &cacheEntry{
		decryptedKey: "test-key",
		expiresAt:    time.Now().Add(1 * time.Hour),
	}
	m.cache.mu.Unlock()

	// Verify it exists
	m.cache.mu.RLock()
	_, exists := m.cache.entries[cacheKey]
	m.cache.mu.RUnlock()
	if !exists {
		t.Fatal("Expected entry to exist before invalidation")
	}

	// Invalidate
	m.InvalidateCache(userID, provider)

	// Verify it's gone
	m.cache.mu.RLock()
	_, exists = m.cache.entries[cacheKey]
	m.cache.mu.RUnlock()
	if exists {
		t.Error("Expected entry to be removed after invalidation")
	}
}

func TestCleanExpiredCache(t *testing.T) {
	m := &ProviderKeyManager{
		cache: &providerKeyCache{
			entries: make(map[string]*cacheEntry),
		},
	}

	// Add expired and non-expired entries
	m.cache.mu.Lock()
	m.cache.entries["expired1"] = &cacheEntry{
		decryptedKey: "test1",
		expiresAt:    time.Now().Add(-1 * time.Hour),
	}
	m.cache.entries["expired2"] = &cacheEntry{
		decryptedKey: "test2",
		expiresAt:    time.Now().Add(-30 * time.Minute),
	}
	m.cache.entries["valid"] = &cacheEntry{
		decryptedKey: "test3",
		expiresAt:    time.Now().Add(1 * time.Hour),
	}
	m.cache.mu.Unlock()

	// Clean expired entries
	m.CleanExpiredCache()

	// Verify only valid entry remains
	m.cache.mu.RLock()
	defer m.cache.mu.RUnlock()

	if len(m.cache.entries) != 1 {
		t.Errorf("Expected 1 entry after cleaning, got %d", len(m.cache.entries))
	}

	if _, exists := m.cache.entries["valid"]; !exists {
		t.Error("Expected valid entry to remain after cleaning")
	}
}
