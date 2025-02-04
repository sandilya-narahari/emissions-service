package server

import (
	"fmt"
	"net/http"

	"emissions-cache-service/internal/handler"
	"emissions-cache-service/internal/service"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// requestIDMiddleware injects a unique request ID into each request for traceability.
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := uuid.New().String()
		r.Header.Set("X-Request-ID", reqID)
		w.Header().Set("X-Request-ID", reqID)
		next.ServeHTTP(w, r)
	})
}

// recoveryMiddleware recovers from panics and returns a 500 error.
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// HTTPServer wraps the http.Server.
type HTTPServer struct {
	*http.Server
}

// NewHTTPServer creates a new HTTP server with routes and middleware applied.
func NewHTTPServer(service service.MeasureService, host string, port int) *HTTPServer {
	r := mux.NewRouter()

	// Initialize handlers.
	measureHandler := handler.NewMeasureHandler(service)

	// Register routes.
	r.HandleFunc("/v1/emissions/measure", measureHandler.Measure).Methods("POST")
	r.HandleFunc("/v1/health", measureHandler.HealthCheck).Methods("GET")

	// Apply middleware.
	r.Use(requestIDMiddleware)
	r.Use(recoveryMiddleware)

	addr := fmt.Sprintf("%s:%d", host, port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	return &HTTPServer{srv}
}
