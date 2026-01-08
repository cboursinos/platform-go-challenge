package storage

import (
	"context"
	"testing"
	"time"

	"github.com/gwi/platform-go-challenge/internal/models"
)

// TestMemoryStorage_GetFavoritesPaginated_WithFilters tests filtering by asset type
func TestMemoryStorage_GetFavoritesPaginated_WithFilters(t *testing.T) {
	storage := NewMemoryStorage()
	userID := "user1"

	// Create default list first
	list, err := storage.CreateList(context.Background(), userID, "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Add different asset types
	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "chart1",
			Type:        models.AssetTypeChart,
			Description: "Chart 1",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Title: "Chart 1",
	}
	_, err = storage.CreateAsset(context.Background(), userID, chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	storage.AddFavorite(context.Background(), userID, chart.GetID(), list.Reference, nil)

	insight := &models.Insight{
		BaseAsset: models.BaseAsset{
			ID:          "insight1",
			Type:        models.AssetTypeInsight,
			Description: "Insight 1",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Text: "Test insight",
	}
	_, err = storage.CreateAsset(context.Background(), userID, insight)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	storage.AddFavorite(context.Background(), userID, insight.GetID(), list.Reference, nil)

	audience := &models.Audience{
		BaseAsset: models.BaseAsset{
			ID:          "audience1",
			Type:        models.AssetTypeAudience,
			Description: "Audience 1",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}
	_, err = storage.CreateAsset(context.Background(), userID, audience)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	storage.AddFavorite(context.Background(), userID, audience.GetID(), list.Reference, nil)

	// Filter by chart type
	chartType := models.AssetTypeChart
	filters := &models.FilterParams{AssetType: &chartType}
	favorites, totalCount, err := storage.GetFavoritesPaginated(context.Background(), userID, list.Reference, 1, 10, filters)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if totalCount != 1 {
		t.Errorf("Expected 1 chart favorite, got %d", totalCount)
	}
	if len(favorites) != 1 {
		t.Errorf("Expected 1 favorite, got %d", len(favorites))
	}
	if favorites[0].Asset.GetType() != models.AssetTypeChart {
		t.Errorf("Expected chart type, got %s", favorites[0].Asset.GetType())
	}

	// Filter by insight type
	insightType := models.AssetTypeInsight
	filters = &models.FilterParams{AssetType: &insightType}
	favorites, totalCount, err = storage.GetFavoritesPaginated(context.Background(), userID, list.Reference, 1, 10, filters)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if totalCount != 1 {
		t.Errorf("Expected 1 insight favorite, got %d", totalCount)
	}
	if favorites[0].Asset.GetType() != models.AssetTypeInsight {
		t.Errorf("Expected insight type, got %s", favorites[0].Asset.GetType())
	}
}

// TestMemoryStorage_GetAllFavoritesPaginated_WithFilters tests filtering across all lists
func TestMemoryStorage_GetAllFavoritesPaginated_WithFilters(t *testing.T) {
	storage := NewMemoryStorage()
	userID := "user1"

	// Create lists
	list1, _ := storage.CreateList(context.Background(), userID, "list1")
	list2, _ := storage.CreateList(context.Background(), userID, "list2")

	// Add different asset types to different lists
	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "chart1",
			Type:        models.AssetTypeChart,
			Description: "Chart 1",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Title: "Chart 1",
	}
	_, err := storage.CreateAsset(context.Background(), userID, chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	storage.AddFavorite(context.Background(), userID, chart.GetID(), list1.Reference, nil)

	insight := &models.Insight{
		BaseAsset: models.BaseAsset{
			ID:          "insight1",
			Type:        models.AssetTypeInsight,
			Description: "Insight 1",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Text: "Test insight",
	}
	_, err = storage.CreateAsset(context.Background(), userID, insight)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	storage.AddFavorite(context.Background(), userID, insight.GetID(), list2.Reference, nil)

	// Filter by chart type across all lists
	chartType := models.AssetTypeChart
	filters := &models.FilterParams{AssetType: &chartType}
	favorites, totalCount, err := storage.GetAllFavoritesPaginated(context.Background(), userID, 1, 10, filters)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if totalCount != 1 {
		t.Errorf("Expected 1 chart favorite, got %d", totalCount)
	}
	if favorites[0].Asset.GetType() != models.AssetTypeChart {
		t.Errorf("Expected chart type, got %s", favorites[0].Asset.GetType())
	}
}
