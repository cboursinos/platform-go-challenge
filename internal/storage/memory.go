package storage

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gwi/platform-go-challenge/internal/models"
)

// MemoryStorage is a thread-safe in-memory storage implementation
// Implements the Storage interface for backward compatibility
type MemoryStorage struct {
	mu        sync.RWMutex
	lists     map[string]map[string]*models.List      // userID -> listReference -> List
	favorites map[string]map[string]map[string]*models.Favorite // userID -> listReference -> assetID -> Favorite
	assets    map[string]models.Asset                            // assetID -> Asset
}

// NewMemoryStorage creates a new in-memory storage instance
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		lists:     make(map[string]map[string]*models.List),
		favorites: make(map[string]map[string]map[string]*models.Favorite),
		assets:    make(map[string]models.Asset),
	}
}

// CreateList creates a new list for a user
func (s *MemoryStorage) CreateList(userID, listName string) (*models.List, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	listReference := fmt.Sprintf("list_%s_%s", userID, listName)

	// Initialize user's lists map if needed
	if s.lists[userID] == nil {
		s.lists[userID] = make(map[string]*models.List)
	}

	// Check if list already exists
	if existing, exists := s.lists[userID][listReference]; exists {
		return existing, nil
	}

	// Create new list
	now := time.Now()
	list := &models.List{
		ID:        len(s.lists[userID]) + 1,
		Reference: listReference,
		UserID:    userID,
		Name:      listName,
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.lists[userID][listReference] = list
	return list, nil
}

// GetList retrieves a list by reference
func (s *MemoryStorage) GetList(userID, listReference string) (*models.List, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.lists[userID] == nil {
		return nil, fmt.Errorf("list with reference %s not found", listReference)
	}

	list, exists := s.lists[userID][listReference]
	if !exists {
		return nil, fmt.Errorf("list with reference %s not found", listReference)
	}

	return list, nil
}

// GetListByName retrieves a list by name for a user
func (s *MemoryStorage) GetListByName(userID, listName string) (*models.List, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.lists[userID] == nil {
		return nil, fmt.Errorf("list with name %s not found", listName)
	}

	for _, list := range s.lists[userID] {
		if list.Name == listName {
			return list, nil
		}
	}

	return nil, fmt.Errorf("list with name %s not found", listName)
}

// GetAllLists returns all lists for a user
func (s *MemoryStorage) GetAllLists(userID string) ([]*models.List, error) {
	s.mu.RLock()
	listsMap := s.lists[userID]
	hasLists := listsMap != nil && len(listsMap) > 0
	s.mu.RUnlock()

	if !hasLists {
		// Create default list if none exist
		defaultList, err := s.CreateList(userID, "default")
		if err != nil {
			return nil, err
		}
		return []*models.List{defaultList}, nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	lists := make([]*models.List, 0, len(s.lists[userID]))
	for _, list := range s.lists[userID] {
		lists = append(lists, list)
	}

	return lists, nil
}

// DeleteList deletes a list and all its favorites (CASCADE)
func (s *MemoryStorage) DeleteList(userID, listReference string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.lists[userID] == nil {
		return fmt.Errorf("list with reference %s not found", listReference)
	}

	if _, exists := s.lists[userID][listReference]; !exists {
		return fmt.Errorf("list with reference %s not found", listReference)
	}

	// Delete list
	delete(s.lists[userID], listReference)

	// Delete all favorites in this list
	if s.favorites[userID] != nil && s.favorites[userID][listReference] != nil {
		delete(s.favorites[userID], listReference)
	}

	return nil
}

// CreateAsset creates a new asset
func (s *MemoryStorage) CreateAsset(userID string, asset models.Asset) (models.Asset, error) {
	if asset == nil {
		return nil, fmt.Errorf("asset is required")
	}

	assetReference := asset.GetID()
	if assetReference == "" {
		return nil, fmt.Errorf("asset ID (reference) is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	asset.SetUpdatedAt(now)

	// Check if asset already exists
	if _, exists := s.assets[assetReference]; exists {
		return nil, fmt.Errorf("asset with reference %s already exists", assetReference)
	}

	// Store the asset
	s.assets[assetReference] = asset
	return asset, nil
}

// UpdateAsset updates an existing asset
func (s *MemoryStorage) UpdateAsset(assetReference string, asset models.Asset) (models.Asset, error) {
	if asset == nil {
		return nil, fmt.Errorf("asset is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if asset exists
	if _, exists := s.assets[assetReference]; !exists {
		return nil, ErrAssetNotFound
	}

	now := time.Now()
	asset.SetUpdatedAt(now)

	// Update the asset
	s.assets[assetReference] = asset

	// Update asset in all user favorites
	for userID, lists := range s.favorites {
		for listReference := range lists {
			if fav, exists := s.favorites[userID][listReference][assetReference]; exists {
				fav.Asset = asset
				fav.UpdatedAt = now
			}
		}
	}

	return asset, nil
}

// DeleteAsset deletes an asset
func (s *MemoryStorage) DeleteAsset(assetReference string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if asset exists
	if _, exists := s.assets[assetReference]; !exists {
		return ErrAssetNotFound
	}

	// Delete asset
	delete(s.assets, assetReference)

	// Delete all favorites that reference this asset
	for userIDKey, lists := range s.favorites {
		for listReference := range lists {
			if _, exists := s.favorites[userIDKey][listReference][assetReference]; exists {
				delete(s.favorites[userIDKey][listReference], assetReference)
			}
		}
	}

	return nil
}

// AddFavorite adds or updates a favorite for a user in a specific list
// assetReference is the asset reference (e.g., "user1_favorite1")
// sortOrder is optional - if nil, will use next sequential value
func (s *MemoryStorage) AddFavorite(userID, assetReference, listReference string, sortOrder *int) (*models.Favorite, error) {
	// Get or create default list if listReference is empty
	if listReference == "" {
		list, err := s.GetListByName(userID, "default")
		if err != nil {
			// Create default list if it doesn't exist
			list, err = s.CreateList(userID, "default")
			if err != nil {
				return nil, err
			}
		}
		listReference = list.Reference
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Verify asset exists
	asset, exists := s.assets[assetReference]
	if !exists {
		return nil, ErrAssetNotFound
	}

	now := time.Now()

	// Initialize user's favorites map if needed
	if s.favorites[userID] == nil {
		s.favorites[userID] = make(map[string]map[string]*models.Favorite)
	}
	if s.favorites[userID][listReference] == nil {
		s.favorites[userID][listReference] = make(map[string]*models.Favorite)
	}

	// Get list for response
	var listName string
	if s.lists[userID] != nil && s.lists[userID][listReference] != nil {
		listName = s.lists[userID][listReference].Name
	} else {
		// Extract list name from reference (list_user1_default -> default)
		// For simplicity, we'll use "default" if we can't find it
		listName = "default"
	}

	// Check if favorite already exists
	if existing, exists := s.favorites[userID][listReference][assetReference]; exists {
		if sortOrder != nil {
			existing.SortOrder = *sortOrder
		}
		existing.UpdatedAt = now
		existing.Asset = asset
		return existing, nil
	}

	// Calculate sort_order
	var finalSortOrder int
	if sortOrder != nil {
		finalSortOrder = *sortOrder
	} else {
		// Calculate next sort_order for this list
		maxSortOrder := -1
		for _, fav := range s.favorites[userID][listReference] {
			if fav.SortOrder > maxSortOrder {
				maxSortOrder = fav.SortOrder
			}
		}
		finalSortOrder = maxSortOrder + 1
	}

	// Create new favorite
	// Format: fav_{assetReference} (e.g., "fav_user1_favorite1")
	favoriteReference := fmt.Sprintf("fav_%s", assetReference)
	favorite := &models.Favorite{
		ID:        len(s.favorites[userID][listReference]),
		Reference: favoriteReference,
		UserID:    userID,
		AssetID:   assetReference,
		ListID:    0, // Not used in memory storage
		ListRef:   listReference,
		ListName:  listName,
		SortOrder: finalSortOrder,
		Asset:     asset,
		AddedAt:   now,
		UpdatedAt: now,
	}

	s.favorites[userID][listReference][assetReference] = favorite
	return favorite, nil
}

// RemoveFavorite removes a favorite for a user from a specific list
func (s *MemoryStorage) RemoveFavorite(userID, assetID, listReference string) error {
	// Get or find default list if listReference is empty
	if listReference == "" {
		list, err := s.GetListByName(userID, "default")
		if err != nil {
			return ErrFavoriteNotFound
		}
		listReference = list.Reference
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.favorites[userID] == nil || s.favorites[userID][listReference] == nil {
		return ErrFavoriteNotFound
	}

	if _, exists := s.favorites[userID][listReference][assetID]; !exists {
		return ErrFavoriteNotFound
	}

	delete(s.favorites[userID][listReference], assetID)
	return nil
}

// RemoveFavoriteByReference removes a favorite by its reference
func (s *MemoryStorage) RemoveFavoriteByReference(favoriteReference string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Parse reference format: fav_{assetReference} (e.g., "fav_user1_favorite1")
	// For simplicity, iterate through all favorites
	for userID, lists := range s.favorites {
		for listReference, assets := range lists {
			for assetID, fav := range assets {
				if fav.Reference == favoriteReference {
					delete(s.favorites[userID][listReference], assetID)
					return nil
				}
			}
		}
	}
	return ErrFavoriteNotFound
}

// GetFavorites returns all favorites for a user in a specific list
func (s *MemoryStorage) GetFavorites(userID, listReference string) ([]*models.Favorite, error) {
	// Get or find default list if listReference is empty
	if listReference == "" {
		list, err := s.GetListByName(userID, "default")
		if err != nil {
			// Create default list if it doesn't exist
			list, err = s.CreateList(userID, "default")
			if err != nil {
				return nil, err
			}
		}
		listReference = list.Reference
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.favorites[userID] == nil || s.favorites[userID][listReference] == nil {
		return []*models.Favorite{}, nil
	}

	favorites := make([]*models.Favorite, 0, len(s.favorites[userID][listReference]))
	for _, fav := range s.favorites[userID][listReference] {
		// Ensure asset is up to date
		if asset, exists := s.assets[fav.AssetID]; exists {
			fav.Asset = asset
		}
		favorites = append(favorites, fav)
	}

	// Sort by sort_order (ascending), then by added_at (descending)
	sort.Slice(favorites, func(i, j int) bool {
		if favorites[i].SortOrder != favorites[j].SortOrder {
			return favorites[i].SortOrder < favorites[j].SortOrder
		}
		return favorites[i].AddedAt.After(favorites[j].AddedAt)
	})

	return favorites, nil
}

// GetFavoritesPaginated returns paginated favorites for a user in a specific list
func (s *MemoryStorage) GetFavoritesPaginated(userID, listReference string, page, pageSize int, filters *models.FilterParams) ([]*models.Favorite, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	allFavorites, err := s.GetFavorites(userID, listReference)
	if err != nil {
		return nil, 0, err
	}

	// Apply filters
	filteredFavorites := s.applyFilters(allFavorites, filters)

	totalCount := len(filteredFavorites)
	offset := (page - 1) * pageSize

	if offset >= totalCount {
		return []*models.Favorite{}, totalCount, nil
	}

	end := offset + pageSize
	if end > totalCount {
		end = totalCount
	}

	return filteredFavorites[offset:end], totalCount, nil
}

// applyFilters applies filter parameters to a list of favorites
func (s *MemoryStorage) applyFilters(favorites []*models.Favorite, filters *models.FilterParams) []*models.Favorite {
	if filters == nil {
		return favorites
	}

	var filtered []*models.Favorite
	for _, fav := range favorites {
		// Filter by asset type
		if filters.AssetType != nil && fav.Asset.GetType() != *filters.AssetType {
			continue
		}

		// Filter by date_from (added_at >= date_from)
		if filters.DateFrom != nil && fav.AddedAt.Before(*filters.DateFrom) {
			continue
		}

		// Filter by date_to (added_at <= date_to)
		if filters.DateTo != nil && fav.AddedAt.After(*filters.DateTo) {
			continue
		}

		filtered = append(filtered, fav)
	}

	return filtered
}

// GetAllFavorites returns all favorites for a user across all lists
func (s *MemoryStorage) GetAllFavorites(userID string) ([]*models.Favorite, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.favorites[userID] == nil {
		return []*models.Favorite{}, nil
	}

	var allFavorites []*models.Favorite
	for listReference := range s.favorites[userID] {
		if s.favorites[userID][listReference] != nil {
			for _, fav := range s.favorites[userID][listReference] {
				// Ensure asset is up to date
				if asset, exists := s.assets[fav.AssetID]; exists {
					fav.Asset = asset
				}
				allFavorites = append(allFavorites, fav)
			}
		}
	}

	// Sort by list_id (ascending), then sort_order (ascending), then added_at (descending)
	sort.Slice(allFavorites, func(i, j int) bool {
		// First sort by list reference (to group by list)
		if allFavorites[i].ListRef != allFavorites[j].ListRef {
			return allFavorites[i].ListRef < allFavorites[j].ListRef
		}
		// Then by sort_order
		if allFavorites[i].SortOrder != allFavorites[j].SortOrder {
			return allFavorites[i].SortOrder < allFavorites[j].SortOrder
		}
		// Finally by added_at
		return allFavorites[i].AddedAt.After(allFavorites[j].AddedAt)
	})

	return allFavorites, nil
}

// GetAllFavoritesPaginated returns paginated favorites for a user across all lists
func (s *MemoryStorage) GetAllFavoritesPaginated(userID string, page, pageSize int, filters *models.FilterParams) ([]*models.Favorite, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	allFavorites, err := s.GetAllFavorites(userID)
	if err != nil {
		return nil, 0, err
	}

	// Apply filters
	filteredFavorites := s.applyFilters(allFavorites, filters)

	totalCount := len(filteredFavorites)
	offset := (page - 1) * pageSize

	if offset >= totalCount {
		return []*models.Favorite{}, totalCount, nil
	}

	end := offset + pageSize
	if end > totalCount {
		end = totalCount
	}

	return filteredFavorites[offset:end], totalCount, nil
}

// GetAllListsPaginated returns paginated lists for a user
func (s *MemoryStorage) GetAllListsPaginated(userID string, page, pageSize int) ([]*models.List, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	allLists, err := s.GetAllLists(userID)
	if err != nil {
		return nil, 0, err
	}

	totalCount := len(allLists)
	offset := (page - 1) * pageSize

	if offset >= totalCount {
		return []*models.List{}, totalCount, nil
	}

	end := offset + pageSize
	if end > totalCount {
		end = totalCount
	}

	return allLists[offset:end], totalCount, nil
}

// UpdateAssetDescription updates the description of an asset
func (s *MemoryStorage) UpdateAssetDescription(assetID, description string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	asset, exists := s.assets[assetID]
	if !exists {
		return ErrAssetNotFound
	}

	asset.SetDescription(description)
	asset.SetUpdatedAt(time.Now())
	s.assets[assetID] = asset

	// Update asset in all user favorites
	for userID, lists := range s.favorites {
		for listReference := range lists {
			if fav, exists := s.favorites[userID][listReference][assetID]; exists {
				fav.Asset = asset
				fav.UpdatedAt = time.Now()
			}
		}
	}

	return nil
}

// GetAsset retrieves an asset by ID
func (s *MemoryStorage) GetAsset(assetID string) (models.Asset, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	asset, exists := s.assets[assetID]
	if !exists {
		return nil, ErrAssetNotFound
	}

	return asset, nil
}

// GetAllAssetsPaginated returns paginated assets
func (s *MemoryStorage) GetAllAssetsPaginated(page, pageSize int, filters *models.FilterParams) ([]models.Asset, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Convert map to slice
	var allAssets []models.Asset
	for _, asset := range s.assets {
		allAssets = append(allAssets, asset)
	}

	// Apply filters
	filteredAssets := s.applyAssetFilters(allAssets, filters)

	// Sort by created_at descending (newest first)
	sort.Slice(filteredAssets, func(i, j int) bool {
		return filteredAssets[i].GetCreatedAt().After(filteredAssets[j].GetCreatedAt())
	})

	totalCount := len(filteredAssets)
	offset := (page - 1) * pageSize

	if offset >= totalCount {
		return []models.Asset{}, totalCount, nil
	}

	end := offset + pageSize
	if end > totalCount {
		end = totalCount
	}

	return filteredAssets[offset:end], totalCount, nil
}

// applyAssetFilters applies filters to assets
func (s *MemoryStorage) applyAssetFilters(assets []models.Asset, filters *models.FilterParams) []models.Asset {
	if filters == nil {
		return assets
	}

	var filtered []models.Asset
	for _, asset := range assets {
		// Filter by asset type
		if filters.AssetType != nil && asset.GetType() != *filters.AssetType {
			continue
		}

		// Filter by date_from (created_at)
		if filters.DateFrom != nil && asset.GetCreatedAt().Before(*filters.DateFrom) {
			continue
		}

		// Filter by date_to (created_at)
		if filters.DateTo != nil && asset.GetCreatedAt().After(*filters.DateTo) {
			continue
		}

		filtered = append(filtered, asset)
	}

	return filtered
}

// Storage errors
var (
	ErrFavoriteNotFound = &StorageError{Message: "favorite not found", Code: "NOT_FOUND"}
	ErrAssetNotFound    = &StorageError{Message: "asset not found", Code: "NOT_FOUND"}
)

// StorageError represents a storage operation error
type StorageError struct {
	Message string
	Code    string
}

func (e *StorageError) Error() string {
	return e.Message
}

// GenerateNextAssetID generates the next sequential asset ID for a user
// Format: {userReference}_favorite{number} (e.g., "user1_favorite1", "user1_favorite2")
func (s *MemoryStorage) GenerateNextAssetID(userReference string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	prefix := fmt.Sprintf("%s_favorite", userReference)
	maxNumber := 0

	// Find the highest numbered asset for this user
	for assetID := range s.assets {
		if strings.HasPrefix(assetID, prefix) {
			// Extract number from asset ID (e.g., "user1_favorite5" -> 5)
			numStr := assetID[len(prefix):]
			var num int
			if _, err := fmt.Sscanf(numStr, "%d", &num); err == nil {
				if num > maxNumber {
					maxNumber = num
				}
			}
		}
	}

	nextNumber := maxNumber + 1
	return fmt.Sprintf("%s_favorite%d", userReference, nextNumber), nil
}

// UpdateFavoriteSortOrder updates the sort order of a favorite
func (s *MemoryStorage) UpdateFavoriteSortOrder(userReference, favoriteReference string, sortOrder int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find the favorite by reference
	if s.favorites[userReference] == nil {
		return ErrFavoriteNotFound
	}

	for listReference := range s.favorites[userReference] {
		if s.favorites[userReference][listReference] != nil {
			for _, fav := range s.favorites[userReference][listReference] {
				if fav.Reference == favoriteReference {
					fav.SortOrder = sortOrder
					fav.UpdatedAt = time.Now()
					return nil
				}
			}
		}
	}

	return ErrFavoriteNotFound
}

// Close implements the Storage interface (no-op for in-memory storage)
func (s *MemoryStorage) Close() error {
	return nil
}
