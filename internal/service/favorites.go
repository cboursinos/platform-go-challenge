package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gwi/platform-go-challenge/internal/models"
	"github.com/gwi/platform-go-challenge/internal/storage"
)

// FavoritesService handles business logic for favorites
type FavoritesService struct {
	storage storage.Storage
}

// NewFavoritesService creates a new favorites service
func NewFavoritesService(storage storage.Storage) *FavoritesService {
	return &FavoritesService{
		storage: storage,
	}
}

// CreateList creates a new list for a user
func (s *FavoritesService) CreateList(userID, listName string) (*models.List, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if listName == "" {
		return nil, fmt.Errorf("list name is required")
	}
	return s.storage.CreateList(userID, listName)
}

// GetList retrieves a list by reference
func (s *FavoritesService) GetList(userID, listReference string) (*models.List, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if listReference == "" {
		return nil, fmt.Errorf("list reference is required")
	}
	return s.storage.GetList(userID, listReference)
}

// GetListByName retrieves a list by name for a user
func (s *FavoritesService) GetListByName(userID, listName string) (*models.List, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if listName == "" {
		listName = "default"
	}
	return s.storage.GetListByName(userID, listName)
}

// GetAllLists returns all lists for a user
func (s *FavoritesService) GetAllLists(userID string) ([]*models.List, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	return s.storage.GetAllLists(userID)
}

// GetAllListsPaginated returns paginated lists for a user
func (s *FavoritesService) GetAllListsPaginated(userID string, page, pageSize int) ([]*models.List, int, error) {
	if userID == "" {
		return nil, 0, fmt.Errorf("user ID is required")
	}
	return s.storage.GetAllListsPaginated(userID, page, pageSize)
}

// GetAllFavorites returns all favorites for a user across all lists
func (s *FavoritesService) GetAllFavorites(userID string) ([]*models.Favorite, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	return s.storage.GetAllFavorites(userID)
}

// GetAllFavoritesPaginated returns paginated favorites for a user across all lists
func (s *FavoritesService) GetAllFavoritesPaginated(userID string, page, pageSize int, filters *models.FilterParams) ([]*models.Favorite, int, error) {
	if userID == "" {
		return nil, 0, fmt.Errorf("user ID is required")
	}
	return s.storage.GetAllFavoritesPaginated(userID, page, pageSize, filters)
}

// DeleteList deletes a list and all its favorites
func (s *FavoritesService) DeleteList(userID, listReference string) error {
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}
	if listReference == "" {
		return fmt.Errorf("list reference is required")
	}
	return s.storage.DeleteList(userID, listReference)
}

// CreateAsset creates a new asset
func (s *FavoritesService) CreateAsset(userID string, asset models.Asset) (models.Asset, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	if asset == nil {
		return nil, fmt.Errorf("asset is required")
	}

	// Validate asset
	if err := s.ValidateAsset(asset); err != nil {
		return nil, fmt.Errorf("invalid asset: %w", err)
	}

	return s.storage.CreateAsset(userID, asset)
}

// UpdateAsset updates an existing asset
func (s *FavoritesService) UpdateAsset(assetReference string, asset models.Asset) (models.Asset, error) {
	if assetReference == "" {
		return nil, fmt.Errorf("asset reference is required")
	}

	if asset == nil {
		return nil, fmt.Errorf("asset is required")
	}

	// Validate asset
	if err := s.ValidateAsset(asset); err != nil {
		return nil, fmt.Errorf("invalid asset: %w", err)
	}

	return s.storage.UpdateAsset(assetReference, asset)
}

// DeleteAsset deletes an asset
func (s *FavoritesService) DeleteAsset(assetReference string) error {
	if assetReference == "" {
		return fmt.Errorf("asset reference is required")
	}

	return s.storage.DeleteAsset(assetReference)
}

