package main

import (
	"agent-gateway/api"
	"agent-gateway/gateway"
	"agent-gateway/llm"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	timeout, err := config.GetTimeout()
	if err != nil {
		log.Fatalf("Invalid timeout value: %v", err)
	}

	log.Println("=== Agent Gateway v2 ===")
	log.Printf("Starting gateway with:")
	log.Printf("  Programs directory: %s", config.Programs.Directory)
	log.Printf("  Ollama URL: %s", config.Ollama.URL)
	log.Printf("  Ollama Model: %s", config.Ollama.Model)
	log.Printf("  Execution timeout: %s", timeout)

	// Initialize registry
	log.Println("Loading programs...")
	registry, err := gateway.NewRegistry(config.Programs.Directory)
	if err != nil {
		log.Fatalf("Failed to create registry: %v", err)
	}
	log.Printf("✓ Loaded %d programs:", registry.Count())
	for name, program := range registry.GetAllPrograms() {
		log.Printf("  - %s: %s", name, program.Manifest.Description)
	}

	// Initialize executor
	executor := gateway.NewExecutor(timeout)
	log.Printf("✓ Executor initialized")

	// Initialize Ollama client
	llmClient := llm.NewOllamaClient(config.Ollama.URL, config.Ollama.Model)
	log.Printf("✓ Ollama client initialized")

	// Initialize gateway
	gw := gateway.NewGateway(registry, executor, llmClient)
	log.Printf("✓ Gateway initialized")

	// Initialize API handler
	handler := api.NewHandler(gw)

	// Setup HTTP routes
	http.HandleFunc("/chat", handler.HandleChat)
	http.HandleFunc("/health", handler.HandleHealth)
	http.HandleFunc("/programs", handler.HandlePrograms)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", config.Server.Port),
		Handler:      http.DefaultServeMux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("✓ HTTP server listening on :%s", config.Server.Port)
		log.Println("\nAvailable endpoints:")
		log.Println("  POST /chat       - Send chat messages")
		log.Println("  GET  /health     - Health check")
		log.Println("  GET  /programs   - List available programs")
		log.Println("\nGateway is ready! Press Ctrl+C to stop.")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("\nShutting down gracefully...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Gateway stopped.")
}
