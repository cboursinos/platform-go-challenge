package storage

import (
	"testing"
	"time"
)

func TestRedisCache_GetSet(t *testing.T) {
	skipIfRedisUnavailable(t)

	addr := getTestRedisAddr()
	password := getTestRedisPassword()
	cache, err := NewRedisCache(addr, password, 1, 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create Redis cache: %v", err)
	}
	defer cache.Close()

	key := "test:key:1"
	value := []byte("test value")

	// Set value
	err = cache.Set(key, value)
	if err != nil {
		t.Fatalf("Failed to set cache: %v", err)
	}

	// Get value
	retrieved, err := cache.Get(key)
	if err != nil {
		t.Fatalf("Failed to get cache: %v", err)
	}

	if string(retrieved) != string(value) {
		t.Errorf("Expected value %s, got %s", string(value), string(retrieved))
	}
}

func TestRedisCache_GetMiss(t *testing.T) {
	skipIfRedisUnavailable(t)

	addr := getTestRedisAddr()
	password := getTestRedisPassword()
	cache, err := NewRedisCache(addr, password, 1, 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create Redis cache: %v", err)
	}
	defer cache.Close()

	// Try to get non-existent key
	_, err = cache.Get("test:key:nonexistent")
	if err != ErrCacheMiss {
		t.Errorf("Expected ErrCacheMiss, got %v", err)
	}
}

func TestRedisCache_Delete(t *testing.T) {
	skipIfRedisUnavailable(t)

	addr := getTestRedisAddr()
	password := getTestRedisPassword()
	cache, err := NewRedisCache(addr, password, 1, 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create Redis cache: %v", err)
	}
	defer cache.Close()

	key := "test:key:delete"
	value := []byte("test value")

	// Set value
	cache.Set(key, value)

	// Verify it exists
	_, err = cache.Get(key)
	if err != nil {
		t.Fatalf("Key should exist: %v", err)
	}

	// Delete value
	err = cache.Delete(key)
	if err != nil {
		t.Fatalf("Failed to delete cache: %v", err)
	}

	// Verify it's gone
	_, err = cache.Get(key)
	if err != ErrCacheMiss {
		t.Errorf("Expected ErrCacheMiss after deletion, got %v", err)
	}
}

func TestRedisCache_SetWithTTL(t *testing.T) {
	skipIfRedisUnavailable(t)

	addr := getTestRedisAddr()
	password := getTestRedisPassword()
	cache, err := NewRedisCache(addr, password, 1, 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create Redis cache: %v", err)
	}
	defer cache.Close()

	key := "test:key:ttl"
	value := []byte("test value")
	ttl := 2 * time.Second

	// Set value with short TTL
	err = cache.SetWithTTL(key, value, ttl)
	if err != nil {
		t.Fatalf("Failed to set cache with TTL: %v", err)
	}

	// Verify it exists
	_, err = cache.Get(key)
	if err != nil {
		t.Fatalf("Key should exist: %v", err)
	}

	// Wait for expiration
	time.Sleep(3 * time.Second)

	// Verify it's expired
	_, err = cache.Get(key)
	if err != ErrCacheMiss {
		t.Errorf("Expected ErrCacheMiss after expiration, got %v", err)
	}
}

func TestRedisCache_DeletePattern(t *testing.T) {
	skipIfRedisUnavailable(t)

	addr := getTestRedisAddr()
	password := getTestRedisPassword()
	cache, err := NewRedisCache(addr, password, 1, 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create Redis cache: %v", err)
	}
	defer cache.Close()

	// Set multiple keys with pattern
	keys := []string{
		"test:pattern:key1",
		"test:pattern:key2",
		"test:pattern:key3",
		"test:other:key1",
	}

	for _, key := range keys {
		cache.Set(key, []byte("value"))
	}

	// Delete pattern
	err = cache.DeletePattern("test:pattern:*")
	if err != nil {
		t.Fatalf("Failed to delete pattern: %v", err)
	}

	// Verify pattern keys are deleted
	for _, key := range keys[:3] {
		_, err = cache.Get(key)
		if err != ErrCacheMiss {
			t.Errorf("Expected key %s to be deleted, but it exists", key)
		}
	}

	// Verify other key still exists
	_, err = cache.Get(keys[3])
	if err != nil {
		t.Errorf("Expected key %s to still exist, but it's gone", keys[3])
	}
}

func TestRedisCache_ConcurrentAccess(t *testing.T) {
	skipIfRedisUnavailable(t)

	addr := getTestRedisAddr()
	password := getTestRedisPassword()
	cache, err := NewRedisCache(addr, password, 1, 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to create Redis cache: %v", err)
	}
	defer cache.Close()

	done := make(chan bool, 10)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(id int) {
			key := "test:concurrent:" + string(rune('0'+id))
			value := []byte("value")
			cache.Set(key, value)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all keys exist
	for i := 0; i < 10; i++ {
		key := "test:concurrent:" + string(rune('0'+i))
		_, err := cache.Get(key)
		if err != nil {
			t.Errorf("Expected key %s to exist, got error: %v", key, err)
		}
	}
}

func TestRedisCache_KeyGeneration(t *testing.T) {
	userID := "user123"
	listReference := "list_user123_default"
	assetID := "asset456"

	userKey := cacheKeyUserFavorites(userID, listReference)
	expectedUserKey := "favorites:user:user123:list:list_user123_default"
	if userKey != expectedUserKey {
		t.Errorf("Expected cache key %s, got %s", expectedUserKey, userKey)
	}

	assetKey := cacheKeyAsset(assetID)
	expectedAssetKey := "asset:asset456"
	if assetKey != expectedAssetKey {
		t.Errorf("Expected cache key %s, got %s", expectedAssetKey, assetKey)
	}
}

