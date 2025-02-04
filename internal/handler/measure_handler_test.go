package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"emissions-cache-service/internal/handler"
	"emissions-cache-service/internal/models"
)

type dummyMeasureService struct{}

func (d *dummyMeasureService) GetMeasure(ctx context.Context, req models.MeasureRequest) (*models.MeasureResponse, error) {
	return &models.MeasureResponse{
		RequestID:      "unique-dummy-id",
		TotalEmissions: 100.0,
		Rows: []models.MeasureRowResponse{
			{
				PropertyID:     1,
				PropertyName:   "Dummy Property",
				TotalEmissions: 100.0,
				Cached:         false,
			},
		},
	}, nil
}

func TestHealthCheck(t *testing.T) {
	h := handler.NewMeasureHandler(&dummyMeasureService{})
	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	w := httptest.NewRecorder()

	h.HealthCheck(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if body["status"] != "healthy" {
		t.Errorf("Expected status 'healthy', got %s", body["status"])
	}
}

func TestMeasureHandler_Success(t *testing.T) {
	h := handler.NewMeasureHandler(&dummyMeasureService{})
	reqBody := models.MeasureRequest{
		Rows: []models.MeasureRow{
			{
				Country:     "US",
				Channel:     "online",
				Impressions: 1000,
				InventoryID: "inv-001",
				UTCDatetime: "2025-01-01T12:00:00Z",
				IsPriority:  false,
			},
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/v1/emissions/measure", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Measure(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.StatusCode)
	}
	var response models.MeasureResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}
	if response.RequestID != "unique-dummy-id" {
		t.Errorf("Expected RequestID 'unique-dummy-id', got %s", response.RequestID)
	}
}

func TestMeasureHandler_InvalidJSON(t *testing.T) {
	h := handler.NewMeasureHandler(&dummyMeasureService{})
	invalidJSON := []byte("{invalid json}")
	req := httptest.NewRequest(http.MethodPost, "/v1/emissions/measure", bytes.NewReader(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Measure(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status BadRequest for invalid JSON, got %v", resp.StatusCode)
	}
}
