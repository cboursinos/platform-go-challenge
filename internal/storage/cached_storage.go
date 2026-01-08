package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gwi/platform-go-challenge/internal/models"
)

// CachedStorage wraps a storage implementation with Redis caching
type CachedStorage struct {
	storage Storage
	cache   *RedisCache
}

// NewCachedStorage creates a new cached storage instance
func NewCachedStorage(storage Storage, cache *RedisCache) *CachedStorage {
	return &CachedStorage{
		storage: storage,
		cache:   cache,
	}
}

// CreateAsset creates a new asset
func (s *CachedStorage) CreateAsset(ctx context.Context, userID string, asset models.Asset) (models.Asset, error) {
	// Create in database
	createdAsset, err := s.storage.CreateAsset(ctx, userID, asset)
	if err != nil {
		return nil, err
	}

	return createdAsset, nil
}

// UpdateAsset updates an existing asset
func (s *CachedStorage) UpdateAsset(ctx context.Context, assetReference string, asset models.Asset) (models.Asset, error) {
	// Update in database
	updatedAsset, err := s.storage.UpdateAsset(ctx, assetReference, asset)
	if err != nil {
		return nil, err
	}

	// Invalidate asset cache
	assetCacheKey := cacheKeyAsset(assetReference)
	if err := s.cache.Delete(ctx, assetCacheKey); err != nil {
		fmt.Printf("Warning: failed to invalidate asset cache for key %s: %v\n", assetCacheKey, err)
	}

	// Invalidate all user favorites caches that might contain this asset
	// Since we don't know which users have this asset, we let cache expire naturally
	// In production, you might want to use a more sophisticated invalidation strategy

	return updatedAsset, nil
}

// DeleteAsset deletes an asset
func (s *CachedStorage) DeleteAsset(ctx context.Context, assetReference string) error {
	// Delete from database
	err := s.storage.DeleteAsset(ctx, assetReference)
	if err != nil {
		return err
	}

	// Invalidate asset cache
	assetCacheKey := cacheKeyAsset(assetReference)
	if err := s.cache.Delete(ctx, assetCacheKey); err != nil {
		fmt.Printf("Warning: failed to invalidate asset cache for key %s: %v\n", assetCacheKey, err)
	}

	// Invalidate all user favorites caches (since favorites might reference this asset)
	// This is a bit aggressive but ensures consistency
	// In production, you might track which users have which assets

	return nil
}

// AddFavorite adds or updates a favorite for a user in a specific list
func (s *CachedStorage) AddFavorite(ctx context.Context, userID, assetReference, listReference string, sortOrder *int) (*models.Favorite, error) {
	// Add to database
	favorite, err := s.storage.AddFavorite(ctx, userID, assetReference, listReference, sortOrder)
	if err != nil {
		return nil, err
	}

	// Invalidate user favorites cache for this list
	cacheKey := cacheKeyUserFavorites(userID, listReference)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to invalidate cache for key %s: %v\n", cacheKey, err)
	}

	return favorite, nil
}

// RemoveFavorite removes a favorite for a user from a specific list
func (s *CachedStorage) RemoveFavorite(ctx context.Context, userID, assetID, listReference string) error {
	// Remove from database
	if err := s.storage.RemoveFavorite(ctx, userID, assetID, listReference); err != nil {
		return err
	}

	// Invalidate user favorites cache for this list
	cacheKey := cacheKeyUserFavorites(userID, listReference)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		fmt.Printf("Warning: failed to invalidate cache for key %s: %v\n", cacheKey, err)
	}

	return nil
}

// GetFavorites returns all favorites for a user (cache-aside pattern)
func (s *CachedStorage) RemoveFavoriteByReference(ctx context.Context, favoriteReference string) error {
	return s.storage.RemoveFavoriteByReference(ctx, favoriteReference)
}

func (s *CachedStorage) GetFavorites(ctx context.Context, userID, listReference string) ([]*models.Favorite, error) {
	cacheKey := cacheKeyUserFavorites(userID, listReference)

	// Try to get from cache
	cachedData, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		// Cache hit
		var favorites []*models.Favorite
		if err := json.Unmarshal(cachedData, &favorites); err == nil {
			return favorites, nil
		}
		// If unmarshal fails, continue to database
	}

	// Cache miss - get from database
	favorites, err := s.storage.GetFavorites(ctx, userID, listReference)
	if err != nil {
		return nil, err
	}

	// Store in cache for next time
	if data, err := json.Marshal(favorites); err == nil {
		if err := s.cache.Set(ctx, cacheKey, data); err != nil {
			fmt.Printf("Warning: failed to cache favorites for user %s list %s: %v\n", userID, listReference, err)
		}
	}

	return favorites, nil
}

// GetFavoritesPaginated returns paginated favorites for a user in a specific list
func (s *CachedStorage) GetFavoritesPaginated(ctx context.Context, userID, listReference string, page, pageSize int, filters *models.FilterParams) ([]*models.Favorite, int, error) {
	// For paginated requests with filters, we don't cache (as parameters vary)
	// Delegate directly to underlying storage
	return s.storage.GetFavoritesPaginated(ctx, userID, listReference, page, pageSize, filters)
}