// AddFavorite adds an asset to a user's favorites in a specific list
// assetReference is the asset reference (e.g., "user1_favorite1")
// listReference is optional, defaults to "default"
// sortOrder is optional, defaults to next sequential value
func (s *FavoritesService) AddFavorite(userID, assetReference, listReference string, sortOrder *int) (*models.Favorite, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	if assetReference == "" {
		return nil, fmt.Errorf("asset reference is required")
	}

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

	return s.storage.AddFavorite(userID, assetReference, listReference, sortOrder)
}

// RemoveFavorite removes an asset from a user's favorites
func (s *FavoritesService) RemoveFavorite(userID, assetID, listReference string) error {
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	if assetID == "" {
		return fmt.Errorf("asset ID is required")
	}

	// Get or find default list if listReference is empty
	if listReference == "" {
		list, err := s.GetListByName(userID, "default")
		if err != nil {
			return fmt.Errorf("default list not found: %w", err)
		}
		listReference = list.Reference
	}

	return s.storage.RemoveFavorite(userID, assetID, listReference)
}

func (s *FavoritesService) RemoveFavoriteByReference(favoriteReference string) error {
	if favoriteReference == "" {
		return fmt.Errorf("favorite reference is required")
	}

	return s.storage.RemoveFavoriteByReference(favoriteReference)
}

// GetFavorites retrieves all favorites for a user
func (s *FavoritesService) GetFavorites(userID, listReference string) ([]*models.Favorite, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

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

	return s.storage.GetFavorites(userID, listReference)
}

// GetFavoritesPaginated returns paginated favorites for a user in a specific list
func (s *FavoritesService) GetFavoritesPaginated(userID, listReference string, page, pageSize int, filters *models.FilterParams) ([]*models.Favorite, int, error) {
	if userID == "" {
		return nil, 0, fmt.Errorf("user ID is required")
	}

	// Get or find default list if listReference is empty
	if listReference == "" {
		list, err := s.GetListByName(userID, "default")
		if err != nil {
			// Create default list if it doesn't exist
			list, err = s.CreateList(userID, "default")
			if err != nil {
				return nil, 0, err
			}
		}
		listReference = list.Reference
	}

	return s.storage.GetFavoritesPaginated(userID, listReference, page, pageSize, filters)
}

// UpdateDescription updates the description of an asset
func (s *FavoritesService) UpdateDescription(assetID, description string) error {
	if assetID == "" {
		return fmt.Errorf("asset ID is required")
	}

	if description == "" {
		return fmt.Errorf("description cannot be empty")
	}

	return s.storage.UpdateAssetDescription(assetID, description)
}

// UpdateFavoriteSortOrder updates the sort order of a favorite
func (s *FavoritesService) UpdateFavoriteSortOrder(userReference, favoriteReference string, sortOrder int) error {
	if userReference == "" {
		return fmt.Errorf("user reference is required")
	}

	if favoriteReference == "" {
		return fmt.Errorf("favorite reference is required")
	}

	if sortOrder < 0 {
		return fmt.Errorf("sort order must be non-negative")
	}

	return s.storage.UpdateFavoriteSortOrder(userReference, favoriteReference, sortOrder)
}

// GetAsset retrieves an asset by reference
func (s *FavoritesService) GetAsset(assetReference string) (models.Asset, error) {
	if assetReference == "" {
		return nil, fmt.Errorf("asset reference is required")
	}
	return s.storage.GetAsset(assetReference)
}

// GetAllAssetsPaginated retrieves paginated assets with optional filters
func (s *FavoritesService) GetAllAssetsPaginated(page, pageSize int, filters *models.FilterParams) ([]models.Asset, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return s.storage.GetAllAssetsPaginated(page, pageSize, filters)
}

// GenerateNextAssetID generates the next sequential asset ID for a user
// Format: {userReference}_favorite{number} (e.g., "user1_favorite1", "user1_favorite2")
func (s *FavoritesService) GenerateNextAssetID(userReference string) (string, error) {
	if userReference == "" {
		return "", fmt.Errorf("user reference is required")
	}
	return s.storage.GenerateNextAssetID(userReference)
}

