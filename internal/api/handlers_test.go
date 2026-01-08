package api

import (
	"context"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/gwi/platform-go-challenge/internal/models"
	"github.com/gwi/platform-go-challenge/internal/service"
	"github.com/gwi/platform-go-challenge/internal/storage"
)

func setupTestHandler() *Handler {
	storage := storage.NewMemoryStorage()
	service := service.NewFavoritesService(storage)
	return NewHandler(service)
}

func TestHandler_GetAllFavorites_Basic(t *testing.T) {
	handler := setupTestHandler()

	// Create default list first
	list, err := handler.service.CreateList(context.Background(), "user1", "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Create an asset first
	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "user1_favorite1",
			Type:        models.AssetTypeChart,
			Description: "Test Chart",
		},
		Title: "Sales Chart",
	}
	_, err = handler.service.CreateAsset(context.Background(), "user1", chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	// Add a favorite first
	handler.service.AddFavorite(context.Background(),"user1", "user1_favorite1", list.Reference, nil)

	req := httptest.NewRequest("GET", "/api/v1/users/user1/favorites", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response models.PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Page != 1 {
		t.Errorf("Expected page 1, got %d", response.Page)
	}
	if response.TotalCount < 1 {
		t.Errorf("Expected at least 1 favorite, got %d", response.TotalCount)
	}
}

func TestHandler_AddFavoriteToDefault(t *testing.T) {
	handler := setupTestHandler()

	// Create an asset first
	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "user1_favorite1",
			Type:        models.AssetTypeChart,
			Description: "Test Chart",
		},
		Title: "Sales Chart",
	}
	_, err := handler.service.CreateAsset(context.Background(), "user1", chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}

	requestBody := map[string]interface{}{
		"asset_reference": "user1_favorite1",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/v1/users/user1/favorites", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestHandler_RemoveFavoriteFromDefault(t *testing.T) {
	handler := setupTestHandler()

	// Create default list first
	list, err := handler.service.CreateList(context.Background(), "user1", "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Create an asset first
	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "user1_favorite1",
			Type:        models.AssetTypeChart,
			Description: "Test Chart",
		},
		Title: "Sales Chart",
	}
	_, err = handler.service.CreateAsset(context.Background(), "user1", chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	// Add a favorite first
	handler.service.AddFavorite(context.Background(),"user1", "user1_favorite1", list.Reference, nil)

	req := httptest.NewRequest("DELETE", "/api/v1/users/user1/favorites/user1_favorite1", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandler_UpdateDescription(t *testing.T) {
	handler := setupTestHandler()

	// Create default list first
	list, err := handler.service.CreateList(context.Background(), "user1", "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Add a favorite first
	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "chart1",
			Type:        models.AssetTypeChart,
			Description: "Old Description",
		},
		Title: "Sales Chart",
	}
	_, err = handler.service.CreateAsset(context.Background(), "user1", chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	handler.service.AddFavorite(context.Background(),"user1", "chart1", list.Reference, nil)

	requestBody := map[string]interface{}{
		"description": "New Description",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PATCH", "/api/v1/users/user1/favorites/chart1/description", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestHandler_GetAllFavorites(t *testing.T) {
	handler := setupTestHandler()

	// Create two lists
	list1, err := handler.service.CreateList(context.Background(), "user1", "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}
	list2, err := handler.service.CreateList(context.Background(), "user1", "work")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Add favorites to different lists
	chart1 := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "chart1",
			Type:        models.AssetTypeChart,
			Description: "Chart 1",
		},
		Title: "Chart 1",
	}
	chart2 := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "chart2",
			Type:        models.AssetTypeChart,
			Description: "Chart 2",
		},
		Title: "Chart 2",
	}

	_, err = handler.service.CreateAsset(context.Background(), "user1", chart1)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	_, err = handler.service.CreateAsset(context.Background(), "user1", chart2)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	handler.service.AddFavorite(context.Background(),"user1", "chart1", list1.Reference, nil)
	handler.service.AddFavorite(context.Background(),"user1", "chart2", list2.Reference, nil)

	req := httptest.NewRequest("GET", "/api/v1/users/user1/favorites", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response models.PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.TotalCount != 2 {
		t.Errorf("Expected 2 favorites, got %d", response.TotalCount)
	}
}

func TestHandler_GetAllFavorites_Pagination(t *testing.T) {
	handler := setupTestHandler()

	// Create list
	list, err := handler.service.CreateList(context.Background(), "user1", "default")
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
			},
			Title: fmt.Sprintf("Chart %d", i),
		}
		_, err := handler.service.CreateAsset(context.Background(), "user1", chart)
		if err != nil {
			t.Fatalf("Failed to create asset: %v", err)
		}
		handler.service.AddFavorite(context.Background(),"user1", chart.ID, list.Reference, nil)
	}

	// Test first page
	req := httptest.NewRequest("GET", "/api/v1/users/user1/favorites?page=1&page_size=10", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response models.PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Page != 1 {
		t.Errorf("Expected page 1, got %d", response.Page)
	}
	if response.PageSize != 10 {
		t.Errorf("Expected page size 10, got %d", response.PageSize)
	}
	if response.TotalCount != 25 {
		t.Errorf("Expected total count 25, got %d", response.TotalCount)
	}
	if !response.HasNext {
		t.Error("Expected HasNext to be true")
	}
	if response.HasPrev {
		t.Error("Expected HasPrev to be false")
	}
	
	// Verify actual data count matches page size
	var favorites []models.Favorite
	if dataBytes, err := json.Marshal(response.Data); err == nil {
		if err := json.Unmarshal(dataBytes, &favorites); err == nil {
			if len(favorites) != 10 {
				t.Errorf("Expected 10 items in data array, got %d", len(favorites))
			}
		}
	}

	// Test second page
	req2 := httptest.NewRequest("GET", "/api/v1/users/user1/favorites?page=2&page_size=10", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	var response2 models.PaginatedResponse
	json.Unmarshal(w2.Body.Bytes(), &response2)

	if response2.Page != 2 {
		t.Errorf("Expected page 2, got %d", response2.Page)
	}
	if !response2.HasPrev {
		t.Error("Expected HasPrev to be true on page 2")
	}
	
	// Verify actual data count matches page size
	var favorites2 []models.Favorite
	if dataBytes, err := json.Marshal(response2.Data); err == nil {
		if err := json.Unmarshal(dataBytes, &favorites2); err == nil {
			if len(favorites2) != 10 {
				t.Errorf("Expected 10 items in data array on page 2, got %d", len(favorites2))
			}
		}
	}
}

func TestHandler_GetFavorites_Pagination(t *testing.T) {
	handler := setupTestHandler()

	// Create list
	list, err := handler.service.CreateList(context.Background(), "user1", "default")
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
			},
			Title: fmt.Sprintf("Chart %d", i),
		}
		_, err := handler.service.CreateAsset(context.Background(), "user1", chart)
		if err != nil {
			t.Fatalf("Failed to create asset: %v", err)
		}
		handler.service.AddFavorite(context.Background(),"user1", chart.ID, list.Reference, nil)
	}

	// Test first page
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/users/user1/lists/%s/favorites?page=1&page_size=10", list.Reference), nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response models.PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Page != 1 {
		t.Errorf("Expected page 1, got %d", response.Page)
	}
	if response.PageSize != 10 {
		t.Errorf("Expected page size 10, got %d", response.PageSize)
	}
	if response.TotalCount != 25 {
		t.Errorf("Expected total count 25, got %d", response.TotalCount)
	}
	if !response.HasNext {
		t.Error("Expected HasNext to be true")
	}
	if response.HasPrev {
		t.Error("Expected HasPrev to be false")
	}

	// Verify actual data count matches page size
	var favorites []models.Favorite
	if dataBytes, err := json.Marshal(response.Data); err == nil {
		if err := json.Unmarshal(dataBytes, &favorites); err == nil {
			if len(favorites) != 10 {
				t.Errorf("Expected 10 items in data array, got %d", len(favorites))
			}
		}
	}

	// Test second page
	req2 := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/users/user1/lists/%s/favorites?page=2&page_size=10", list.Reference), nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	var response2 models.PaginatedResponse
	json.Unmarshal(w2.Body.Bytes(), &response2)

	if response2.Page != 2 {
		t.Errorf("Expected page 2, got %d", response2.Page)
	}
	if !response2.HasPrev {
		t.Error("Expected HasPrev to be true on page 2")
	}
}

