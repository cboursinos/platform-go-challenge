package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gwi/platform-go-challenge/internal/errors"
	"github.com/gwi/platform-go-challenge/internal/models"
)

// CreateAssetFromJSONImproved creates an asset from JSON data using json.Unmarshal
// This is an improved version that uses the standard library's JSON unmarshaling
// instead of manual field extraction.
func (s *FavoritesService) CreateAssetFromJSONImproved(data map[string]interface{}) (models.Asset, error) {
	// Convert map to JSON bytes for unmarshaling
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, errors.NewValidationError("invalid JSON structure", err)
	}

	// First pass: determine asset type using a lightweight discriminator
	type typeDiscriminator struct {
		Type models.AssetType `json:"type"`
	}
	var disc typeDiscriminator
	if err := json.Unmarshal(jsonBytes, &disc); err != nil {
		return nil, errors.NewValidationError("failed to determine asset type", err)
	}

	if disc.Type == "" {
		return nil, errors.NewValidationError("asset type is required", nil)
	}

	// Second pass: unmarshal to specific type based on discriminator
	now := time.Now()
	switch disc.Type {
	case models.AssetTypeChart:
		var chart models.Chart
		if err := json.Unmarshal(jsonBytes, &chart); err != nil {
			return nil, errors.NewValidationError("invalid chart data", err)
		}
		// Set timestamps (not provided in JSON)
		chart.CreatedAt = now
		chart.UpdatedAt = now
		return &chart, nil

	case models.AssetTypeInsight:
		var insight models.Insight
		if err := json.Unmarshal(jsonBytes, &insight); err != nil {
			return nil, errors.NewValidationError("invalid insight data", err)
		}
		insight.CreatedAt = now
		insight.UpdatedAt = now
		return &insight, nil

	case models.AssetTypeAudience:
		var audience models.Audience
		if err := json.Unmarshal(jsonBytes, &audience); err != nil {
			return nil, errors.NewValidationError("invalid audience data", err)
		}
		audience.CreatedAt = now
		audience.UpdatedAt = now
		return &audience, nil

	default:
		return nil, errors.NewValidationError(
			fmt.Sprintf("unknown asset type: %s", disc.Type), nil)
	}
}

// Alternative: Direct unmarshaling from json.RawMessage
// This would be even better if we can change the handler to pass json.RawMessage
func (s *FavoritesService) CreateAssetFromJSONBytes(jsonBytes json.RawMessage) (models.Asset, error) {
	// First pass: get type
	type typeDiscriminator struct {
		Type models.AssetType `json:"type"`
	}
	var disc typeDiscriminator
	if err := json.Unmarshal(jsonBytes, &disc); err != nil {
		return nil, errors.NewValidationError("failed to determine asset type", err)
	}

	if disc.Type == "" {
		return nil, errors.NewValidationError("asset type is required", nil)
	}

	// Second pass: unmarshal to specific type
	now := time.Now()
	switch disc.Type {
	case models.AssetTypeChart:
		var chart models.Chart
		if err := json.Unmarshal(jsonBytes, &chart); err != nil {
			return nil, errors.NewValidationError("invalid chart data", err)
		}
		chart.CreatedAt = now
		chart.UpdatedAt = now
		return &chart, nil

	case models.AssetTypeInsight:
		var insight models.Insight
		if err := json.Unmarshal(jsonBytes, &insight); err != nil {
			return nil, errors.NewValidationError("invalid insight data", err)
		}
		insight.CreatedAt = now
		insight.UpdatedAt = now
		return &insight, nil

	case models.AssetTypeAudience:
		var audience models.Audience
		if err := json.Unmarshal(jsonBytes, &audience); err != nil {
			return nil, errors.NewValidationError("invalid audience data", err)
		}
		audience.CreatedAt = now
		audience.UpdatedAt = now
		return &audience, nil

	default:
		return nil, errors.NewValidationError(
			fmt.Sprintf("unknown asset type: %s", disc.Type), nil)
	}
}
