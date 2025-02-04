package handler

import (
	"encoding/json"
	"net/http"

	"emissions-cache-service/internal/errors"
	"emissions-cache-service/internal/models"
	"emissions-cache-service/internal/service"
)

// MeasureHandler handles HTTP requests for emissions measurement.
type MeasureHandler struct {
	measureService service.MeasureService
}

// NewMeasureHandler creates a new MeasureHandler.
func NewMeasureHandler(ms service.MeasureService) *MeasureHandler {
	return &MeasureHandler{measureService: ms}
}

// HealthCheck handles the health check endpoint.
func (h *MeasureHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// Measure handles the emissions measurement endpoint.
func (h *MeasureHandler) Measure(w http.ResponseWriter, r *http.Request) {
	var req models.MeasureRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, errors.NewValidationError("invalid JSON request"))
		return
	}

	response, err := h.measureService.GetMeasure(r.Context(), req)
	if err != nil {
		respondWithError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		respondWithError(w, errors.NewInternalError("failed to encode response", err))
		return
	}
}

// respondWithError sends an error response in JSON format.
// It includes error details for development purposes.
func respondWithError(w http.ResponseWriter, err error) {
	code, message := errors.ToHTTPError(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	response := map[string]string{
		"error": message,
	}

	// Include underlying error detail in development mode.
	if svcErr, ok := err.(*errors.ServiceError); ok && svcErr.Err != nil {
		response["detail"] = svcErr.Err.Error()
	}

	json.NewEncoder(w).Encode(response)
}
