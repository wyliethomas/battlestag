package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"agent-gateway/config"
	"agent-gateway/db"
	"agent-gateway/executor"
	"agent-gateway/handlers"
	"agent-gateway/middleware"

	"github.com/gorilla/mux"
)

var (
	version   = "1.0.0"
	startTime = time.Now()
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Printf("Agent Gateway v%s\n", version)
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup logging
	accessLog, errorLog, accessFile, errorFile := setupLogging(cfg)
	defer accessFile.Close()
	defer errorFile.Close()

	log.Printf("Starting Agent Gateway v%s", version)
	log.Printf("Configuration loaded from: %s", *configPath)

	// Initialize executor for running agent programs
	exec := executor.NewExecutor(
		cfg.Agents.Stoic.ExecutablePath,
		cfg.Agents.Tech.ExecutablePath,
	)
	log.Println("Program executor initialized")

	// Initialize database manager (optional, for advanced features)
	var dbManager *db.Manager
	if cfg.Agents.Stoic.DBPath != "" || cfg.Agents.Tech.DBPath != "" {
		dbManager, err = db.NewManager(cfg)
		if err != nil {
			log.Printf("Warning: Failed to initialize database manager: %v", err)
			log.Println("Advanced features (random, latest, all) may not work")
		} else {
			defer dbManager.Close()
			log.Println("Database connections initialized (for advanced features)")
		}
	}

	// Create handlers
	stoicHandler := handlers.NewStoicHandler(dbManager, exec)
	techHandler := handlers.NewTechHandler(dbManager, exec)
	metaHandler := handlers.NewMetaHandler(dbManager, startTime)

	// Create middleware
	auth := middleware.NewAuthMiddleware(cfg.Auth.APIKey)
	logMiddleware := middleware.Logging(accessLog)

	// Setup router
	router := mux.NewRouter()

	// Apply CORS to all routes
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middleware.CORS(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r)
			})(w, r)
		})
	})

	// Meta endpoints (no auth required for health check)
	router.HandleFunc("/api/health", logMiddleware(metaHandler.GetHealth)).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/stats", logMiddleware(auth.Authenticate(metaHandler.GetStats))).Methods("GET", "OPTIONS")

	// Stoic endpoints (require auth)
	router.HandleFunc("/api/stoic/today", logMiddleware(auth.Authenticate(stoicHandler.GetToday))).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/stoic/date/{date}", logMiddleware(auth.Authenticate(stoicHandler.GetByDate))).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/stoic/random", logMiddleware(auth.Authenticate(stoicHandler.GetRandom))).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/stoic/latest/{n}", logMiddleware(auth.Authenticate(stoicHandler.GetLatest))).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/stoic/all", logMiddleware(auth.Authenticate(stoicHandler.GetAll))).Methods("GET", "OPTIONS")

	// Tech endpoints (require auth)
	router.HandleFunc("/api/tech/today", logMiddleware(auth.Authenticate(techHandler.GetToday))).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/tech/date/{date}", logMiddleware(auth.Authenticate(techHandler.GetByDate))).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/tech/random", logMiddleware(auth.Authenticate(techHandler.GetRandom))).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/tech/latest/{n}", logMiddleware(auth.Authenticate(techHandler.GetLatest))).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/tech/category/{category}", logMiddleware(auth.Authenticate(techHandler.GetByCategory))).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/tech/all", logMiddleware(auth.Authenticate(techHandler.GetAll))).Methods("GET", "OPTIONS")

	// Root endpoint
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"service":"Agent Gateway","version":"%s","status":"running"}`, version)
	}).Methods("GET")

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Handle graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("Shutting down server...")
		if err := server.Close(); err != nil {
			errorLog.Printf("Error during shutdown: %v", err)
		}
	}()

	log.Printf("Server listening on %s", addr)
	log.Println("Press Ctrl+C to stop")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		errorLog.Fatalf("Server failed: %v", err)
	}

	log.Println("Server stopped")
}

// setupLogging configures logging to both stdout and files
func setupLogging(cfg *config.Config) (*log.Logger, *log.Logger, *os.File, *os.File) {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
	}

	// Open access log file
	accessFile, err := os.OpenFile(cfg.Logging.AccessLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open access log: %v", err)
	}

	// Open error log file
	errorFile, err := os.OpenFile(cfg.Logging.ErrorLog, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open error log: %v", err)
	}

	// Create multi-writers (stdout + file)
	accessWriter := io.MultiWriter(os.Stdout, accessFile)
	errorWriter := io.MultiWriter(os.Stderr, errorFile)

	// Create loggers
	accessLog := log.New(accessWriter, "[ACCESS] ", log.LstdFlags)
	errorLog := log.New(errorWriter, "[ERROR] ", log.LstdFlags|log.Lshortfile)

	// Set default logger to error log
	log.SetOutput(errorWriter)
	log.SetFlags(log.LstdFlags)

	return accessLog, errorLog, accessFile, errorFile
}
