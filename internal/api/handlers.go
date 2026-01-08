package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gwi/platform-go-challenge/internal/errors"
	"github.com/gwi/platform-go-challenge/internal/models"
	"github.com/gwi/platform-go-challenge/internal/service"
)

// Handler handles HTTP requests
type Handler struct {
	service *service.FavoritesService
}

// NewHandler creates a new handler instance
func NewHandler(service *service.FavoritesService) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers all API routes
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1", func(r chi.Router) {
		// Asset management operations (outside of users)
		r.Route("/assets", func(r chi.Router) {
			r.Post("/", h.CreateAsset)
			r.Get("/", h.GetAllAssets)
			r.Get("/{assetReference}", h.GetAsset)
			r.Patch("/{assetReference}", h.UpdateAsset)
			r.Delete("/{assetReference}", h.DeleteAsset)
		})
		
		r.Route("/users/{userID}", func(r chi.Router) {
			// Get all lists for a user
			r.Get("/lists", h.GetAllLists)
			
			// List operations
			r.Post("/lists", h.CreateList)
			r.Get("/lists/{listReference}", h.GetList)
			r.Delete("/lists/{listReference}", h.DeleteList)
			
			// Favorites operations with list support
			r.Route("/lists/{listReference}/favorites", func(r chi.Router) {
				r.Get("/", h.GetFavorites)
				r.Post("/", h.AddFavorite)
				r.Delete("/{assetID}", h.RemoveFavorite)
			})
			
			// Favorites operations (all favorites across all lists with pagination)
			r.Route("/favorites", func(r chi.Router) {
				r.Get("/", h.GetAllFavorites)
				r.Post("/", h.AddFavoriteToDefault)
				r.Delete("/{assetID}", h.RemoveFavoriteFromDefault)
				r.Patch("/{assetID}/description", h.UpdateDescription)
				r.Patch("/{favoriteReference}/sort-order", h.UpdateSortOrder)
			})
		})
		r.Get("/health", h.HealthCheck)
	})
}

// GetAllLists retrieves all favorites lists for a user (with pagination)
func (h *Handler) GetAllLists(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		respondWithErrorLegacy(w, http.StatusBadRequest, "user ID is required")
		return
	}

	// Parse pagination parameters
	page := parseIntQueryParam(r, "page", 1)
	pageSize := parseIntQueryParam(r, "page_size", 20)

	lists, totalCount, err := h.service.GetAllListsPaginated(ctx, userID, page, pageSize)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	respondWithJSON(w, http.StatusOK, models.PaginatedResponse{
		Data:       lists,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	})
}

// CreateList creates a new list for a user
func (h *Handler) CreateList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		respondWithErrorLegacy(w, http.StatusBadRequest, "user ID is required")
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithErrorLegacy(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if req.Name == "" {
		respondWithErrorLegacy(w, http.StatusBadRequest, "list name is required")
		return
	}

	list, err := h.service.CreateList(ctx, userID, req.Name)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	respondWithJSON(w, http.StatusCreated, list)
}

// GetList retrieves a list by reference
func (h *Handler) GetList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "userID")
	listReference := chi.URLParam(r, "listReference")
	if userID == "" {
		respondWithErrorLegacy(w, http.StatusBadRequest, "user ID is required")
		return
	}
	if listReference == "" {
		respondWithErrorLegacy(w, http.StatusBadRequest, "list reference is required")
		return
	}

	list, err := h.service.GetList(ctx, userID, listReference)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	respondWithJSON(w, http.StatusOK, list)
}

// DeleteList deletes a list and all its favorites
func (h *Handler) DeleteList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "userID")
	listReference := chi.URLParam(r, "listReference")
	if userID == "" {
		respondWithErrorLegacy(w, http.StatusBadRequest, "user ID is required")
		return
	}
	if listReference == "" {
		respondWithErrorLegacy(w, http.StatusBadRequest, "list reference is required")
		return
	}

	if err := h.service.DeleteList(ctx, userID, listReference); err != nil {
		HandleError(w, r, err)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "list deleted successfully",
	})
}

// GetFavorites retrieves paginated favorites for a user in a specific list
func (h *Handler) GetFavorites(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "userID")
	listReference := chi.URLParam(r, "listReference")
	if userID == "" {
		respondWithErrorLegacy(w, http.StatusBadRequest, "user ID is required")
		return
	}

	// Parse pagination parameters
	page := parseIntQueryParam(r, "page", 1)
	pageSize := parseIntQueryParam(r, "page_size", 20)

	// Parse filter parameters
	filters := parseFilterParams(r)

	favorites, totalCount, err := h.service.GetFavoritesPaginated(ctx, userID, listReference, page, pageSize, filters)
	if err != nil {
		respondWithErrorLegacy(w, http.StatusInternalServerError, err.Error())
		return
	}

	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	respondWithJSON(w, http.StatusOK, models.PaginatedResponse{
		Data:       favorites,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	})
}