// ValidateAsset validates an asset based on its type
func (s *FavoritesService) ValidateAsset(asset models.Asset) error {
	if asset == nil {
		return fmt.Errorf("asset cannot be nil")
	}

	if asset.GetID() == "" {
		return fmt.Errorf("asset ID is required")
	}

	if asset.GetType() == "" {
		return fmt.Errorf("asset type is required")
	}

	switch asset.GetType() {
	case models.AssetTypeChart:
		chart, ok := asset.(*models.Chart)
		if !ok {
			return fmt.Errorf("invalid chart asset")
		}
		if chart.Title == "" {
			return fmt.Errorf("chart title is required")
		}

	case models.AssetTypeInsight:
		insight, ok := asset.(*models.Insight)
		if !ok {
			return fmt.Errorf("invalid insight asset")
		}
		if insight.Text == "" {
			return fmt.Errorf("insight text is required")
		}

	case models.AssetTypeAudience:
		// Audience validation is minimal as per requirements
		_, ok := asset.(*models.Audience)
		if !ok {
			return fmt.Errorf("invalid audience asset")
		}

	default:
		return fmt.Errorf("unknown asset type: %s", asset.GetType())
	}

	return nil
}

// CreateAssetFromJSON creates an asset from JSON data
func (s *FavoritesService) CreateAssetFromJSON(data map[string]interface{}) (models.Asset, error) {
	assetType, ok := data["type"].(string)
	if !ok {
		return nil, fmt.Errorf("asset type is required")
	}

	now := time.Now()
	id, _ := data["id"].(string)
	description, _ := data["description"].(string)

	baseAsset := models.BaseAsset{
		ID:          id,
		Type:        models.AssetType(assetType),
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	switch models.AssetType(assetType) {
	case models.AssetTypeChart:
		chart := &models.Chart{
			BaseAsset: baseAsset,
		}
		if title, ok := data["title"].(string); ok {
			chart.Title = title
		}
		if xAxis, ok := data["x_axis"].(string); ok {
			chart.XAxis = xAxis
		}
		if yAxis, ok := data["y_axis"].(string); ok {
			chart.YAxis = yAxis
		}
		// Parse data points
		if dataPoints, ok := data["data"].([]interface{}); ok {
			chart.Data = make([]models.DataPoint, 0, len(dataPoints))
			for _, dp := range dataPoints {
				if dpMap, ok := dp.(map[string]interface{}); ok {
					dataPoint := models.DataPoint{}
					if label, ok := dpMap["label"].(string); ok {
						dataPoint.Label = label
					}
					if value, ok := dpMap["value"]; ok {
						valueBytes, _ := json.Marshal(value)
						dataPoint.Value = valueBytes
					}
					chart.Data = append(chart.Data, dataPoint)
				}
			}
		}
		return chart, nil

	case models.AssetTypeInsight:
		insight := &models.Insight{
			BaseAsset: baseAsset,
		}
		if text, ok := data["text"].(string); ok {
			insight.Text = text
		}
		return insight, nil

	case models.AssetTypeAudience:
		audience := &models.Audience{
			BaseAsset: baseAsset,
		}
		if gender, ok := data["gender"].(string); ok {
			g := models.Gender(gender)
			audience.Gender = &g
		}
		if country, ok := data["birth_country"].(string); ok {
			audience.BirthCountry = country
		}
		if ageGroupMap, ok := data["age_group"].(map[string]interface{}); ok {
			ageGroup := &models.AgeGroup{}
			if min, ok := ageGroupMap["min"].(float64); ok {
				ageGroup.Min = int(min)
			}
			if max, ok := ageGroupMap["max"].(float64); ok {
				ageGroup.Max = int(max)
			}
			audience.AgeGroup = ageGroup
		}
		if hours, ok := data["social_media_hours_min"].(float64); ok {
			audience.SocialMediaHoursMin = &hours
		}
		if purchases, ok := data["purchases_last_month"].(float64); ok {
			p := int(purchases)
			audience.PurchasesLastMonth = &p
		}
		return audience, nil

	default:
		return nil, fmt.Errorf("unknown asset type: %s", assetType)
	}
}

