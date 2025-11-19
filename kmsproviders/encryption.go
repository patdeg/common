package kmsproviders

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log/slog"
	"sync"
	"time"

	kms "cloud.google.com/go/kms/apiv1"
	"cloud.google.com/go/kms/apiv1/kmspb"
)

// ProviderKeyManager handles encryption/decryption of third-party provider API keys
// (OpenAI, Groq, Anthropic, Google) using Google Cloud KMS for enterprise-grade security.
type ProviderKeyManager struct {
	kmsClient *kms.KeyManagementClient
	keyName   string // Full KMS key resource name
	cache     *providerKeyCache
}

// ProviderKeySource indicates where the provider key came from
type ProviderKeySource string

const (
	ProviderKeySourceTransient ProviderKeySource = "transient" // From Authorization header (dual-key)
	ProviderKeySourceCached    ProviderKeySource = "cached"    // From in-memory cache
	ProviderKeySourceStored    ProviderKeySource = "stored"    // From KMS-encrypted Datastore
)

// providerKeyCache holds decrypted provider keys in memory to avoid repeated KMS calls
type providerKeyCache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
}

type cacheEntry struct {
	decryptedKey string
	expiresAt    time.Time
}

// NewProviderKeyManager creates a new provider key manager with KMS encryption
func NewProviderKeyManager(ctx context.Context, projectID, location, keyRing, keyID string) (*ProviderKeyManager, error) {
	client, err := kms.NewKeyManagementClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create KMS client: %w", err)
	}

	keyName := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s",
		projectID, location, keyRing, keyID)

	return &ProviderKeyManager{
		kmsClient: client,
		keyName:   keyName,
		cache: &providerKeyCache{
			entries: make(map[string]*cacheEntry),
		},
	}, nil
}

// EncryptProviderKey encrypts a provider API key using Google Cloud KMS
func (m *ProviderKeyManager) EncryptProviderKey(ctx context.Context, providerKey string) (string, error) {
	// Use KMS to encrypt the provider key
	req := &kmspb.EncryptRequest{
		Name:      m.keyName,
		Plaintext: []byte(providerKey),
	}

	result, err := m.kmsClient.Encrypt(ctx, req)
	if err != nil {
		slog.ErrorContext(ctx, "KMS encryption failed", "error", err)
		return "", fmt.Errorf("kms encryption failed: %w", err)
	}

	// Return base64-encoded ciphertext for storage in Datastore
	encrypted := base64.StdEncoding.EncodeToString(result.Ciphertext)

	slog.InfoContext(ctx, "Provider key encrypted with KMS",
		"kms_key", m.keyName,
		"ciphertext_length", len(encrypted))

	return encrypted, nil
}

// DecryptProviderKey decrypts a provider API key using Google Cloud KMS
// Returns the decrypted key from cache if available, otherwise decrypts with KMS
func (m *ProviderKeyManager) DecryptProviderKey(ctx context.Context, userID, provider, encryptedKey string) (string, ProviderKeySource, error) {
	// Check cache first
	cacheKey := m.makeCacheKey(userID, provider)
	if cachedKey := m.getFromCache(cacheKey); cachedKey != "" {
		slog.InfoContext(ctx, "Provider key retrieved from cache",
			"user_id", userID,
			"provider", provider,
			"source", "cached")
		return cachedKey, ProviderKeySourceCached, nil
	}

	// Decode from base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedKey)
	if err != nil {
		return "", "", fmt.Errorf("base64 decode failed: %w", err)
	}

	// Use KMS to decrypt
	req := &kmspb.DecryptRequest{
		Name:       m.keyName,
		Ciphertext: ciphertext,
	}

	result, err := m.kmsClient.Decrypt(ctx, req)
	if err != nil {
		slog.ErrorContext(ctx, "KMS decryption failed", "error", err, "user_id", userID, "provider", provider)
		return "", "", fmt.Errorf("kms decryption failed: %w", err)
	}

	decryptedKey := string(result.Plaintext)

	// Cache the decrypted key for this session (15 minutes)
	m.addToCache(cacheKey, decryptedKey, 15*time.Minute)

	slog.InfoContext(ctx, "Provider key decrypted with KMS and cached",
		"user_id", userID,
		"provider", provider,
		"source", "stored",
		"cache_duration", "15m")

	return decryptedKey, ProviderKeySourceStored, nil
}

// makeCacheKey creates a cache key from userID and provider
func (m *ProviderKeyManager) makeCacheKey(userID, provider string) string {
	// Use SHA256 hash to avoid storing user IDs in cache keys
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", userID, provider)))
	return base64.URLEncoding.EncodeToString(hash[:])
}

// getFromCache retrieves a decrypted key from cache if not expired
func (m *ProviderKeyManager) getFromCache(cacheKey string) string {
	m.cache.mu.RLock()
	defer m.cache.mu.RUnlock()

	entry, exists := m.cache.entries[cacheKey]
	if !exists {
		return ""
	}

	// Check if expired
	if time.Now().After(entry.expiresAt) {
		return ""
	}

	return entry.decryptedKey
}

// addToCache stores a decrypted key in cache with expiration
func (m *ProviderKeyManager) addToCache(cacheKey, decryptedKey string, duration time.Duration) {
	m.cache.mu.Lock()
	defer m.cache.mu.Unlock()

	m.cache.entries[cacheKey] = &cacheEntry{
		decryptedKey: decryptedKey,
		expiresAt:    time.Now().Add(duration),
	}
}

// InvalidateCache removes a specific provider key from cache
func (m *ProviderKeyManager) InvalidateCache(userID, provider string) {
	cacheKey := m.makeCacheKey(userID, provider)

	m.cache.mu.Lock()
	defer m.cache.mu.Unlock()

	delete(m.cache.entries, cacheKey)
}

// CleanExpiredCache removes expired entries from cache (should be called periodically)
func (m *ProviderKeyManager) CleanExpiredCache() {
	m.cache.mu.Lock()
	defer m.cache.mu.Unlock()

	now := time.Now()
	for key, entry := range m.cache.entries {
		if now.After(entry.expiresAt) {
			delete(m.cache.entries, key)
		}
	}
}

// Close closes the KMS client
func (m *ProviderKeyManager) Close() error {
	return m.kmsClient.Close()
}

// MaskKey returns a masked version of an API key for logging (first 4 + last 4 chars)
func MaskKey(key string) string {
	if len(key) < 8 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
