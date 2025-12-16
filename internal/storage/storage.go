package storage

import (
	"github.com/gwi/platform-go-challenge/internal/models"
)

// Storage defines the interface for storage operations
type Storage interface {
	// List operations
	CreateList(userID, listName string) (*models.List, error)
	GetList(userID, listReference string) (*models.List, error)
	GetListByName(userID, listName string) (*models.List, error)
	GetAllLists(userID string) ([]*models.List, error)
	GetAllListsPaginated(userID string, page, pageSize int) ([]*models.List, int, error) // Returns lists, total count, error
	DeleteList(userID, listReference string) error
	
	// Favorite operations
	AddFavorite(userID, assetReference, listReference string, sortOrder *int) (*models.Favorite, error)
	RemoveFavorite(userID, assetID, listReference string) error
	RemoveFavoriteByReference(favoriteReference string) error
	GetFavorites(userID, listReference string) ([]*models.Favorite, error)
	GetFavoritesPaginated(userID, listReference string, page, pageSize int, filters *models.FilterParams) ([]*models.Favorite, int, error) // Returns favorites, total count, error
	GetAllFavorites(userID string) ([]*models.Favorite, error) // Get favorites across all lists
	GetAllFavoritesPaginated(userID string, page, pageSize int, filters *models.FilterParams) ([]*models.Favorite, int, error) // Returns favorites, total count, error
	
	// Asset operations
	CreateAsset(userID string, asset models.Asset) (models.Asset, error)
	UpdateAsset(assetReference string, asset models.Asset) (models.Asset, error)
	DeleteAsset(assetReference string) error
	UpdateAssetDescription(assetID, description string) error
	GetAsset(assetID string) (models.Asset, error)
	GetAllAssetsPaginated(page, pageSize int, filters *models.FilterParams) ([]models.Asset, int, error) // Returns assets, total count, error
	GenerateNextAssetID(userReference string) (string, error) // Generate next sequential asset ID for a user
	
	// Sort order operations
	UpdateFavoriteSortOrder(userReference, favoriteReference string, sortOrder int) error
	
	Close() error
}

