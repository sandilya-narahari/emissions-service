package scope3_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"emissions-cache-service/internal/client/scope3"
)

func TestGetEmissionsSuccess(t *testing.T) {
	tests := []struct {
		name     string
		response scope3.MeasureResponse
		wantErr  bool
	}{
		{
			name: "successful response",
			response: scope3.MeasureResponse{
				RequestID:      "req-123",
				TotalEmissions: 100.0,
				Rows: []scope3.MeasureRowResponse{
					{
						TotalEmissions: 50.0,
						Internal: scope3.InternalData{
							PropertyID:   1,
							PropertyName: "Test Property",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing inventory coverage",
			response: scope3.MeasureResponse{
				RequestID:      "req-124",
				TotalEmissions: 0.0,
				Rows: []scope3.MeasureRowResponse{
					{
						InventoryCoverage: "missing",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				if r.Method != http.MethodPost {
					t.Errorf("Expected method POST, got %s", r.Method)
				}
				if auth := r.Header.Get("Authorization"); !strings.HasPrefix(auth, "Bearer ") {
					t.Errorf("Expected Authorization header to start with 'Bearer ', got %s", auth)
				}

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(tt.response)
			}))
			defer ts.Close()

			client := scope3.NewClient(
				ts.URL,
				"dummy-token",
				scope3.WithTimeout(5*time.Second),
				scope3.WithUserAgent("TestClient/1.0"),
			)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			req := scope3.MeasureRequest{
				Rows: []scope3.MeasureRow{
					{
						Country:     "US",
						Channel:     "online",
						Impressions: 1000,
						InventoryID: "inv-001",
						UTCDatetime: time.Now().UTC().Format(time.RFC3339),
					},
				},
				IncludeRows: true,
				Latest:      true,
				Fields:      "emissionsBreakdown",
			}

			resp, err := client.GetEmissions(ctx, req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetEmissions() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if resp.RequestID != tt.response.RequestID {
					t.Errorf("Expected RequestID %s, got %s", tt.response.RequestID, resp.RequestID)
				}
				if len(resp.Rows) != len(tt.response.Rows) {
					t.Errorf("Expected %d rows, got %d", len(tt.response.Rows), len(resp.Rows))
				}
			}
		})
	}
}

func TestGetEmissionsErrors(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   string
		wantErr    bool
	}{
		{
			name:       "internal server error",
			statusCode: http.StatusInternalServerError,
			response:   "Internal Server Error",
			wantErr:    true,
		},
		{
			name:       "rate limit exceeded",
			statusCode: http.StatusTooManyRequests,
			response:   "Rate limit exceeded",
			wantErr:    true,
		},
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			response:   "Invalid token",
			wantErr:    true,
		},
		{
			name:       "malformed response",
			statusCode: http.StatusOK,
			response:   "{invalid json}",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer ts.Close()

			client := scope3.NewClient(ts.URL, "dummy-token")
			ctx := context.Background()

			req := scope3.MeasureRequest{
				Rows: []scope3.MeasureRow{
					{
						Country:     "US",
						Channel:     "online",
						Impressions: 1000,
						InventoryID: "inv-001",
						UTCDatetime: time.Now().UTC().Format(time.RFC3339),
					},
				},
			}

			_, err := client.GetEmissions(ctx, req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEmissions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClientOptions(t *testing.T) {
	tests := []struct {
		name          string
		timeout       time.Duration
		userAgent     string
		expectedError bool
	}{
		{
			name:          "short timeout",
			timeout:       1 * time.Millisecond,
			userAgent:     "TestClient/1.0",
			expectedError: true,
		},
		{
			name:          "custom user agent",
			timeout:       5 * time.Second,
			userAgent:     "CustomAgent/2.0",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a slow server
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectedError {
					time.Sleep(100 * time.Millisecond)
				}
				if ua := r.Header.Get("User-Agent"); ua != tt.userAgent {
					t.Errorf("Expected User-Agent %s, got %s", tt.userAgent, ua)
				}
				json.NewEncoder(w).Encode(scope3.MeasureResponse{})
			}))
			defer ts.Close()

			client := scope3.NewClient(
				ts.URL,
				"dummy-token",
				scope3.WithTimeout(tt.timeout),
				scope3.WithUserAgent(tt.userAgent),
			)

			ctx := context.Background()
			_, err := client.GetEmissions(ctx, scope3.MeasureRequest{})

			if (err != nil) != tt.expectedError {
				t.Errorf("GetEmissions() error = %v, expectedError %v", err, tt.expectedError)
			}
		})
	}
}