// GetAllFavorites retrieves all favorites for a user across all lists (with pagination)
func (h *Handler) GetAllFavorites(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		respondWithErrorLegacy(w, http.StatusBadRequest, "user ID is required")
		return
	}

	// Parse pagination parameters
	page := parseIntQueryParam(r, "page", 1)
	pageSize := parseIntQueryParam(r, "page_size", 20)

	// Parse filter parameters
	filters := parseFilterParams(r)

	favorites, totalCount, err := h.service.GetAllFavoritesPaginated(ctx, userID, page, pageSize, filters)
	if err != nil {
		respondWithErrorLegacy(w, http.StatusInternalServerError, err.Error())
		return
	}

	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	respondWithJSON(w, http.StatusOK, models.PaginatedResponse{
		Data:       favorites,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	})
}

// CreateAsset creates a new asset
func (h *Handler) CreateAsset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req models.CreateAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithErrorLegacy(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	// Get userID from request body
	userID := req.UserID
	if userID == "" {
		respondWithErrorLegacy(w, http.StatusBadRequest, "user_id is required in request body")
		return
	}

	// Parse asset from JSON
	var assetData map[string]interface{}
	if err := json.Unmarshal(req.Asset, &assetData); err != nil {
		respondWithErrorLegacy(w, http.StatusBadRequest, fmt.Sprintf("invalid asset data: %v", err))
		return
	}

	// Remove ID from request if provided - we'll generate it automatically
	delete(assetData, "id")

	// Generate next sequential asset ID for the user
	assetID, err := h.service.GenerateNextAssetID(ctx, userID)
	if err != nil {
		respondWithErrorLegacy(w, http.StatusInternalServerError, fmt.Sprintf("failed to generate asset ID: %v", err))
		return
	}

	// Set the generated ID
	assetData["id"] = assetID

	// Create asset using service
	asset, err := h.service.CreateAssetFromJSON(assetData)
	if err != nil {
		respondWithErrorLegacy(w, http.StatusBadRequest, fmt.Sprintf("failed to create asset: %v", err))
		return
	}

	// Create asset in storage
	createdAsset, err := h.service.CreateAsset(ctx, userID, asset)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	respondWithJSON(w, http.StatusCreated, createdAsset)
}

// GetAllAssets retrieves all assets with pagination and optional filters
func (h *Handler) GetAllAssets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Parse pagination parameters
	page := parseIntQueryParam(r, "page", 1)
	pageSize := parseIntQueryParam(r, "page_size", 20)

	// Parse filter parameters
	filters := parseFilterParams(r)

	assets, totalCount, err := h.service.GetAllAssetsPaginated(ctx, page, pageSize, filters)
	if err != nil {
		respondWithErrorLegacy(w, http.StatusInternalServerError, err.Error())
		return
	}

	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	respondWithJSON(w, http.StatusOK, models.PaginatedResponse{
		Data:       assets,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	})
}

// GetAsset retrieves an asset by reference
func (h *Handler) GetAsset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	assetReference := chi.URLParam(r, "assetReference")
	if assetReference == "" {
		respondWithErrorLegacy(w, http.StatusBadRequest, "asset reference is required")
		return
	}

	asset, err := h.service.GetAsset(ctx, assetReference)
	if err != nil {
		HandleError(w, r, err)
		return
	}

	respondWithJSON(w, http.StatusOK, asset)
}

