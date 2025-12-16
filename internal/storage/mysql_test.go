package storage

import (
	"testing"
	"time"

	"github.com/gwi/platform-go-challenge/internal/models"
)

func TestMySQLStorage_AddFavorite(t *testing.T) {
	skipIfMySQLUnavailable(t)

	dsn := getTestMySQLDSN()
	storage, err := NewMySQLStorageWithDriver(dsn, "mysql", nil) // nil = no New Relic in tests
	if err != nil {
		t.Fatalf("Failed to create MySQL storage: %v", err)
	}
	defer storage.Close()

	userID := "test_user_1"
	
	// Create test user first
	if err := createTestUser(storage.db, userID, "test_user_1@example.com", "Test User 1"); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	
	// Create default list first
	list, err := storage.CreateList(userID, "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}
	
	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "chart_test_1",
			Type:        models.AssetTypeChart,
			Description: "Test Chart",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Title: "Sales Chart",
		XAxis: "Month",
		YAxis: "Revenue",
	}

	_, err = storage.CreateAsset(userID, chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	favorite, err := storage.AddFavorite(userID, chart.GetID(), list.Reference, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if favorite.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, favorite.UserID)
	}

	if favorite.AssetID != chart.GetID() {
		t.Errorf("Expected asset ID %s, got %s", chart.GetID(), favorite.AssetID)
	}

	// Verify asset was stored
	retrievedAsset, err := storage.GetAsset(chart.GetID())
	if err != nil {
		t.Fatalf("Failed to retrieve asset: %v", err)
	}

	if retrievedAsset.GetID() != chart.GetID() {
		t.Errorf("Expected asset ID %s, got %s", chart.GetID(), retrievedAsset.GetID())
	}
}