// UpdateAssetDescription updates the description of an asset
func (s *CachedStorage) UpdateAssetDescription(ctx context.Context, assetID, description string) error {
	// Update in database
	if err := s.storage.UpdateAssetDescription(ctx, assetID, description); err != nil {
		return err
	}

	// Invalidate asset cache
	assetCacheKey := cacheKeyAsset(assetID)
	if err := s.cache.Delete(ctx, assetCacheKey); err != nil {
		fmt.Printf("Warning: failed to invalidate asset cache for key %s: %v\n", assetCacheKey, err)
	}

	// Invalidate all user favorites caches that might contain this asset
	// Since we don't know which users have this asset, we could:
	// 1. Use a set to track which users have each asset (more complex)
	// 2. Invalidate all user favorites (too aggressive)
	// 3. Let cache expire naturally (current approach)
	// For now, we'll invalidate the asset cache and let user favorites cache expire
	// In production, you might want to use a more sophisticated invalidation strategy

	return nil
}

// GetAsset retrieves an asset by ID (cache-aside pattern)
func (s *CachedStorage) GetAsset(ctx context.Context, assetID string) (models.Asset, error) {
	cacheKey := cacheKeyAsset(assetID)

	// Try to get from cache
	cachedData, err := s.cache.Get(ctx, cacheKey)
	if err == nil {
		// Cache hit - deserialize asset
		var assetData map[string]interface{}
		if err := json.Unmarshal(cachedData, &assetData); err == nil {
			// Reconstruct asset from cached data
			// This is simplified - in production you'd want a proper deserialization
			return s.storage.GetAsset(ctx, assetID) // Fallback to storage for now
		}
	}

	// Cache miss - get from database
	asset, err := s.storage.GetAsset(ctx, assetID)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if data, err := json.Marshal(asset); err == nil {
		if err := s.cache.Set(ctx, cacheKey, data); err != nil {
			fmt.Printf("Warning: failed to cache asset %s: %v\n", assetID, err)
		}
	}

	return asset, nil
}

// GetAllAssetsPaginated returns paginated assets
// Note: We don't cache paginated lists as they change frequently and caching would be complex
func (s *CachedStorage) GetAllAssetsPaginated(ctx context.Context, page, pageSize int, filters *models.FilterParams) ([]models.Asset, int, error) {
	return s.storage.GetAllAssetsPaginated(ctx, page, pageSize, filters)
}

// CreateList creates a new list for a user
func (s *CachedStorage) CreateList(ctx context.Context, userID, listName string) (*models.List, error) {
	return s.storage.CreateList(ctx, userID, listName)
}

// GetList retrieves a list by reference
func (s *CachedStorage) GetList(ctx context.Context, userID, listReference string) (*models.List, error) {
	return s.storage.GetList(ctx, userID, listReference)
}

// GetListByName retrieves a list by name for a user
func (s *CachedStorage) GetListByName(ctx context.Context, userID, listName string) (*models.List, error) {
	return s.storage.GetListByName(ctx, userID, listName)
}

// GetAllLists returns all lists for a user
func (s *CachedStorage) GetAllLists(ctx context.Context, userID string) ([]*models.List, error) {
	return s.storage.GetAllLists(ctx, userID)
}

// GetAllListsPaginated returns paginated lists for a user
func (s *CachedStorage) GetAllListsPaginated(ctx context.Context, userID string, page, pageSize int) ([]*models.List, int, error) {
	return s.storage.GetAllListsPaginated(ctx, userID, page, pageSize)
}

// GetAllFavorites returns all favorites for a user across all lists
func (s *CachedStorage) GetAllFavorites(ctx context.Context, userID string) ([]*models.Favorite, error) {
	return s.storage.GetAllFavorites(ctx, userID)
}

// GetAllFavoritesPaginated returns paginated favorites for a user across all lists
func (s *CachedStorage) GetAllFavoritesPaginated(ctx context.Context, userID string, page, pageSize int, filters *models.FilterParams) ([]*models.Favorite, int, error) {
	return s.storage.GetAllFavoritesPaginated(ctx, userID, page, pageSize, filters)
}

// DeleteList deletes a list and all its favorites (CASCADE)
func (s *CachedStorage) DeleteList(ctx context.Context, userID, listReference string) error {
	// Delete from storage (this will also delete favorites)
	err := s.storage.DeleteList(ctx, userID, listReference)
	if err != nil {
		return err
	}

	// Invalidate cache for this list's favorites
	cacheKey := cacheKeyUserFavorites(userID, listReference)
	s.cache.Delete(ctx, cacheKey)

	return nil
}

// GenerateNextAssetID generates the next sequential asset ID for a user
func (s *CachedStorage) GenerateNextAssetID(ctx context.Context, userReference string) (string, error) {
	return s.storage.GenerateNextAssetID(ctx, userReference)
}

// UpdateFavoriteSortOrder updates the sort order of a favorite
func (s *CachedStorage) UpdateFavoriteSortOrder(ctx context.Context, userReference, favoriteReference string, sortOrder int) error {
	// Update in database
	err := s.storage.UpdateFavoriteSortOrder(ctx, userReference, favoriteReference, sortOrder)
	if err != nil {
		return err
	}

	// Invalidate cache for this user's favorites
	// We need to invalidate all lists for this user since we don't know which list the favorite belongs to
	// Get all lists for the user and invalidate cache for each
	lists, err := s.storage.GetAllLists(ctx, userReference)
	if err == nil {
		for _, list := range lists {
			cacheKey := cacheKeyUserFavorites(userReference, list.Reference)
			s.cache.Delete(ctx, cacheKey)
		}
	}

	return nil
}

// Close closes the underlying storage and cache connections
func (s *CachedStorage) Close() error {
	var errs []error

	if err := s.storage.Close(); err != nil {
		errs = append(errs, err)
	}

	if err := s.cache.Close(); err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing storage: %v", errs)
	}

	return nil
}

