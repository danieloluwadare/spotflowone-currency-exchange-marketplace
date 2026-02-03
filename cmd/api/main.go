package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"spotflowone/internal/external/rates"
	"spotflowone/internal/repo"
	"spotflowone/internal/service"
	httptransport "spotflowone/internal/transport/http"
	"spotflowone/seed"
)

// main initializes the application, wires dependencies, seeds allocations, and starts the HTTP server.
// Sets up signal handling for graceful shutdown, creates services and handlers,
// loads seed data, and listens on the port specified by the PORT environment variable (default 8080).
// The server shuts down gracefully when receiving SIGINT or SIGTERM signals.
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	allocationRepo := repo.NewMemoryAllocationRepository()
	allocationService := service.NewAllocationService(allocationRepo)

	rateProvider := rates.NewCachedProvider(rates.NewFrankfurterClient())
	orderService := service.NewOrderService(allocationRepo, rateProvider)

	handler := httptransport.NewHandler(allocationService, orderService)
	router := httptransport.NewRouter(handler)

	for _, alloc := range seed.Allocations() {
		if err := allocationService.AddAllocation(ctx, alloc); err != nil {
			log.Fatalf("seed allocation failed: %v", err)
		}
	}

	server := &http.Server{
		Addr:              ":" + getEnv("PORT", "8080"),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	log.Printf("api listening on %s", server.Addr)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
}

// getEnv retrieves an environment variable value, returning the fallback if not set or empty.
// Returns the environment variable value if it exists and is non-empty, otherwise returns the fallback.
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
