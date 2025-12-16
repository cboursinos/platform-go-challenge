package models

import (
	"encoding/json"
	"time"
)

// AssetType represents the type of asset
type AssetType string

const (
	AssetTypeChart    AssetType = "chart"
	AssetTypeInsight  AssetType = "insight"
	AssetTypeAudience AssetType = "audience"
)

// BaseAsset contains common attributes for all asset types
type BaseAsset struct {
	ID          string    `json:"id"`
	Type        AssetType `json:"type"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Chart represents a chart asset
type Chart struct {
	BaseAsset
	Title    string   `json:"title"`
	XAxis    string   `json:"x_axis"`
	YAxis    string   `json:"y_axis"`
	Data     []DataPoint `json:"data"`
}

// DataPoint represents a single data point in a chart
type DataPoint struct {
	Label string          `json:"label"`
	Value json.RawMessage `json:"value"` // Flexible to support numbers, strings, etc.
}

// Insight represents an insight asset
type Insight struct {
	BaseAsset
	Text string `json:"text"`
}

// Gender represents gender options
type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
)

// AgeGroup represents an age range
type AgeGroup struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// Audience represents an audience asset
type Audience struct {
	BaseAsset
	Gender              *Gender   `json:"gender,omitempty"`
	BirthCountry        string    `json:"birth_country,omitempty"`
	AgeGroup            *AgeGroup `json:"age_group,omitempty"`
	SocialMediaHoursMin *float64  `json:"social_media_hours_min,omitempty"`
	PurchasesLastMonth  *int      `json:"purchases_last_month,omitempty"`
}

// List represents a favorites list for organizing user favorites
type List struct {
	ID        int       `json:"id,omitempty"`        // Internal database ID
	Reference string    `json:"reference"`          // API reference (e.g., "list_user1_default")
	UserID    string    `json:"user_id"`           // User reference (e.g., "user1")
	Name      string    `json:"name"`              // List name (e.g., "default", "work", "personal")
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Favorite represents a user's favorite asset with metadata
// Links a user, asset, and list together
type Favorite struct {
	ID        int       `json:"-"`                  // Internal database ID (not returned in API responses)
	Reference string    `json:"reference"`          // API reference (e.g., "fav_user1_favorite1")
	UserID    string    `json:"user_id"`           // User reference (e.g., "user1")
	AssetID   string    `json:"asset_id"`           // Asset reference (e.g., "chart1")
	ListID    int       `json:"-"`                  // Internal list ID (not returned in API responses)
	ListRef   string    `json:"list_reference"`     // List reference (e.g., "list_user1_default")
	ListName  string    `json:"list_name"`         // List name (e.g., "default", "work", "personal")
	SortOrder int       `json:"sort_order"`         // Sort order within the list (0-based, lower numbers appear first)
	Asset     Asset     `json:"asset"`              // The asset object
	AddedAt   time.Time `json:"added_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Asset is an interface that all asset types implement
type Asset interface {
	GetID() string
	GetType() AssetType
	GetDescription() string
	SetDescription(string)
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	SetUpdatedAt(time.Time)
}

// Ensure all asset types implement the Asset interface
var (
	_ Asset = (*Chart)(nil)
	_ Asset = (*Insight)(nil)
	_ Asset = (*Audience)(nil)
)

// Chart methods
func (c *Chart) GetID() string                { return c.BaseAsset.ID }
func (c *Chart) GetType() AssetType          { return c.BaseAsset.Type }
func (c *Chart) GetDescription() string       { return c.BaseAsset.Description }
func (c *Chart) SetDescription(d string)      { c.BaseAsset.Description = d }
func (c *Chart) GetCreatedAt() time.Time      { return c.BaseAsset.CreatedAt }
func (c *Chart) GetUpdatedAt() time.Time      { return c.BaseAsset.UpdatedAt }
func (c *Chart) SetUpdatedAt(t time.Time)     { c.BaseAsset.UpdatedAt = t }

// Insight methods
func (i *Insight) GetID() string                { return i.BaseAsset.ID }
func (i *Insight) GetType() AssetType          { return i.BaseAsset.Type }
func (i *Insight) GetDescription() string       { return i.BaseAsset.Description }
func (i *Insight) SetDescription(d string)      { i.BaseAsset.Description = d }
func (i *Insight) GetCreatedAt() time.Time      { return i.BaseAsset.CreatedAt }
func (i *Insight) GetUpdatedAt() time.Time      { return i.BaseAsset.UpdatedAt }
func (i *Insight) SetUpdatedAt(t time.Time)     { i.BaseAsset.UpdatedAt = t }

// Audience methods
func (a *Audience) GetID() string                { return a.BaseAsset.ID }
func (a *Audience) GetType() AssetType          { return a.BaseAsset.Type }
func (a *Audience) GetDescription() string       { return a.BaseAsset.Description }
func (a *Audience) SetDescription(d string)      { a.BaseAsset.Description = d }
func (a *Audience) GetCreatedAt() time.Time      { return a.BaseAsset.CreatedAt }
func (a *Audience) GetUpdatedAt() time.Time      { return a.BaseAsset.UpdatedAt }
func (a *Audience) SetUpdatedAt(t time.Time)     { a.BaseAsset.UpdatedAt = t }

// Request/Response DTOs


// UpdateDescriptionRequest represents a request to update an asset description
type UpdateDescriptionRequest struct {
	Description string `json:"description"`
}

// UpdateSortOrderRequest represents a request to update a favorite's sort order
type UpdateSortOrderRequest struct {
	SortOrder int `json:"sort_order"` // New sort order value
}

// CreateAssetRequest represents a request to create an asset
type CreateAssetRequest struct {
	UserID string          `json:"user_id"` // User reference (e.g., "user1")
	Asset  json.RawMessage `json:"asset"`    // Raw JSON to parse dynamically
}

// UpdateAssetRequest represents a request to update an asset
type UpdateAssetRequest struct {
	Asset json.RawMessage `json:"asset"` // Raw JSON to parse dynamically
}

// AddFavoriteRequest represents a request to add a favorite (updated to use references only)
type AddFavoriteRequest struct {
	AssetReference string `json:"asset_reference"` // Asset reference (e.g., "user1_favorite1")
	ListReference  string `json:"list_reference,omitempty"` // Optional list reference, defaults to "default"
	SortOrder      *int   `json:"sort_order,omitempty"` // Optional sort order, defaults to next sequential value
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page     int `json:"page"`      // 1-based page number
	PageSize int `json:"page_size"` // Number of items per page
}

// FilterParams represents filtering parameters for favorites
type FilterParams struct {
	AssetType *AssetType `json:"asset_type,omitempty"` // Filter by asset type (chart, insight, audience)
	DateFrom  *time.Time `json:"date_from,omitempty"`  // Filter favorites added after this date
	DateTo    *time.Time `json:"date_to,omitempty"`    // Filter favorites added before this date
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
	TotalCount int          `json:"total_count"`
	TotalPages int          `json:"total_pages"`
	HasNext    bool         `json:"has_next"`
	HasPrev    bool         `json:"has_prev"`
}