// UpdateAsset updates an existing asset
func (h *Handler) UpdateAsset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	assetReference := chi.URLParam(r, "assetReference")
	if assetReference == "" {
		respondWithErrorLegacy(w, http.StatusBadRequest, "asset reference is required")
		return
	}

	var req models.UpdateAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithErrorLegacy(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	// Parse asset from JSON
	var assetData map[string]interface{}
	if err := json.Unmarshal(req.Asset, &assetData); err != nil {
		respondWithErrorLegacy(w, http.StatusBadRequest, fmt.Sprintf("invalid asset data: %v", err))
		return
	}

	// Ensure the ID matches the reference
	assetData["id"] = assetReference

	// Create asset using service
	asset, err := h.service.CreateAssetFromJSON(assetData)
	if err != nil {
		respondWithErrorLegacy(w, http.StatusBadRequest, fmt.Sprintf("failed to parse asset: %v", err))
		return
	}

	// Update asset in storage
	updatedAsset, err := h.service.UpdateAsset(ctx, assetReference, asset)
	if err != nil {
		if err.Error() == "asset not found" {
			respondWithErrorLegacy(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithErrorLegacy(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, updatedAsset)
}

// DeleteAsset deletes an asset
func (h *Handler) DeleteAsset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	assetReference := chi.URLParam(r, "assetReference")
	if assetReference == "" {
		respondWithErrorLegacy(w, http.StatusBadRequest, "asset reference is required")
		return
	}

	if err := h.service.DeleteAsset(ctx, assetReference); err != nil {
		if err.Error() == "asset not found" {
			respondWithErrorLegacy(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithErrorLegacy(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "asset deleted successfully",
	})
}

// AddFavorite adds an asset to a user's favorites in a specific list
func (h *Handler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "userID")
	listReference := chi.URLParam(r, "listReference")
	if userID == "" {
		respondWithErrorLegacy(w, http.StatusBadRequest, "user ID is required")
		return
	}

	var req models.AddFavoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithErrorLegacy(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if req.AssetReference == "" {
		respondWithErrorLegacy(w, http.StatusBadRequest, "asset_reference is required")
		return
	}

	// Use listReference from URL if provided, otherwise from request, otherwise default
	if listReference == "" {
		if req.ListReference != "" {
			listReference = req.ListReference
		}
		// If still empty, will default to "default" in service layer
	}

	favorite, err := h.service.AddFavorite(ctx, userID, req.AssetReference, listReference, req.SortOrder)
	if err != nil {
		if err.Error() == "asset with reference "+req.AssetReference+" not found" || err.Error() == "asset not found" {
			respondWithErrorLegacy(w, http.StatusNotFound, "asset not found")
			return
		}
		respondWithErrorLegacy(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, favorite)
}

// AddFavoriteToDefault adds an asset to a user's favorites in the default list
func (h *Handler) AddFavoriteToDefault(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		respondWithErrorLegacy(w, http.StatusBadRequest, "user ID is required")
		return
	}

	var req models.AddFavoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithErrorLegacy(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if req.AssetReference == "" {
		respondWithErrorLegacy(w, http.StatusBadRequest, "asset_reference is required")
		return
	}

	// Use listReference from request if provided, otherwise will default to "default" in service layer
	favorite, err := h.service.AddFavorite(ctx, userID, req.AssetReference, req.ListReference, req.SortOrder)
	if err != nil {
		if err.Error() == "asset with reference "+req.AssetReference+" not found" || err.Error() == "asset not found" {
			respondWithErrorLegacy(w, http.StatusNotFound, "asset not found")
			return
		}
		respondWithErrorLegacy(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, favorite)
}

// RemoveFavorite removes an asset from a user's favorites in a specific list
func (h *Handler) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "userID")
	listReference := chi.URLParam(r, "listReference")
	assetID := chi.URLParam(r, "assetID")

	if userID == "" {
		respondWithErrorLegacy(w, http.StatusBadRequest, "user ID is required")
		return
	}

	if assetID == "" {
		respondWithErrorLegacy(w, http.StatusBadRequest, "asset ID is required")
		return
	}

	// Get or find default list if listReference is empty
	if listReference == "" {
		list, err := h.service.GetListByName(ctx, userID, "default")
		if err != nil {
			respondWithErrorLegacy(w, http.StatusNotFound, "default list not found")
			return
		}
		listReference = list.Reference
	}

	if err := h.service.RemoveFavorite(ctx, userID, assetID, listReference); err != nil {
		if err.Error() == "favorite not found" {
			respondWithErrorLegacy(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithErrorLegacy(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "favorite removed successfully",
	})
}

// RemoveFavoriteFromDefault removes an asset from a user's favorites in the default list
func (h *Handler) RemoveFavoriteFromDefault(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "userID")
	assetID := chi.URLParam(r, "assetID")

	if userID == "" {
		respondWithErrorLegacy(w, http.StatusBadRequest, "user ID is required")
		return
	}

	if assetID == "" {
		respondWithErrorLegacy(w, http.StatusBadRequest, "asset ID is required")
		return
	}

	// Get default list
	list, err := h.service.GetListByName(ctx, userID, "default")
	if err != nil {
		respondWithErrorLegacy(w, http.StatusNotFound, "default list not found")
		return
	}

	if err := h.service.RemoveFavorite(ctx, userID, assetID, list.Reference); err != nil {
		if err.Error() == "favorite not found" {
			respondWithErrorLegacy(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithErrorLegacy(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "favorite removed successfully",
	})
}

// UpdateDescription updates the description of an asset
func (h *Handler) UpdateDescription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	assetID := chi.URLParam(r, "assetID")
	if assetID == "" {
		respondWithErrorLegacy(w, http.StatusBadRequest, "asset ID is required")
		return
	}

	var req models.UpdateDescriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithErrorLegacy(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if err := h.service.UpdateDescription(ctx, assetID, req.Description); err != nil {
		if err.Error() == "asset not found" {
			respondWithErrorLegacy(w, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "description cannot be empty" {
			respondWithErrorLegacy(w, http.StatusBadRequest, err.Error())
			return
		}
		respondWithErrorLegacy(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "description updated successfully",
	})
}

// UpdateSortOrder updates the sort order of a favorite
func (h *Handler) UpdateSortOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		respondWithErrorLegacy(w, http.StatusBadRequest, "user ID is required")
		return
	}

	favoriteReference := chi.URLParam(r, "favoriteReference")
	if favoriteReference == "" {
		respondWithErrorLegacy(w, http.StatusBadRequest, "favorite reference is required")
		return
	}

	var req models.UpdateSortOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithErrorLegacy(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if err := h.service.UpdateFavoriteSortOrder(ctx, userID, favoriteReference, req.SortOrder); err != nil {
		if err.Error() == fmt.Sprintf("favorite with reference %s not found", favoriteReference) || 
		   err.Error() == "favorite not found" {
			respondWithErrorLegacy(w, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "sort order must be non-negative" {
			respondWithErrorLegacy(w, http.StatusBadRequest, err.Error())
			return
		}
		respondWithErrorLegacy(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "sort order updated successfully",
	})
}

// HealthCheck returns the health status of the service
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "favorites-api",
	})
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

// respondWithErrorLegacy is a helper for handler-level validation errors
// For service errors, use HandleError instead
func respondWithErrorLegacy(w http.ResponseWriter, code int, message string) {
	var errorCode errors.ErrorCode
	switch code {
	case http.StatusBadRequest:
		errorCode = errors.ErrorCodeValidation
	case http.StatusNotFound:
		errorCode = errors.ErrorCodeNotFound
	case http.StatusConflict:
		errorCode = errors.ErrorCodeConflict
	case http.StatusInternalServerError:
		errorCode = errors.ErrorCodeInternal
	default:
		errorCode = errors.ErrorCodeInternal
	}
	
	appErr := &errors.AppError{
		Code:       errorCode,
		Message:    message,
		HTTPStatus: code,
	}
	HandleError(w, nil, appErr)
}

// parseIntQueryParam parses an integer query parameter with a default value
func parseIntQueryParam(r *http.Request, key string, defaultValue int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return defaultValue
	}
	return parsed
}

// parseFilterParams parses filter parameters from query string
func parseFilterParams(r *http.Request) *models.FilterParams {
	filters := &models.FilterParams{}
	query := r.URL.Query()

	// Parse asset_type filter
	if assetTypeStr := query.Get("asset_type"); assetTypeStr != "" {
		assetType := models.AssetType(assetTypeStr)
		// Validate asset type
		if assetType == models.AssetTypeChart || assetType == models.AssetTypeInsight || assetType == models.AssetTypeAudience {
			filters.AssetType = &assetType
		}
	}

	// Parse date_from filter (ISO 8601 format: 2006-01-02T15:04:05Z or 2006-01-02)
	if dateFromStr := query.Get("date_from"); dateFromStr != "" {
		dateFrom, err := parseDate(dateFromStr)
		if err == nil {
			filters.DateFrom = &dateFrom
		}
	}

	// Parse date_to filter (ISO 8601 format: 2006-01-02T15:04:05Z or 2006-01-02)
	if dateToStr := query.Get("date_to"); dateToStr != "" {
		dateTo, err := parseDate(dateToStr)
		if err == nil {
			// Set to end of day for date_to
			dateTo = time.Date(dateTo.Year(), dateTo.Month(), dateTo.Day(), 23, 59, 59, 999999999, dateTo.Location())
			filters.DateTo = &dateTo
		}
	}

	// Return nil if no filters were set
	if filters.AssetType == nil && filters.DateFrom == nil && filters.DateTo == nil {
		return nil
	}

	return filters
}

// parseDate parses a date string in various formats
func parseDate(dateStr string) (time.Time, error) {
	// Try ISO 8601 with time
	if t, err := time.Parse(time.RFC3339, dateStr); err == nil {
		return t, nil
	}
	// Try ISO 8601 date only
	if t, err := time.Parse("2006-01-02", dateStr); err == nil {
		return t, nil
	}
	// Try RFC3339Nano
	if t, err := time.Parse(time.RFC3339Nano, dateStr); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid date format: %s", dateStr)
}
