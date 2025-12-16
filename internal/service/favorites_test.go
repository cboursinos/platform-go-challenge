package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/gwi/platform-go-challenge/internal/models"
	"github.com/gwi/platform-go-challenge/internal/storage"
)

func TestFavoritesService_AddFavorite(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	// Create default list first
	list, err := service.CreateList("user1", "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}
	
	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "chart1",
			Type:        models.AssetTypeChart,
			Description: "Test Chart",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Title: "Sales Chart",
		XAxis: "Month",
		YAxis: "Revenue",
	}

	_, err = service.CreateAsset("user1", chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}

	favorite, err := service.AddFavorite("user1", "chart1", list.Reference, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if favorite == nil {
		t.Fatal("Expected favorite, got nil")
	}

	if favorite.UserID != "user1" {
		t.Errorf("Expected user ID user1, got %s", favorite.UserID)
	}
}

func TestFavoritesService_AddFavorite_Validation(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	// Create default list first
	list, err := service.CreateList("user1", "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Test empty user ID
	_, err = service.AddFavorite("", "chart1", list.Reference, nil)
	if err == nil {
		t.Error("Expected error for empty user ID")
	}

	// Test empty asset reference
	_, err = service.AddFavorite("user1", "", list.Reference, nil)
	if err == nil {
		t.Error("Expected error for empty asset reference")
	}

	// Test non-existent asset
	_, err = service.AddFavorite("user1", "nonexistent", list.Reference, nil)
	if err == nil {
		t.Error("Expected error for non-existent asset")
	}
}

func TestFavoritesService_RemoveFavorite(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	// Create default list first
	list, err := service.CreateList("user1", "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}
	
	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "chart1",
			Type:        models.AssetTypeChart,
			Description: "Test Chart",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Title: "Sales Chart",
	}

	_, err = service.CreateAsset("user1", chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	service.AddFavorite("user1", "chart1", list.Reference, nil)

	err = service.RemoveFavorite("user1", "chart1", list.Reference)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestFavoritesService_GetFavorites(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	// Create default list first
	list, err := service.CreateList("user1", "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}
	
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

	_, err = service.CreateAsset("user1", chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	_, err = service.CreateAsset("user1", insight)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	service.AddFavorite("user1", "chart1", list.Reference, nil)
	service.AddFavorite("user1", "insight1", list.Reference, nil)

	favorites, err := service.GetFavorites("user1", list.Reference)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(favorites) != 2 {
		t.Errorf("Expected 2 favorites, got %d", len(favorites))
	}
}

func TestFavoritesService_GetFavoritesPaginated(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	// Create default list first
	list, err := service.CreateList("user1", "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Add multiple favorites
	for i := 1; i <= 25; i++ {
		chart := &models.Chart{
			BaseAsset: models.BaseAsset{
				ID:          fmt.Sprintf("chart%d", i),
				Type:        models.AssetTypeChart,
				Description: fmt.Sprintf("Chart %d", i),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			Title: fmt.Sprintf("Chart %d", i),
		}
		_, err := service.CreateAsset("user1", chart)
		if err != nil {
			t.Fatalf("Failed to create asset: %v", err)
		}
		service.AddFavorite("user1", fmt.Sprintf("chart%d", i), list.Reference, nil)
	}

	// Test pagination
	favorites, totalCount, err := service.GetFavoritesPaginated("user1", list.Reference, 1, 10, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(favorites) != 10 {
		t.Errorf("Expected 10 favorites on page 1, got %d", len(favorites))
	}
	if totalCount != 25 {
		t.Errorf("Expected total count 25, got %d", totalCount)
	}

	// Test second page
	favorites2, totalCount2, err := service.GetFavoritesPaginated("user1", list.Reference, 2, 10, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(favorites2) != 10 {
		t.Errorf("Expected 10 favorites on page 2, got %d", len(favorites2))
	}
	if totalCount2 != 25 {
		t.Errorf("Expected total count 25, got %d", totalCount2)
	}
}

func TestFavoritesService_UpdateDescription(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	// Create default list first
	list, err := service.CreateList("user1", "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}
	
	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "chart1",
			Type:        models.AssetTypeChart,
			Description: "Old Description",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Title: "Sales Chart",
	}

	_, err = service.CreateAsset("user1", chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	service.AddFavorite("user1", "chart1", list.Reference, nil)

	err = service.UpdateDescription("chart1", "New Description")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	favorites, _ := service.GetFavorites("user1", list.Reference)
	if favorites[0].Asset.GetDescription() != "New Description" {
		t.Errorf("Expected 'New Description', got %s", favorites[0].Asset.GetDescription())
	}
}

func TestFavoritesService_ValidateAsset(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	// Test valid chart
	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:   "chart1",
			Type: models.AssetTypeChart,
		},
		Title: "Test Chart",
	}
	err := service.ValidateAsset(chart)
	if err != nil {
		t.Errorf("Expected no error for valid chart, got %v", err)
	}

	// Test chart without title
	chartNoTitle := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:   "chart2",
			Type: models.AssetTypeChart,
		},
	}
	err = service.ValidateAsset(chartNoTitle)
	if err == nil {
		t.Error("Expected error for chart without title")
	}

	// Test valid insight
	insight := &models.Insight{
		BaseAsset: models.BaseAsset{
			ID:   "insight1",
			Type: models.AssetTypeInsight,
		},
		Text: "Test insight",
	}
	err = service.ValidateAsset(insight)
	if err != nil {
		t.Errorf("Expected no error for valid insight, got %v", err)
	}

	// Test insight without text
	insightNoText := &models.Insight{
		BaseAsset: models.BaseAsset{
			ID:   "insight2",
			Type: models.AssetTypeInsight,
		},
	}
	err = service.ValidateAsset(insightNoText)
	if err == nil {
		t.Error("Expected error for insight without text")
	}

	// Test valid audience
	audience := &models.Audience{
		BaseAsset: models.BaseAsset{
			ID:   "audience1",
			Type: models.AssetTypeAudience,
		},
	}
	err = service.ValidateAsset(audience)
	if err != nil {
		t.Errorf("Expected no error for valid audience, got %v", err)
	}
}

func TestFavoritesService_CreateAssetFromJSON(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	// Test chart creation
	chartData := map[string]interface{}{
		"id":          "chart1",
		"type":        "chart",
		"description": "Test Chart",
		"title":       "Sales Chart",
		"x_axis":      "Month",
		"y_axis":      "Revenue",
		"data": []map[string]interface{}{
			{"label": "Jan", "value": 100},
			{"label": "Feb", "value": 200},
		},
	}

	asset, err := service.CreateAssetFromJSON(chartData)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	chart, ok := asset.(*models.Chart)
	if !ok {
		t.Fatal("Expected Chart type")
	}

	if chart.Title != "Sales Chart" {
		t.Errorf("Expected title 'Sales Chart', got %s", chart.Title)
	}

	// Test insight creation
	insightData := map[string]interface{}{
		"id":          "insight1",
		"type":        "insight",
		"description": "Test Insight",
		"text":        "40% of millennials spend more than 3 hours on social media daily",
	}

	asset, err = service.CreateAssetFromJSON(insightData)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	insight, ok := asset.(*models.Insight)
	if !ok {
		t.Fatal("Expected Insight type")
	}

	if insight.Text != "40% of millennials spend more than 3 hours on social media daily" {
		t.Errorf("Unexpected insight text: %s", insight.Text)
	}

	// Test audience creation
	gender := "male"
	hours := 3.0
	purchases := 5
	audienceData := map[string]interface{}{
		"id":                      "audience1",
		"type":                    "audience",
		"description":             "Test Audience",
		"gender":                  gender,
		"birth_country":           "USA",
		"age_group":               map[string]interface{}{"min": 24.0, "max": 35.0},
		"social_media_hours_min":  hours,
		"purchases_last_month":    float64(purchases),
	}

	asset, err = service.CreateAssetFromJSON(audienceData)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	audience, ok := asset.(*models.Audience)
	if !ok {
		t.Fatal("Expected Audience type")
	}

	if audience.Gender == nil || *audience.Gender != models.Gender(gender) {
		t.Errorf("Unexpected gender: %v", audience.Gender)
	}

	if audience.BirthCountry != "USA" {
		t.Errorf("Expected birth country 'USA', got %s", audience.BirthCountry)
	}
}

func TestFavoritesService_GetAllFavorites(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	// Create two lists
	list1, err := service.CreateList("user1", "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}
	list2, err := service.CreateList("user1", "work")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Add favorites to different lists
	chart1 := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "chart1",
			Type:        models.AssetTypeChart,
			Description: "Chart 1",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Title: "Chart 1",
	}
	insight1 := &models.Insight{
		BaseAsset: models.BaseAsset{
			ID:          "insight1",
			Type:        models.AssetTypeInsight,
			Description: "Insight 1",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Text: "Test insight",
	}

	_, err = service.CreateAsset("user1", chart1)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	_, err = service.CreateAsset("user1", insight1)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	service.AddFavorite("user1", "chart1", list1.Reference, nil)
	service.AddFavorite("user1", "insight1", list2.Reference, nil)

	// Get all favorites across all lists
	allFavorites, err := service.GetAllFavorites("user1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(allFavorites) != 2 {
		t.Errorf("Expected 2 favorites across all lists, got %d", len(allFavorites))
	}
}

func TestFavoritesService_GetAllFavoritesPaginated(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	// Create list
	list, err := service.CreateList("user1", "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Add multiple favorites
	for i := 1; i <= 25; i++ {
		chart := &models.Chart{
			BaseAsset: models.BaseAsset{
				ID:          fmt.Sprintf("chart%d", i),
				Type:        models.AssetTypeChart,
				Description: fmt.Sprintf("Chart %d", i),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			Title: fmt.Sprintf("Chart %d", i),
		}
		_, err := service.CreateAsset("user1", chart)
		if err != nil {
			t.Fatalf("Failed to create asset: %v", err)
		}
		service.AddFavorite("user1", fmt.Sprintf("chart%d", i), list.Reference, nil)
	}

	// Test pagination
	favorites, totalCount, err := service.GetAllFavoritesPaginated("user1", 1, 10, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(favorites) != 10 {
		t.Errorf("Expected 10 favorites on page 1, got %d", len(favorites))
	}
	if totalCount != 25 {
		t.Errorf("Expected total count 25, got %d", totalCount)
	}

	// Test second page
	favorites2, totalCount2, err := service.GetAllFavoritesPaginated("user1", 2, 10, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(favorites2) != 10 {
		t.Errorf("Expected 10 favorites on page 2, got %d", len(favorites2))
	}
	if totalCount2 != 25 {
		t.Errorf("Expected total count 25, got %d", totalCount2)
	}
}

func TestFavoritesService_GetAllListsPaginated(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	// Create multiple lists
	for i := 1; i <= 15; i++ {
		_, err := service.CreateList("user1", fmt.Sprintf("list%d", i))
		if err != nil {
			t.Fatalf("Failed to create list: %v", err)
		}
	}

	// Test pagination
	lists, totalCount, err := service.GetAllListsPaginated("user1", 1, 5)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(lists) != 5 {
		t.Errorf("Expected 5 lists on page 1, got %d", len(lists))
	}
	if totalCount < 15 {
		t.Errorf("Expected at least 15 total lists, got %d", totalCount)
	}
}

func TestFavoritesService_GetList(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	// Create a list
	list, err := service.CreateList("user1", "work")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Get the list
	retrievedList, err := service.GetList("user1", list.Reference)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if retrievedList.Reference != list.Reference {
		t.Errorf("Expected list reference %s, got %s", list.Reference, retrievedList.Reference)
	}
	if retrievedList.Name != "work" {
		t.Errorf("Expected list name 'work', got %s", retrievedList.Name)
	}
}

func TestFavoritesService_GetList_NotFound(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	_, err := service.GetList("user1", "nonexistent_list")
	if err == nil {
		t.Error("Expected error for nonexistent list")
	}
}

func TestFavoritesService_GetList_EmptyUserID(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	_, err := service.GetList("", "list1")
	if err == nil {
		t.Error("Expected error for empty user ID")
	}
}

func TestFavoritesService_GetListByName(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	// Create a list
	_, err := service.CreateList("user1", "work")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Get the list by name
	list, err := service.GetListByName("user1", "work")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if list.Name != "work" {
		t.Errorf("Expected list name 'work', got %s", list.Name)
	}
}

func TestFavoritesService_GetListByName_NotFound(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	_, err := service.GetListByName("user1", "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent list name")
	}
}

func TestFavoritesService_GetAllLists(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	// Create multiple lists
	_, err := service.CreateList("user1", "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}
	_, err = service.CreateList("user1", "work")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Get all lists
	lists, err := service.GetAllLists("user1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(lists) < 2 {
		t.Errorf("Expected at least 2 lists, got %d", len(lists))
	}
}

func TestFavoritesService_GetAllLists_Empty(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	lists, err := service.GetAllLists("user1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// GetAllLists automatically creates a default list if none exist
	if len(lists) != 1 {
		t.Errorf("Expected 1 default list, got %d", len(lists))
	}
	if lists[0].Name != "default" {
		t.Errorf("Expected default list name 'default', got %s", lists[0].Name)
	}
}

func TestFavoritesService_DeleteList(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	// Create a list
	list, err := service.CreateList("user1", "work")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Delete the list
	err = service.DeleteList("user1", list.Reference)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify list is deleted
	_, err = service.GetList("user1", list.Reference)
	if err == nil {
		t.Error("Expected list to be deleted")
	}
}

func TestFavoritesService_DeleteList_NotFound(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	err := service.DeleteList("user1", "nonexistent_list")
	if err == nil {
		t.Error("Expected error for nonexistent list")
	}
}

func TestFavoritesService_RemoveFavoriteByReference(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	// Create list
	list, err := service.CreateList("user1", "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Add favorite
	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "chart1",
			Type:        models.AssetTypeChart,
			Description: "Test Chart",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Title: "Sales Chart",
	}
	_, err = service.CreateAsset("user1", chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	favorite, err := service.AddFavorite("user1", "chart1", list.Reference, nil)
	if err != nil {
		t.Fatalf("Failed to add favorite: %v", err)
	}

	// Remove by reference
	err = service.RemoveFavoriteByReference(favorite.Reference)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify favorite is removed
	favorites, _ := service.GetFavorites("user1", list.Reference)
	if len(favorites) != 0 {
		t.Errorf("Expected 0 favorites after removal, got %d", len(favorites))
	}
}

func TestFavoritesService_RemoveFavoriteByReference_NotFound(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	err := service.RemoveFavoriteByReference("nonexistent_reference")
	if err == nil {
		t.Error("Expected error for nonexistent favorite reference")
	}
}

func TestFavoritesService_GetFavorites_EmptyList(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	// Create list but don't add favorites
	list, err := service.CreateList("user1", "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	favorites, err := service.GetFavorites("user1", list.Reference)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(favorites) != 0 {
		t.Errorf("Expected 0 favorites, got %d", len(favorites))
	}
}

func TestFavoritesService_GetFavorites_CreatesDefaultList(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	// Get favorites without creating a list first (should create default)
	favorites, err := service.GetFavorites("user1", "")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(favorites) != 0 {
		t.Errorf("Expected 0 favorites, got %d", len(favorites))
	}

	// Verify default list was created
	_, err = service.GetListByName("user1", "default")
	if err != nil {
		t.Error("Expected default list to be created")
	}
}

func TestFavoritesService_GetFavoritesPaginated_EmptyList(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	// Create list but don't add favorites
	list, err := service.CreateList("user1", "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	favorites, totalCount, err := service.GetFavoritesPaginated("user1", list.Reference, 1, 10, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(favorites) != 0 {
		t.Errorf("Expected 0 favorites, got %d", len(favorites))
	}
	if totalCount != 0 {
		t.Errorf("Expected total count 0, got %d", totalCount)
	}
}

func TestFavoritesService_GetFavoritesPaginated_BoundaryConditions(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	// Create list
	list, err := service.CreateList("user1", "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Add exactly 5 favorites
	for i := 1; i <= 5; i++ {
		chart := &models.Chart{
			BaseAsset: models.BaseAsset{
				ID:          fmt.Sprintf("chart%d", i),
				Type:        models.AssetTypeChart,
				Description: fmt.Sprintf("Chart %d", i),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			Title: fmt.Sprintf("Chart %d", i),
		}
		_, err := service.CreateAsset("user1", chart)
		if err != nil {
			t.Fatalf("Failed to create asset: %v", err)
		}
		service.AddFavorite("user1", fmt.Sprintf("chart%d", i), list.Reference, nil)
	}

	// Test page beyond total
	favorites, totalCount, err := service.GetFavoritesPaginated("user1", list.Reference, 999, 10, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(favorites) != 0 {
		t.Errorf("Expected 0 favorites for page beyond total, got %d", len(favorites))
	}
	if totalCount != 5 {
		t.Errorf("Expected total count 5, got %d", totalCount)
	}
}

func TestFavoritesService_UpdateDescription_NotFound(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	err := service.UpdateDescription("nonexistent_asset", "New Description")
	if err == nil {
		t.Error("Expected error for nonexistent asset")
	}
}

func TestFavoritesService_ValidateAsset_NilAsset(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	err := service.ValidateAsset(nil)
	if err == nil {
		t.Error("Expected error for nil asset")
	}
}

func TestFavoritesService_ValidateAsset_EmptyID(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:   "",
			Type: models.AssetTypeChart,
		},
		Title: "Test",
	}

	err := service.ValidateAsset(chart)
	if err == nil {
		t.Error("Expected error for empty asset ID")
	}
}

func TestFavoritesService_ValidateAsset_EmptyType(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:   "chart1",
			Type: "",
		},
		Title: "Test",
	}

	err := service.ValidateAsset(chart)
	if err == nil {
		t.Error("Expected error for empty asset type")
	}
}

// MockAsset is a test helper that implements the Asset interface
type MockAsset struct {
	models.BaseAsset
}

func (m *MockAsset) GetID() string                { return m.BaseAsset.ID }
func (m *MockAsset) GetType() models.AssetType   { return m.BaseAsset.Type }
func (m *MockAsset) GetDescription() string       { return m.BaseAsset.Description }
func (m *MockAsset) SetDescription(d string)      { m.BaseAsset.Description = d }
func (m *MockAsset) GetCreatedAt() time.Time      { return m.BaseAsset.CreatedAt }
func (m *MockAsset) GetUpdatedAt() time.Time      { return m.BaseAsset.UpdatedAt }
func (m *MockAsset) SetUpdatedAt(t time.Time)     { m.BaseAsset.UpdatedAt = t }

func TestFavoritesService_ValidateAsset_UnknownType(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	// Create a mock asset with unknown type
	mockAsset := &MockAsset{
		BaseAsset: models.BaseAsset{
			ID:        "asset1",
			Type:      "unknown_type",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	err := service.ValidateAsset(mockAsset)
	if err == nil {
		t.Error("Expected error for unknown asset type")
	}
}

func TestFavoritesService_CreateAssetFromJSON_InvalidType(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	data := map[string]interface{}{
		"id":   "asset1",
		"type": "invalid_type",
	}

	_, err := service.CreateAssetFromJSON(data)
	if err == nil {
		t.Error("Expected error for invalid asset type")
	}
}

func TestFavoritesService_CreateAssetFromJSON_MissingType(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	data := map[string]interface{}{
		"id": "asset1",
		// Missing type
	}

	_, err := service.CreateAssetFromJSON(data)
	if err == nil {
		t.Error("Expected error for missing asset type")
	}
}

func TestFavoritesService_GetAllFavoritesPaginated_EmptyUser(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	favorites, totalCount, err := service.GetAllFavoritesPaginated("user1", 1, 10, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(favorites) != 0 {
		t.Errorf("Expected 0 favorites, got %d", len(favorites))
	}
	if totalCount != 0 {
		t.Errorf("Expected total count 0, got %d", totalCount)
	}
}

func TestFavoritesService_GetAllListsPaginated_Empty(t *testing.T) {
	storage := storage.NewMemoryStorage()
	service := NewFavoritesService(storage)

	lists, totalCount, err := service.GetAllListsPaginated("user1", 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// GetAllListsPaginated calls GetAllLists which automatically creates a default list if none exist
	if len(lists) != 1 {
		t.Errorf("Expected 1 default list, got %d", len(lists))
	}
	if totalCount != 1 {
		t.Errorf("Expected total count 1, got %d", totalCount)
	}
	if lists[0].Name != "default" {
		t.Errorf("Expected default list name 'default', got %s", lists[0].Name)
	}
}

