package service

import (
	"context"
	"fmt"

	"emissions-cache-service/internal/client/scope3"
	"emissions-cache-service/internal/errors"
	"emissions-cache-service/internal/models"

	"github.com/google/uuid"
)

// CacheRepository abstracts the cache implementation.
type CacheRepository interface {
	Set(key string, value interface{}, isPriority bool)
	Get(key string) (interface{}, bool)
}

// Scope3Client abstracts the Scope3 API client.
type Scope3Client interface {
	GetEmissions(ctx context.Context, req scope3.MeasureRequest) (*scope3.MeasureResponse, error)
}

// MeasureService defines the interface for the emissions measurement service.
type MeasureService interface {
	GetMeasure(ctx context.Context, req models.MeasureRequest) (*models.MeasureResponse, error)
}

// measureService implements the MeasureService interface.
type measureService struct {
	cache        CacheRepository
	scope3Client Scope3Client
}

// NewMeasureService creates a new instance of measureService.
func NewMeasureService(cache CacheRepository, client Scope3Client) MeasureService {
	return &measureService{
		cache:        cache,
		scope3Client: client,
	}
}

// generateCacheKey creates a composite key for caching based on key fields.
func generateCacheKey(row models.MeasureRow) string {
	return fmt.Sprintf("%s-%s-%d-%s", row.Country, row.Channel, row.Impressions, row.InventoryID)
}

// GetMeasure retrieves emissions data, either from cache or via the Scope3 API.
func (m *measureService) GetMeasure(ctx context.Context, req models.MeasureRequest) (*models.MeasureResponse, error) {
	if len(req.Rows) == 0 {
		return nil, errors.NewValidationError("no rows provided in request")
	}

	// Validate each row.
	for i, row := range req.Rows {
		if row.Country == "" {
			return nil, errors.NewValidationError(fmt.Sprintf("country is required for row %d", i))
		}
		if row.Channel == "" {
			return nil, errors.NewValidationError(fmt.Sprintf("channel is required for row %d", i))
		}
		if row.Impressions <= 0 {
			return nil, errors.NewValidationError(fmt.Sprintf("impressions must be positive for row %d", i))
		}
		if row.InventoryID == "" {
			return nil, errors.NewValidationError(fmt.Sprintf("inventoryId is required for row %d", i))
		}
	}

	// Map original rows by composite key.
	originalRowsMap := make(map[string]models.MeasureRow)
	for _, row := range req.Rows {
		key := generateCacheKey(row)
		originalRowsMap[key] = row
	}

	var modelRows []models.MeasureRowResponse
	var uncachedRows []scope3.MeasureRow

	// Check the cache for each row.
	for _, row := range req.Rows {
		key := generateCacheKey(row)
		if cachedValue, found := m.cache.Get(key); found {
			if cachedRow, ok := cachedValue.(scope3.MeasureRowResponse); ok {
				modelRows = append(modelRows, models.MeasureRowResponse{
					PropertyID:     cachedRow.Internal.PropertyID,
					PropertyName:   cachedRow.Internal.PropertyName,
					TotalEmissions: cachedRow.TotalEmissions,
					Cached:         true,
				})
				continue
			}
		}
		// Prepare row for API request if not cached.
		uncachedRows = append(uncachedRows, scope3.MeasureRow{
			Country:     row.Country,
			Channel:     row.Channel,
			Impressions: row.Impressions,
			InventoryID: row.InventoryID,
			UTCDatetime: row.UTCDatetime,
		})
	}

	// Generate a unique request ID for tracking.
	requestID := uuid.New().String()

	// If all rows are cached, return the aggregated response immediately.
	if len(uncachedRows) == 0 {
		return &models.MeasureResponse{
			RequestID:      requestID,
			TotalEmissions: sumEmissions(modelRows),
			Rows:           modelRows,
		}, nil
	}

	// Call Scope3 API for rows that are not cached.
	apiResponse, err := m.scope3Client.GetEmissions(ctx, scope3.MeasureRequest{Rows: uncachedRows})
	if err != nil {
		return nil, errors.NewExternalError("failed to fetch emissions data from Scope3", err)
	}

	// Process API response and update the cache.
	for i, apiRow := range apiResponse.Rows {
		uncachedRow := uncachedRows[i]
		key := generateCacheKey(models.MeasureRow{
			Country:     uncachedRow.Country,
			Channel:     uncachedRow.Channel,
			Impressions: uncachedRow.Impressions,
			InventoryID: uncachedRow.InventoryID,
		})
		originalRow, exists := originalRowsMap[key]
		if !exists {
			return nil, errors.NewInternalError(
				"cache key mismatch",
				fmt.Errorf("could not find original row for key: %s", key),
			)
		}
		m.cache.Set(key, apiRow, originalRow.IsPriority)
		// If the API indicates missing inventory coverage, mark accordingly.
		if apiRow.InventoryCoverage == "missing" {
			modelRows = append(modelRows, models.MeasureRowResponse{
				InventoryCoverage: "missing",
				Cached:            false,
			})
		} else {
			modelRows = append(modelRows, models.MeasureRowResponse{
				PropertyID:     apiRow.Internal.PropertyID,
				PropertyName:   apiRow.Internal.PropertyName,
				TotalEmissions: apiRow.TotalEmissions,
				Cached:         false,
			})
		}
	}

	return &models.MeasureResponse{
		RequestID:      requestID,
		TotalEmissions: sumEmissions(modelRows),
		Rows:           modelRows,
	}, nil
}

// sumEmissions aggregates the total emissions from all rows.
func sumEmissions(rows []models.MeasureRowResponse) float64 {
	total := 0.0
	for _, row := range rows {
		total += row.TotalEmissions
	}
	return total
}
