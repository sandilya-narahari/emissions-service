package service_test

import (
	"context"
	"testing"

	"emissions-cache-service/internal/client/scope3"
	"emissions-cache-service/internal/models"
	"emissions-cache-service/internal/service"
)

type mockCache struct {
	store map[string]interface{}
}

func (m *mockCache) Set(key string, value interface{}, isPriority bool) {
	m.store[key] = value
}

func (m *mockCache) Get(key string) (interface{}, bool) {
	v, ok := m.store[key]
	return v, ok
}

type mockScope3Client struct {
	response *scope3.MeasureResponse
	err      error
}

func (m *mockScope3Client) GetEmissions(ctx context.Context, req scope3.MeasureRequest) (*scope3.MeasureResponse, error) {
	return m.response, m.err
}

func TestGetMeasureAllCached(t *testing.T) {
	// Prepare a mock cache with two pre-cached responses.
	cacheStore := make(map[string]interface{})
	key1 := "US-online-1000-inv-001"
	key2 := "UK-tv-500-inv-002"
	row1 := scope3.MeasureRowResponse{
		TotalEmissions: 60.0,
		Internal: scope3.InternalData{
			PropertyID:   1,
			PropertyName: "Property 1",
		},
	}
	row2 := scope3.MeasureRowResponse{
		TotalEmissions: 40.0,
		Internal: scope3.InternalData{
			PropertyID:   2,
			PropertyName: "Property 2",
		},
	}
	cacheStore[key1] = row1
	cacheStore[key2] = row2

	mockCacheRepo := &mockCache{store: cacheStore}
	mockScope3 := &mockScope3Client{
		response: nil, // No API call should be made.
		err:      nil,
	}

	svc := service.NewMeasureService(mockCacheRepo, mockScope3)

	req := models.MeasureRequest{
		Rows: []models.MeasureRow{
			{
				Country:     "US",
				Channel:     "online",
				Impressions: 1000,
				InventoryID: "inv-001",
				UTCDatetime: "2025-01-01T12:00:00Z",
				IsPriority:  false,
			},
			{
				Country:     "UK",
				Channel:     "tv",
				Impressions: 500,
				InventoryID: "inv-002",
				UTCDatetime: "2025-01-01T13:00:00Z",
				IsPriority:  true,
			},
		},
	}

	resp, err := svc.GetMeasure(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.RequestID == "" {
		t.Errorf("Expected a non-empty RequestID")
	}
	if resp.TotalEmissions != 100.0 {
		t.Errorf("Expected TotalEmissions 100.0, got %v", resp.TotalEmissions)
	}
	if len(resp.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(resp.Rows))
	}
}

func TestGetMeasurePartialCache(t *testing.T) {
	// Only one row is cached; the other should trigger an API call.
	cacheStore := make(map[string]interface{})
	cacheKey := "US-online-1000-inv-001"
	cachedRow := scope3.MeasureRowResponse{
		TotalEmissions: 60.0,
		Internal: scope3.InternalData{
			PropertyID:   1,
			PropertyName: "Property 1",
		},
	}
	cacheStore[cacheKey] = cachedRow

	mockCacheRepo := &mockCache{store: cacheStore}

	// For the uncached row, simulate an API response.
	apiResponse := &scope3.MeasureResponse{
		RequestID:      "api-req-001",
		TotalEmissions: 40.0,
		Rows: []scope3.MeasureRowResponse{
			{
				TotalEmissions: 40.0,
				Internal: scope3.InternalData{
					PropertyID:   2,
					PropertyName: "Property 2",
				},
			},
		},
	}
	mockScope3 := &mockScope3Client{
		response: apiResponse,
		err:      nil,
	}

	svc := service.NewMeasureService(mockCacheRepo, mockScope3)

	req := models.MeasureRequest{
		Rows: []models.MeasureRow{
			{
				Country:     "US",
				Channel:     "online",
				Impressions: 1000,
				InventoryID: "inv-001",
				UTCDatetime: "2025-01-01T12:00:00Z",
				IsPriority:  false,
			},
			{
				Country:     "UK",
				Channel:     "tv",
				Impressions: 500,
				InventoryID: "inv-002",
				UTCDatetime: "2025-01-01T13:00:00Z",
				IsPriority:  true,
			},
		},
	}

	resp, err := svc.GetMeasure(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.TotalEmissions != 100.0 {
		t.Errorf("Expected TotalEmissions 100.0, got %v", resp.TotalEmissions)
	}
	if resp.RequestID == "" {
		t.Errorf("Expected a non-empty RequestID")
	}
	if len(resp.Rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(resp.Rows))
	}

	// Verify that the API response got cached.
	newKey := "UK-tv-500-inv-002"
	if _, found := mockCacheRepo.Get(newKey); !found {
		t.Errorf("Expected API response to be cached with key %s", newKey)
	}
}

func TestGetMeasureNoRows(t *testing.T) {
	mockCacheRepo := &mockCache{store: make(map[string]interface{})}
	mockScope3 := &mockScope3Client{}
	svc := service.NewMeasureService(mockCacheRepo, mockScope3)

	req := models.MeasureRequest{
		Rows: []models.MeasureRow{},
	}

	_, err := svc.GetMeasure(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for no rows provided, got nil")
	}
}

func TestGetMeasureInvalidRequest(t *testing.T) {
	mockCacheRepo := &mockCache{store: make(map[string]interface{})}
	mockScope3 := &mockScope3Client{}
	svc := service.NewMeasureService(mockCacheRepo, mockScope3)

	req := models.MeasureRequest{} // No rows

	_, err := svc.GetMeasure(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for missing rows, got nil")
	}
}

func TestGetMeasureLargeRequest(t *testing.T) {
	mockCacheRepo := &mockCache{store: make(map[string]interface{})}
	mockScope3 := &mockScope3Client{
		response: &scope3.MeasureResponse{
			RequestID:      "api-req-002",
			TotalEmissions: 5000.0,
			Rows:           make([]scope3.MeasureRowResponse, 100),
		},
		err: nil,
	}

	svc := service.NewMeasureService(mockCacheRepo, mockScope3)

	var rows []models.MeasureRow
	for i := 0; i < 100; i++ {
		rows = append(rows, models.MeasureRow{
			Country: "US", Channel: "online", Impressions: 1000, InventoryID: "inv-001", UTCDatetime: "2025-01-01T12:00:00Z",
		})
	}

	req := models.MeasureRequest{Rows: rows}
	resp, err := svc.GetMeasure(context.Background(), req)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(resp.Rows) != 100 {
		t.Errorf("Expected 100 rows, got %d", len(resp.Rows))
	}
}