func TestMySQLStorage_GetFavorites(t *testing.T) {
	skipIfMySQLUnavailable(t)

	dsn := getTestMySQLDSN()
	storage, err := NewMySQLStorageWithDriver(dsn, "mysql", nil) // nil = no New Relic in tests
	if err != nil {
		t.Fatalf("Failed to create MySQL storage: %v", err)
	}
	defer storage.Close()

	userID := "test_user_2"

	// Create test user first
	if err := createTestUser(storage.db, userID, "test_user_2@example.com", "Test User 2"); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create default list first
	list, err := storage.CreateList(userID, "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Add multiple favorites
	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "chart_test_2",
			Type:        models.AssetTypeChart,
			Description: "Chart 2",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Title: "Chart 2",
	}

	insight := &models.Insight{
		BaseAsset: models.BaseAsset{
			ID:          "insight_test_1",
			Type:        models.AssetTypeInsight,
			Description: "Insight 1",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Text: "Test insight",
	}

	_, err = storage.CreateAsset(userID, chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	_, err = storage.CreateAsset(userID, insight)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	storage.AddFavorite(userID, chart.GetID(), list.Reference, nil)
	storage.AddFavorite(userID, insight.GetID(), list.Reference, nil)

	favorites, err := storage.GetFavorites(userID, list.Reference)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(favorites) != 2 {
		t.Errorf("Expected 2 favorites, got %d", len(favorites))
	}
}

func TestMySQLStorage_RemoveFavorite(t *testing.T) {
	skipIfMySQLUnavailable(t)

	dsn := getTestMySQLDSN()
	storage, err := NewMySQLStorageWithDriver(dsn, "mysql", nil) // nil = no New Relic in tests
	if err != nil {
		t.Fatalf("Failed to create MySQL storage: %v", err)
	}
	defer storage.Close()

	userID := "test_user_3"
	
	// Create test user first
	if err := createTestUser(storage.db, userID, "test_user_3@example.com", "Test User 3"); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	
	// Create default list first
	list, err := storage.CreateList(userID, "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}
	
	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "chart_test_3",
			Type:        models.AssetTypeChart,
			Description: "Test Chart",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Title: "Sales Chart",
	}

	_, err = storage.CreateAsset(userID, chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	storage.AddFavorite(userID, chart.GetID(), list.Reference, nil)

	err = storage.RemoveFavorite(userID, chart.GetID(), list.Reference)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	favorites, _ := storage.GetFavorites(userID, list.Reference)
	if len(favorites) != 0 {
		t.Errorf("Expected 0 favorites after removal, got %d", len(favorites))
	}
}

func TestMySQLStorage_UpdateAssetDescription(t *testing.T) {
	skipIfMySQLUnavailable(t)

	dsn := getTestMySQLDSN()
	storage, err := NewMySQLStorageWithDriver(dsn, "mysql", nil) // nil = no New Relic in tests
	if err != nil {
		t.Fatalf("Failed to create MySQL storage: %v", err)
	}
	defer storage.Close()

	userID := "test_user_4"
	
	// Create test user first
	if err := createTestUser(storage.db, userID, "test_user_4@example.com", "Test User 4"); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	
	// Create default list first
	list, err := storage.CreateList(userID, "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}
	
	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "chart_test_4",
			Type:        models.AssetTypeChart,
			Description: "Old Description",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Title: "Sales Chart",
	}

	_, err = storage.CreateAsset(userID, chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	storage.AddFavorite(userID, chart.GetID(), list.Reference, nil)

	newDescription := "New Description"
	err = storage.UpdateAssetDescription(chart.GetID(), newDescription)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	favorites, _ := storage.GetFavorites(userID, list.Reference)
	if len(favorites) != 1 {
		t.Fatalf("Expected 1 favorite, got %d", len(favorites))
	}

	if favorites[0].Asset.GetDescription() != newDescription {
		t.Errorf("Expected description %s, got %s", newDescription, favorites[0].Asset.GetDescription())
	}
}

func TestMySQLStorage_GetAsset(t *testing.T) {
	skipIfMySQLUnavailable(t)

	dsn := getTestMySQLDSN()
	storage, err := NewMySQLStorageWithDriver(dsn, "mysql", nil) // nil = no New Relic in tests
	if err != nil {
		t.Fatalf("Failed to create MySQL storage: %v", err)
	}
	defer storage.Close()

	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "chart_test_5",
			Type:        models.AssetTypeChart,
			Description: "Test Chart",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Title: "Sales Chart",
		XAxis: "Month",
		YAxis: "Revenue",
	}

	userID := "test_user_5"
	
	// Create test user first
	if err := createTestUser(storage.db, userID, "test_user_5@example.com", "Test User 5"); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	
	// Create default list first
	list, err := storage.CreateList(userID, "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}
	
	_, err = storage.CreateAsset(userID, chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	storage.AddFavorite(userID, chart.GetID(), list.Reference, nil)

	asset, err := storage.GetAsset(chart.GetID())
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if asset.GetID() != chart.GetID() {
		t.Errorf("Expected asset ID %s, got %s", chart.GetID(), asset.GetID())
	}

	retrievedChart, ok := asset.(*models.Chart)
	if !ok {
		t.Fatal("Expected Chart type")
	}

	if retrievedChart.Title != chart.Title {
		t.Errorf("Expected title %s, got %s", chart.Title, retrievedChart.Title)
	}
}

func TestMySQLStorage_AssetTypes(t *testing.T) {
	skipIfMySQLUnavailable(t)

	dsn := getTestMySQLDSN()
	storage, err := NewMySQLStorageWithDriver(dsn, "mysql", nil) // nil = no New Relic in tests
	if err != nil {
		t.Fatalf("Failed to create MySQL storage: %v", err)
	}
	defer storage.Close()

	userID := "test_user_6"

	// Create test user first
	if err := createTestUser(storage.db, userID, "test_user_6@example.com", "Test User 6"); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create default list first
	list, err := storage.CreateList(userID, "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Test Insight
	insight := &models.Insight{
		BaseAsset: models.BaseAsset{
			ID:          "insight_test_2",
			Type:        models.AssetTypeInsight,
			Description: "Test Insight",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Text: "40% of millennials spend more than 3 hours on social media daily",
	}

	_, err = storage.CreateAsset(userID, insight)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	favorite, err := storage.AddFavorite(userID, insight.GetID(), list.Reference, nil)
	if err != nil {
		t.Fatalf("Failed to add insight: %v", err)
	}

	retrievedInsight, ok := favorite.Asset.(*models.Insight)
	if !ok {
		t.Fatal("Expected Insight type")
	}

	if retrievedInsight.Text != insight.Text {
		t.Errorf("Expected text %s, got %s", insight.Text, retrievedInsight.Text)
	}

	// Test Audience
	gender := models.GenderMale
	hours := 3.0
	purchases := 5
	audience := &models.Audience{
		BaseAsset: models.BaseAsset{
			ID:          "audience_test_1",
			Type:        models.AssetTypeAudience,
			Description: "Test Audience",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		Gender:              &gender,
		BirthCountry:        "USA",
		AgeGroup:            &models.AgeGroup{Min: 24, Max: 35},
		SocialMediaHoursMin: &hours,
		PurchasesLastMonth:  &purchases,
	}

	_, err = storage.CreateAsset(userID, audience)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	favorite, err = storage.AddFavorite(userID, audience.GetID(), list.Reference, nil)
	if err != nil {
		t.Fatalf("Failed to add audience: %v", err)
	}

	retrievedAudience, ok := favorite.Asset.(*models.Audience)
	if !ok {
		t.Fatal("Expected Audience type")
	}

	if retrievedAudience.BirthCountry != audience.BirthCountry {
		t.Errorf("Expected birth country %s, got %s", audience.BirthCountry, retrievedAudience.BirthCountry)
	}

	if retrievedAudience.AgeGroup == nil || retrievedAudience.AgeGroup.Min != 24 {
		t.Errorf("Expected age group min 24, got %v", retrievedAudience.AgeGroup)
	}
}

func TestMySQLStorage_ErrorHandling(t *testing.T) {
	skipIfMySQLUnavailable(t)

	dsn := getTestMySQLDSN()
	storage, err := NewMySQLStorageWithDriver(dsn, "mysql", nil) // nil = no New Relic in tests
	if err != nil {
		t.Fatalf("Failed to create MySQL storage: %v", err)
	}
	defer storage.Close()

	// Test removing non-existent favorite
	err = storage.RemoveFavorite("nonexistent_user", "nonexistent_asset", "list_nonexistent_user_default")
	if err != ErrFavoriteNotFound {
		t.Errorf("Expected ErrFavoriteNotFound, got %v", err)
	}

	// Test updating non-existent asset
	err = storage.UpdateAssetDescription("nonexistent_asset", "description")
	if err != ErrAssetNotFound {
		t.Errorf("Expected ErrAssetNotFound, got %v", err)
	}

	// Test getting non-existent asset
	_, err = storage.GetAsset("nonexistent_asset")
	if err != ErrAssetNotFound {
		t.Errorf("Expected ErrAssetNotFound, got %v", err)
	}
}

func TestMySQLStorage_ConcurrentAccess(t *testing.T) {
	skipIfMySQLUnavailable(t)

	dsn := getTestMySQLDSN()
	storage, err := NewMySQLStorageWithDriver(dsn, "mysql", nil) // nil = no New Relic in tests
	if err != nil {
		t.Fatalf("Failed to create MySQL storage: %v", err)
	}
	defer storage.Close()

	userID := "test_user_concurrent"
	
	// Create test user first
	if err := createTestUser(storage.db, userID, "test_user_concurrent@example.com", "Test User Concurrent"); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	
	// Create default list first
	list, err := storage.CreateList(userID, "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}
	
	done := make(chan bool, 10)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(id int) {
			chart := &models.Chart{
				BaseAsset: models.BaseAsset{
					ID:          "chart_concurrent_" + string(rune('0'+id)),
					Type:        models.AssetTypeChart,
					Description: "Concurrent Chart",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				},
				Title: "Chart",
			}
			_, err = storage.CreateAsset(userID, chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	storage.AddFavorite(userID, chart.GetID(), list.Reference, nil)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	favorites, err := storage.GetFavorites(userID, list.Reference)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(favorites) != 10 {
		t.Errorf("Expected 10 favorites, got %d", len(favorites))
	}
}

