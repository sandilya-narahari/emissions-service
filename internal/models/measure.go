package models

// MeasureRequest represents the public API request format.
type MeasureRequest struct {
	Rows []MeasureRow `json:"rows"`
}

// MeasureRow represents a single row in the public API request.
type MeasureRow struct {
	Country     string `json:"country"`
	Channel     string `json:"channel"`
	Impressions int    `json:"impressions"`
	InventoryID string `json:"inventoryId"`
	UTCDatetime string `json:"utcDatetime"`
	IsPriority  bool   `json:"isPriority"`
}

// MeasureResponse represents the public API response.
type MeasureResponse struct {
	RequestID      string               `json:"requestId"`
	TotalEmissions float64              `json:"totalEmissions"`
	Rows           []MeasureRowResponse `json:"rows"`
}

// MeasureRowResponse represents a single row in the public API response.
type MeasureRowResponse struct {
	PropertyID        int     `json:"propertyId,omitempty"`
	PropertyName      string  `json:"propertyName,omitempty"`
	TotalEmissions    float64 `json:"totalEmissions,omitempty"`
	Cached            bool    `json:"cached,omitempty"`
	InventoryCoverage string  `json:"inventoryCoverage,omitempty"`
}
