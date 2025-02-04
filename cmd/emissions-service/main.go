package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"emissions-cache-service/internal/client/scope3"
	"emissions-cache-service/internal/repository/cache"
	"emissions-cache-service/internal/server"
	"emissions-cache-service/internal/service"
	"emissions-cache-service/pkg/config"
)

func main() {
	// Load configuration using Viper.
	cfg, err := config.LoadConfig("/app/config.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Create top-level context for graceful shutdown.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize the cache repository with TTL and cleanup interval.
	cacheTTL, err := cfg.GetCacheTTL()
	if err != nil {
		log.Fatalf("Invalid cache TTL: %v", err)
	}
	cleanupInterval, err := cfg.GetCleanupInterval()
	if err != nil {
		log.Fatalf("Invalid cleanup interval: %v", err)
	}
	emissionsCache := cache.NewInMemoryCache(cacheTTL, cleanupInterval, 0)

	// Initialize the Scope3 client with customizable options.
	scope3Client := scope3.NewClient(
		cfg.Scope3.APIURL,
		cfg.Scope3.Token,
		scope3.WithTimeout(5*time.Second),
	)

	// Initialize the measure service with caching and API client.
	measureService := service.NewMeasureService(emissionsCache, scope3Client)

	// Create and configure the HTTP server.
	srv := server.NewHTTPServer(measureService, cfg.Server.Host, cfg.Server.Port)

	// Start the HTTP server in a separate goroutine.
	go func() {
		log.Printf("Starting server on %s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on %s:%d: %v", cfg.Server.Host, cfg.Server.Port, err)
		}
	}()

	// Listen for termination signals for graceful shutdown.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	log.Println("Shutting down server...")

	// Attempt graceful shutdown.
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server gracefully stopped.")
}