func TestHandler_GetAllLists_Pagination(t *testing.T) {
	handler := setupTestHandler()

	// Create multiple lists
	for i := 1; i <= 15; i++ {
		handler.service.CreateList(context.Background(),"user1", fmt.Sprintf("list%d", i))
	}

	// Test pagination
	req := httptest.NewRequest("GET", "/api/v1/users/user1/lists?page=1&page_size=5", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response models.PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Page != 1 {
		t.Errorf("Expected page 1, got %d", response.Page)
	}
	if response.PageSize != 5 {
		t.Errorf("Expected page size 5, got %d", response.PageSize)
	}
	if response.TotalCount < 15 {
		t.Errorf("Expected at least 15 lists, got %d", response.TotalCount)
	}
}

func TestHandler_HealthCheck(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", response["status"])
	}
}

// ============================================
// Tests for uncovered handlers
// ============================================

func TestHandler_CreateList(t *testing.T) {
	handler := setupTestHandler()

	requestBody := map[string]interface{}{
		"name": "Work",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/v1/users/user1/lists", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var list models.List
	if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if list.Name != "Work" {
		t.Errorf("Expected list name 'Work', got %s", list.Name)
	}
}

func TestHandler_CreateList_EmptyName(t *testing.T) {
	handler := setupTestHandler()

	requestBody := map[string]interface{}{
		"name": "",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/v1/users/user1/lists", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestHandler_CreateList_InvalidJSON(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest("POST", "/api/v1/users/user1/lists", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestHandler_CreateList_EmptyUserID(t *testing.T) {
	handler := setupTestHandler()

	requestBody := map[string]interface{}{
		"name": "Work",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/v1/users//lists", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandler_GetList(t *testing.T) {
	handler := setupTestHandler()

	// Create a list first
	list, err := handler.service.CreateList(context.Background(), "user1", "work")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/users/user1/lists/%s", list.Reference), nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var retrievedList models.List
	if err := json.Unmarshal(w.Body.Bytes(), &retrievedList); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if retrievedList.Reference != list.Reference {
		t.Errorf("Expected list reference %s, got %s", list.Reference, retrievedList.Reference)
	}
}

func TestHandler_GetList_NotFound(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest("GET", "/api/v1/users/user1/lists/nonexistent_list", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestHandler_GetList_EmptyUserID(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest("GET", "/api/v1/users//lists/list1", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandler_GetList_EmptyListReference(t *testing.T) {
	handler := setupTestHandler()

	// Test with empty listReference - chi will parse empty string from URL
	// Use a route that will match but with empty parameter
	req := httptest.NewRequest("GET", "/api/v1/users/user1/lists/", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	// Chi router might return 404 for trailing slash, but handler validates empty param
	// If route doesn't match, we get 404. If it matches with empty param, handler returns 400.
	// For this test, we accept either behavior as the URL is technically invalid
	if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
		t.Errorf("Expected status 400 or 404, got %d", w.Code)
	}
}

func TestHandler_DeleteList(t *testing.T) {
	handler := setupTestHandler()

	// Create a list first
	list, err := handler.service.CreateList(context.Background(), "user1", "work")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Create an asset and add it as a favorite
	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "user1_favorite1",
			Type:        models.AssetTypeChart,
			Description: "Test Chart",
		},
		Title: "Sales Chart",
	}
	_, err = handler.service.CreateAsset(context.Background(), "user1", chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	_, err = handler.service.AddFavorite(context.Background(), "user1", "user1_favorite1", list.Reference, nil)
	if err != nil {
		t.Fatalf("Failed to add favorite: %v", err)
	}

	// Verify favorite exists
	favorites, err := handler.service.GetFavorites(context.Background(), "user1", list.Reference)
	if err != nil {
		t.Fatalf("Failed to get favorites: %v", err)
	}
	if len(favorites) != 1 {
		t.Fatalf("Expected 1 favorite, got %d", len(favorites))
	}

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/users/user1/lists/%s", list.Reference), nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify list is deleted
	_, err = handler.service.GetList(context.Background(), "user1", list.Reference)
	if err == nil {
		t.Error("Expected list to be deleted, but it still exists")
	}

	// Verify favorites are also deleted (should get error or empty result)
	favorites, err = handler.service.GetFavorites(context.Background(), "user1", list.Reference)
	if err == nil && len(favorites) > 0 {
		t.Errorf("Expected favorites to be deleted with the list, but found %d favorites", len(favorites))
	}
}

func TestHandler_DeleteList_NotFound(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest("DELETE", "/api/v1/users/user1/lists/nonexistent_list", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestHandler_AddFavorite(t *testing.T) {
	handler := setupTestHandler()

	// Create a list first
	list, err := handler.service.CreateList(context.Background(), "user1", "work")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Create an asset first
	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "user1_favorite1",
			Type:        models.AssetTypeChart,
			Description: "Work Chart",
		},
		Title: "Q4 Metrics",
	}
	_, err = handler.service.CreateAsset(context.Background(), "user1", chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}

	requestBody := map[string]interface{}{
		"asset_reference": "user1_favorite1",
		"list_reference":  list.Reference,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/users/user1/lists/%s/favorites", list.Reference), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestHandler_AddFavorite_InvalidJSON(t *testing.T) {
	handler := setupTestHandler()

	// Create a list first
	list, err := handler.service.CreateList(context.Background(), "user1", "work")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/users/user1/lists/%s/favorites", list.Reference), bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandler_AddFavorite_MissingAssetReference(t *testing.T) {
	handler := setupTestHandler()

	// Create a list first
	list, err := handler.service.CreateList(context.Background(), "user1", "work")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	requestBody := map[string]interface{}{
		"list_reference": list.Reference,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/users/user1/lists/%s/favorites", list.Reference), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestHandler_AddFavorite_EmptyUserID(t *testing.T) {
	handler := setupTestHandler()

	requestBody := map[string]interface{}{
		"asset_reference": "user1_favorite1",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/v1/users//lists/list1/favorites", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandler_RemoveFavorite(t *testing.T) {
	handler := setupTestHandler()

	// Create a list first
	list, err := handler.service.CreateList(context.Background(), "user1", "work")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Create an asset first
	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "user1_favorite1",
			Type:        models.AssetTypeChart,
			Description: "Test Chart",
		},
		Title: "Sales Chart",
	}
	_, err = handler.service.CreateAsset(context.Background(), "user1", chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	// Add a favorite first
	handler.service.AddFavorite(context.Background(),"user1", "user1_favorite1", list.Reference, nil)

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/users/user1/lists/%s/favorites/user1_favorite1", list.Reference), nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestHandler_RemoveFavorite_NotFound(t *testing.T) {
	handler := setupTestHandler()

	// Create a list first
	list, err := handler.service.CreateList(context.Background(), "user1", "work")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/users/user1/lists/%s/favorites/nonexistent", list.Reference), nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestHandler_RemoveFavorite_EmptyUserID(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest("DELETE", "/api/v1/users//lists/list1/favorites/chart1", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandler_RemoveFavorite_EmptyAssetID(t *testing.T) {
	handler := setupTestHandler()

	// Create a list first
	list, err := handler.service.CreateList(context.Background(), "user1", "work")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Test with empty assetID - chi router might not match trailing slash
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/users/user1/lists/%s/favorites/", list.Reference), nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	// Chi router might return 405 (Method Not Allowed) or 404 for invalid route
	// Handler validates empty param if route matches. For this test, we accept either behavior
	if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound && w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 400, 404, or 405, got %d", w.Code)
	}
}

func TestHandler_AddFavoriteToDefault_InvalidJSON(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest("POST", "/api/v1/users/user1/favorites", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandler_AddFavoriteToDefault_EmptyUserID(t *testing.T) {
	handler := setupTestHandler()

	requestBody := map[string]interface{}{
		"asset_reference": "user1_favorite1",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/v1/users//favorites", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandler_RemoveFavoriteFromDefault_NotFound(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest("DELETE", "/api/v1/users/user1/favorites/nonexistent", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestHandler_RemoveFavoriteFromDefault_NoDefaultList(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest("DELETE", "/api/v1/users/user1/favorites/chart1", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestHandler_UpdateDescription_InvalidJSON(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest("PATCH", "/api/v1/users/user1/favorites/chart1/description", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandler_UpdateDescription_EmptyDescription(t *testing.T) {
	handler := setupTestHandler()

	requestBody := map[string]interface{}{
		"description": "",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PATCH", "/api/v1/users/user1/favorites/chart1/description", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestHandler_UpdateDescription_NotFound(t *testing.T) {
	handler := setupTestHandler()

	requestBody := map[string]interface{}{
		"description": "New Description",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PATCH", "/api/v1/users/user1/favorites/nonexistent/description", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestHandler_UpdateSortOrder(t *testing.T) {
	handler := setupTestHandler()

	// Create default list first
	list, err := handler.service.CreateList(context.Background(), "user1", "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Add a favorite first
	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "chart1",
			Type:        models.AssetTypeChart,
			Description: "Test Chart",
		},
		Title: "Sales Chart",
	}
	_, err = handler.service.CreateAsset(context.Background(), "user1", chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	fav, err := handler.service.AddFavorite(context.Background(), "user1", chart.ID, list.Reference, nil)
	if err != nil {
		t.Fatalf("Failed to add favorite: %v", err)
	}

	// Update sort order
	requestBody := map[string]interface{}{
		"sort_order": 5,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PATCH", fmt.Sprintf("/api/v1/users/user1/favorites/%s/sort-order", fav.Reference), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Verify the sort order was updated
	favorites, err := handler.service.GetFavorites(context.Background(), "user1", list.Reference)
	if err != nil {
		t.Fatalf("Failed to get favorites: %v", err)
	}

	if len(favorites) != 1 {
		t.Fatalf("Expected 1 favorite, got %d", len(favorites))
	}

	if favorites[0].SortOrder != 5 {
		t.Errorf("Expected sort_order 5, got %d", favorites[0].SortOrder)
	}
}

func TestHandler_UpdateSortOrder_InvalidJSON(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest("PATCH", "/api/v1/users/user1/favorites/fav_user1_chart1/sort-order", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandler_UpdateSortOrder_NegativeSortOrder(t *testing.T) {
	handler := setupTestHandler()

	// Create default list first
	list, err := handler.service.CreateList(context.Background(), "user1", "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Add a favorite first
	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "chart1",
			Type:        models.AssetTypeChart,
			Description: "Test Chart",
		},
		Title: "Sales Chart",
	}
	_, err = handler.service.CreateAsset(context.Background(), "user1", chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	fav, err := handler.service.AddFavorite(context.Background(), "user1", chart.ID, list.Reference, nil)
	if err != nil {
		t.Fatalf("Failed to add favorite: %v", err)
	}

	requestBody := map[string]interface{}{
		"sort_order": -1,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PATCH", fmt.Sprintf("/api/v1/users/user1/favorites/%s/sort-order", fav.Reference), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestHandler_UpdateSortOrder_NotFound(t *testing.T) {
	handler := setupTestHandler()

	requestBody := map[string]interface{}{
		"sort_order": 5,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PATCH", "/api/v1/users/user1/favorites/nonexistent/sort-order", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d. Body: %s", w.Code, w.Body.String())
	}
}

func TestHandler_GetFavorites_EmptyList(t *testing.T) {
	handler := setupTestHandler()

	// Create a list but don't add any favorites
	list, err := handler.service.CreateList(context.Background(), "user1", "work")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/users/user1/lists/%s/favorites", list.Reference), nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response models.PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.TotalCount != 0 {
		t.Errorf("Expected 0 favorites, got %d", response.TotalCount)
	}
}

func TestHandler_GetAllFavorites_Empty(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest("GET", "/api/v1/users/user1/favorites", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response models.PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.TotalCount != 0 {
		t.Errorf("Expected 0 favorites, got %d", response.TotalCount)
	}
}

func TestHandler_GetAllFavorites_EmptyUserID(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest("GET", "/api/v1/users//favorites", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandler_GetFavorites_EmptyUserID(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest("GET", "/api/v1/users//lists/list1/favorites", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandler_parseIntQueryParam(t *testing.T) {
	// Test default value
	req := httptest.NewRequest("GET", "/api/v1/users/user1/favorites", nil)
	page := parseIntQueryParam(req, "page", 1)
	if page != 1 {
		t.Errorf("Expected default page 1, got %d", page)
	}

	// Test valid value
	req = httptest.NewRequest("GET", "/api/v1/users/user1/favorites?page=5", nil)
	page = parseIntQueryParam(req, "page", 1)
	if page != 5 {
		t.Errorf("Expected page 5, got %d", page)
	}

	// Test invalid value (should use default)
	req = httptest.NewRequest("GET", "/api/v1/users/user1/favorites?page=abc", nil)
	page = parseIntQueryParam(req, "page", 1)
	if page != 1 {
		t.Errorf("Expected default page 1 for invalid value, got %d", page)
	}

	// Test zero value (should use default)
	req = httptest.NewRequest("GET", "/api/v1/users/user1/favorites?page=0", nil)
	page = parseIntQueryParam(req, "page", 1)
	if page != 1 {
		t.Errorf("Expected default page 1 for zero value, got %d", page)
	}

	// Test negative value (should use default)
	req = httptest.NewRequest("GET", "/api/v1/users/user1/favorites?page=-1", nil)
	page = parseIntQueryParam(req, "page", 1)
	if page != 1 {
		t.Errorf("Expected default page 1 for negative value, got %d", page)
	}
}

func TestHandler_Pagination_BoundaryConditions(t *testing.T) {
	handler := setupTestHandler()

	// Create list
	list, err := handler.service.CreateList(context.Background(), "user1", "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Add exactly 20 favorites (default page size)
	for i := 1; i <= 20; i++ {
		chart := &models.Chart{
			BaseAsset: models.BaseAsset{
				ID:          fmt.Sprintf("chart%d", i),
				Type:        models.AssetTypeChart,
				Description: fmt.Sprintf("Chart %d", i),
			},
			Title: fmt.Sprintf("Chart %d", i),
		}
		_, err := handler.service.CreateAsset(context.Background(), "user1", chart)
		if err != nil {
			t.Fatalf("Failed to create asset: %v", err)
		}
		handler.service.AddFavorite(context.Background(),"user1", chart.ID, list.Reference, nil)
	}

	// Test page beyond total pages
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/users/user1/lists/%s/favorites?page=999&page_size=10", list.Reference), nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response models.PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if len(response.Data.([]interface{})) != 0 {
		t.Errorf("Expected empty data for page beyond total, got %d items", len(response.Data.([]interface{})))
	}
}

func TestHandler_GetAllLists_EmptyUserID(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest("GET", "/api/v1/users//lists", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandler_GetAllLists_Empty(t *testing.T) {
	handler := setupTestHandler()

	req := httptest.NewRequest("GET", "/api/v1/users/user1/lists", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response models.PaginatedResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// GetAllLists automatically creates a default list if none exist
	if response.TotalCount != 1 {
		t.Errorf("Expected 1 default list, got %d", response.TotalCount)
	}

	// Verify the list is the default list
	var lists []*models.List
	if dataBytes, err := json.Marshal(response.Data); err == nil {
		if err := json.Unmarshal(dataBytes, &lists); err == nil {
			if len(lists) != 1 {
				t.Errorf("Expected 1 list in response, got %d", len(lists))
			}
			if lists[0].Name != "default" {
				t.Errorf("Expected default list name 'default', got %s", lists[0].Name)
			}
		}
	}
}



// TestHandler_GetFavorites_WithFilters tests filtering by asset type
func TestHandler_GetFavorites_WithFilters(t *testing.T) {
	handler := setupTestHandler()

	// Create list
	list, err := handler.service.CreateList(context.Background(), "user1", "default")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Add different asset types
	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "chart1",
			Type:        models.AssetTypeChart,
			Description: "Chart 1",
		},
		Title: "Chart 1",
	}
	_, err = handler.service.CreateAsset(context.Background(), "user1", chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	handler.service.AddFavorite(context.Background(),"user1", "chart1", list.Reference, nil)

	insight := &models.Insight{
		BaseAsset: models.BaseAsset{
			ID:          "insight1",
			Type:        models.AssetTypeInsight,
			Description: "Insight 1",
		},
		Text: "Test insight",
	}
	_, err = handler.service.CreateAsset(context.Background(), "user1", insight)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	handler.service.AddFavorite(context.Background(),"user1", "insight1", list.Reference, nil)

	// Test filter by asset type
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/users/user1/lists/%s/favorites?asset_type=chart", list.Reference), nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response models.PaginatedResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.TotalCount != 1 {
		t.Errorf("Expected 1 chart favorite, got %d", response.TotalCount)
	}
}

// TestHandler_GetAllFavorites_WithFilters tests filtering across all lists
func TestHandler_GetAllFavorites_WithFilters(t *testing.T) {
	handler := setupTestHandler()

	// Create lists
	list1, err := handler.service.CreateList(context.Background(), "user1", "list1")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}
	list2, err := handler.service.CreateList(context.Background(), "user1", "list2")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	// Add different asset types
	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "chart1",
			Type:        models.AssetTypeChart,
			Description: "Chart 1",
		},
		Title: "Chart 1",
	}
	_, err = handler.service.CreateAsset(context.Background(), "user1", chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	handler.service.AddFavorite(context.Background(),"user1", "chart1", list1.Reference, nil)

	insight := &models.Insight{
		BaseAsset: models.BaseAsset{
			ID:          "insight1",
			Type:        models.AssetTypeInsight,
			Description: "Insight 1",
		},
		Text: "Test insight",
	}
	_, err = handler.service.CreateAsset(context.Background(), "user1", insight)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	handler.service.AddFavorite(context.Background(),"user1", "insight1", list2.Reference, nil)

	// Test filter by asset type
	req := httptest.NewRequest("GET", "/api/v1/users/user1/favorites?asset_type=chart", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response models.PaginatedResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.TotalCount != 1 {
		t.Errorf("Expected 1 chart favorite, got %d", response.TotalCount)
	}
}

// Asset Management Tests

func TestHandler_CreateAsset(t *testing.T) {
	handler := setupTestHandler()

	assetData := map[string]interface{}{
		"type":        "chart",
		"description": "Test Chart",
		"title":       "Sales Chart",
		"x_axis":      "Month",
		"y_axis":      "Revenue",
	}

	requestBody := map[string]interface{}{
		"user_id": "user1",
		"asset":   assetData,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/v1/assets", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var asset models.Chart
	if err := json.NewDecoder(w.Body).Decode(&asset); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if asset.ID == "" {
		t.Error("Expected asset ID to be generated")
	}
	if asset.Type != models.AssetTypeChart {
		t.Errorf("Expected type chart, got %s", asset.Type)
	}
	if asset.Title != "Sales Chart" {
		t.Errorf("Expected title 'Sales Chart', got %s", asset.Title)
	}
}

func TestHandler_GetAsset(t *testing.T) {
	handler := setupTestHandler()

	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "user1_favorite1",
			Type:        models.AssetTypeChart,
			Description: "Test Chart",
		},
		Title: "Sales Chart",
	}
	_, err := handler.service.CreateAsset(context.Background(), "user1", chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/v1/assets/user1_favorite1", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var asset models.Chart
	if err := json.NewDecoder(w.Body).Decode(&asset); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if asset.ID != "user1_favorite1" {
		t.Errorf("Expected ID 'user1_favorite1', got %s", asset.ID)
	}
}

func TestHandler_GetAllAssets(t *testing.T) {
	handler := setupTestHandler()

	// Create multiple assets
	chart1 := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "user1_favorite1",
			Type:        models.AssetTypeChart,
			Description: "Chart 1",
		},
		Title: "Sales Chart 1",
	}
	chart2 := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "user1_favorite2",
			Type:        models.AssetTypeChart,
			Description: "Chart 2",
		},
		Title: "Sales Chart 2",
	}
	insight1 := &models.Insight{
		BaseAsset: models.BaseAsset{
			ID:          "user2_favorite1",
			Type:        models.AssetTypeInsight,
			Description: "Insight 1",
		},
		Text: "Some insight text",
	}

	_, err := handler.service.CreateAsset(context.Background(), "user1", chart1)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	_, err = handler.service.CreateAsset(context.Background(), "user1", chart2)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	_, err = handler.service.CreateAsset(context.Background(), "user2", insight1)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}

	// Test without pagination
	req := httptest.NewRequest("GET", "/api/v1/assets", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response models.PaginatedResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.TotalCount < 3 {
		t.Errorf("Expected at least 3 assets, got %d", response.TotalCount)
	}
	if response.Page != 1 {
		t.Errorf("Expected page 1, got %d", response.Page)
	}
	if response.PageSize != 20 {
		t.Errorf("Expected page size 20, got %d", response.PageSize)
	}
}

func TestHandler_GetAllAssets_Pagination(t *testing.T) {
	handler := setupTestHandler()

	// Create multiple assets
	for i := 1; i <= 5; i++ {
		chart := &models.Chart{
			BaseAsset: models.BaseAsset{
				ID:          fmt.Sprintf("user1_favorite%d", i),
				Type:        models.AssetTypeChart,
				Description: fmt.Sprintf("Chart %d", i),
			},
			Title: fmt.Sprintf("Sales Chart %d", i),
		}
		_, err := handler.service.CreateAsset(context.Background(), "user1", chart)
		if err != nil {
			t.Fatalf("Failed to create asset: %v", err)
		}
	}

	// Test pagination - page 1, size 2
	req := httptest.NewRequest("GET", "/api/v1/assets?page=1&page_size=2", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response models.PaginatedResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Page != 1 {
		t.Errorf("Expected page 1, got %d", response.Page)
	}
	if response.PageSize != 2 {
		t.Errorf("Expected page size 2, got %d", response.PageSize)
	}
	if len(response.Data.([]interface{})) != 2 {
		t.Errorf("Expected 2 assets, got %d", len(response.Data.([]interface{})))
	}
	if !response.HasNext {
		t.Error("Expected HasNext to be true")
	}
	if response.HasPrev {
		t.Error("Expected HasPrev to be false")
	}
}

func TestHandler_GetAllAssets_WithFilters(t *testing.T) {
	handler := setupTestHandler()

	// Create assets of different types
	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "user1_favorite1",
			Type:        models.AssetTypeChart,
			Description: "Chart",
		},
		Title: "Sales Chart",
	}
	insight := &models.Insight{
		BaseAsset: models.BaseAsset{
			ID:          "user1_favorite2",
			Type:        models.AssetTypeInsight,
			Description: "Insight",
		},
		Text: "Some insight",
	}

	_, err := handler.service.CreateAsset(context.Background(), "user1", chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}
	_, err = handler.service.CreateAsset(context.Background(), "user1", insight)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}

	// Test filter by asset type
	req := httptest.NewRequest("GET", "/api/v1/assets?asset_type=chart", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response models.PaginatedResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Should have at least one chart
	if response.TotalCount < 1 {
		t.Errorf("Expected at least 1 chart, got %d", response.TotalCount)
	}
}

func TestHandler_UpdateAsset(t *testing.T) {
	handler := setupTestHandler()

	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "user1_favorite1",
			Type:        models.AssetTypeChart,
			Description: "Original Description",
		},
		Title: "Original Title",
	}
	_, err := handler.service.CreateAsset(context.Background(), "user1", chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}

	assetData := map[string]interface{}{
		"id":          "user1_favorite1",
		"type":        "chart",
		"description": "Updated Description",
		"title":       "Updated Title",
		"x_axis":      "Month",
		"y_axis":      "Revenue",
	}

	requestBody := map[string]interface{}{
		"asset": assetData,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("PATCH", "/api/v1/assets/user1_favorite1", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var asset models.Chart
	if err := json.NewDecoder(w.Body).Decode(&asset); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if asset.Description != "Updated Description" {
		t.Errorf("Expected description 'Updated Description', got %s", asset.Description)
	}
}

func TestHandler_DeleteAsset(t *testing.T) {
	handler := setupTestHandler()

	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "user1_favorite1",
			Type:        models.AssetTypeChart,
			Description: "Test Chart",
		},
		Title: "Sales Chart",
	}
	_, err := handler.service.CreateAsset(context.Background(), "user1", chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}

	req := httptest.NewRequest("DELETE", "/api/v1/assets/user1_favorite1", nil)
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	_, err = handler.service.GetAsset(context.Background(), "user1_favorite1")
	if err == nil {
		t.Error("Expected asset to be deleted")
	}
}

func TestHandler_AddFavorite_WithAssetReference(t *testing.T) {
	handler := setupTestHandler()

	list, err := handler.service.CreateList(context.Background(), "user1", "work")
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	chart := &models.Chart{
		BaseAsset: models.BaseAsset{
			ID:          "user1_favorite1",
			Type:        models.AssetTypeChart,
			Description: "Work Chart",
		},
		Title: "Q4 Metrics",
	}
	_, err = handler.service.CreateAsset(context.Background(), "user1", chart)
	if err != nil {
		t.Fatalf("Failed to create asset: %v", err)
	}

	requestBody := map[string]interface{}{
		"asset_reference": "user1_favorite1",
		"list_reference":  list.Reference,
		"sort_order":       0,
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/users/user1/lists/%s/favorites", list.Reference), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if assetID, ok := response["asset_id"].(string); !ok || assetID != "user1_favorite1" {
		t.Errorf("Expected asset ID 'user1_favorite1', got %v", response["asset_id"])
	}
}
