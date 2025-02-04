package scope3

// MeasureRequest represents the request payload for the Scope3 API.
type MeasureRequest struct {
	Rows        []MeasureRow `json:"rows"`
	IncludeRows bool         `json:"includeRows"`
	Latest      bool         `json:"latest"`
	Fields      string       `json:"fields"`
}

// MeasureRow represents a single row of data in the Scope3 API request.
type MeasureRow struct {
	Country     string `json:"country"`
	Channel     string `json:"channel"`
	Impressions int    `json:"impressions"`
	InventoryID string `json:"inventoryId"`
	UTCDatetime string `json:"utcDatetime"`
}

// MeasureResponse represents the response from the Scope3 API.
type MeasureResponse struct {
	RequestID      string               `json:"requestId"`
	TotalEmissions float64              `json:"totalEmissions"`
	Rows           []MeasureRowResponse `json:"rows"`
}

// MeasureRowResponse represents a single row of data in the Scope3 API response.
type MeasureRowResponse struct {
	TotalEmissions    float64      `json:"totalEmissions"`
	Internal          InternalData `json:"internal"`
	InventoryCoverage string       `json:"inventoryCoverage,omitempty"`
}

// InternalData contains metadata returned by the Scope3 API.
type InternalData struct {
	PropertyID   int    `json:"propertyId"`
	PropertyName string `json:"propertyName"`
}
