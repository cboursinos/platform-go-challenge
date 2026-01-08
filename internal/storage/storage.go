package storage

import (
	"context"

	"github.com/gwi/platform-go-challenge/internal/models"
)

// Storage defines the interface for storage operations
type Storage interface {
	// List operations
	CreateList(ctx context.Context, userID, listName string) (*models.List, error)
	GetList(ctx context.Context, userID, listReference string) (*models.List, error)
	GetListByName(ctx context.Context, userID, listName string) (*models.List, error)
	GetAllLists(ctx context.Context, userID string) ([]*models.List, error)
	GetAllListsPaginated(ctx context.Context, userID string, page, pageSize int) ([]*models.List, int, error) // Returns lists, total count, error
	DeleteList(ctx context.Context, userID, listReference string) error
	
	// Favorite operations
	AddFavorite(ctx context.Context, userID, assetReference, listReference string, sortOrder *int) (*models.Favorite, error)
	RemoveFavorite(ctx context.Context, userID, assetID, listReference string) error
	RemoveFavoriteByReference(ctx context.Context, favoriteReference string) error
	GetFavorites(ctx context.Context, userID, listReference string) ([]*models.Favorite, error)
	GetFavoritesPaginated(ctx context.Context, userID, listReference string, page, pageSize int, filters *models.FilterParams) ([]*models.Favorite, int, error) // Returns favorites, total count, error
	GetAllFavorites(ctx context.Context, userID string) ([]*models.Favorite, error) // Get favorites across all lists
	GetAllFavoritesPaginated(ctx context.Context, userID string, page, pageSize int, filters *models.FilterParams) ([]*models.Favorite, int, error) // Returns favorites, total count, error
	
	// Asset operations
	CreateAsset(ctx context.Context, userID string, asset models.Asset) (models.Asset, error)
	UpdateAsset(ctx context.Context, assetReference string, asset models.Asset) (models.Asset, error)
	DeleteAsset(ctx context.Context, assetReference string) error
	UpdateAssetDescription(ctx context.Context, assetID, description string) error
	GetAsset(ctx context.Context, assetID string) (models.Asset, error)
	GetAllAssetsPaginated(ctx context.Context, page, pageSize int, filters *models.FilterParams) ([]models.Asset, int, error) // Returns assets, total count, error
	GenerateNextAssetID(ctx context.Context, userReference string) (string, error) // Generate next sequential asset ID for a user
	
	// Sort order operations
	UpdateFavoriteSortOrder(ctx context.Context, userReference, favoriteReference string, sortOrder int) error
	
	Close() error
}

