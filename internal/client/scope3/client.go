package scope3

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"emissions-cache-service/internal/errors"
)

// ClientOption defines a functional option for configuring the Scope3 client.
type ClientOption func(*Client)

// Client represents a client for interacting with the Scope3 API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
	userAgent  string
}

// WithTimeout sets a custom timeout for the HTTP client.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithUserAgent sets a custom User-Agent header for the client.
func WithUserAgent(userAgent string) ClientOption {
	return func(c *Client) {
		c.userAgent = userAgent
	}
}

// NewClient initializes a new Scope3 client with the given options.
func NewClient(baseURL, token string, opts ...ClientOption) *Client {
	c := &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 10 * time.Second, // Default timeout.
		},
		userAgent: "EmissionsService/1.0",
	}

	// Apply provided options.
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// GetEmissions makes a POST request to the Scope3 API to retrieve emissions data.
func (c *Client) GetEmissions(ctx context.Context, req MeasureRequest) (*MeasureResponse, error) {
	url := fmt.Sprintf("%s/measure?includeRows=true&latest=true&fields=emissionsBreakdown", c.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, errors.NewInternalError("failed to marshal request", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, errors.NewInternalError("failed to create request", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		// Log detailed error context if needed.
		return nil, errors.NewExternalError("failed to make request to Scope3 API", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.NewExternalError("failed to read response body", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, errors.NewExternalError(
			fmt.Sprintf("Scope3 API error (status: %d)", resp.StatusCode),
			fmt.Errorf("response: %s", string(responseBody)),
		)
	}

	var measureResp MeasureResponse
	if err := json.Unmarshal(responseBody, &measureResp); err != nil {
		return nil, errors.NewInternalError("failed to unmarshal response", err)
	}

	return &measureResp, nil
}
