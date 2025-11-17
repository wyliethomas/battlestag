package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"financial-document-watcher/db"
)

const (
	defaultConfigPath = "./watches.json"
	defaultDBPath     = "./watcher.db"
)

// WatchConfig represents a single watch configuration
type WatchConfig struct {
	WatchID        string `json:"watch_id"`
	WatchPath      string `json:"watch_path"`
	FilePattern    string `json:"file_pattern"`
	ExecutablePath string `json:"executable_path"`
	ProcessedPath  string `json:"processed_path"`
}

func main() {
	// CLI flags
	configPath := flag.String("config", defaultConfigPath, "Path to watches.json config file")
	dbPath := flag.String("db", defaultDBPath, "Path to SQLite database")
	dryRun := flag.Bool("dry-run", false, "Show what would be processed without executing")
	flag.Parse()

	log.Printf("Financial Document Watcher started at %s", time.Now().Format(time.RFC3339))

	// Load configuration
	watches, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	log.Printf("Loaded %d watch configurations", len(watches))

	// Initialize database
	database, err := db.New(*dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Process each watch
	totalProcessed := 0
	totalErrors := 0

	for _, watch := range watches {
		processed, errors := processWatch(watch, database, *dryRun)
		totalProcessed += processed
		totalErrors += errors
	}

	log.Printf("Watcher run completed. Processed: %d, Errors: %d", totalProcessed, totalErrors)
}

// loadConfig reads and parses the watches.json configuration file
func loadConfig(path string) ([]WatchConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var watches []WatchConfig
	if err := json.Unmarshal(data, &watches); err != nil {
		return nil, fmt.Errorf("parse config JSON: %w", err)
	}

	// Validate configurations
	for i, w := range watches {
		if w.WatchID == "" {
			return nil, fmt.Errorf("watch %d: watch_id is required", i)
		}
		if w.WatchPath == "" {
			return nil, fmt.Errorf("watch %d: watch_path is required", i)
		}
		if w.FilePattern == "" {
			return nil, fmt.Errorf("watch %d: file_pattern is required", i)
		}
		if w.ExecutablePath == "" {
			return nil, fmt.Errorf("watch %d: executable_path is required", i)
		}
		if w.ProcessedPath == "" {
			return nil, fmt.Errorf("watch %d: processed_path is required", i)
		}
	}

	return watches, nil
}

// processWatch handles a single watch configuration
func processWatch(watch WatchConfig, database *db.DB, dryRun bool) (processed, errors int) {
	log.Printf("[%s] Checking watch path: %s", watch.WatchID, watch.WatchPath)

	// Check if watch path exists
	if _, err := os.Stat(watch.WatchPath); os.IsNotExist(err) {
		log.Printf("[%s] WARNING: Watch path does not exist: %s", watch.WatchID, watch.WatchPath)
		return 0, 0
	}

	// Find matching files
	pattern := filepath.Join(watch.WatchPath, watch.FilePattern)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		log.Printf("[%s] ERROR: Failed to glob pattern %s: %v", watch.WatchID, pattern, err)
		return 0, 1
	}

	if len(matches) == 0 {
		log.Printf("[%s] No files match pattern: %s", watch.WatchID, watch.FilePattern)
		return 0, 0
	}

	log.Printf("[%s] Found %d matching files", watch.WatchID, len(matches))

	// Process each matching file
	for _, filePath := range matches {
		// Check if already processed
		alreadyProcessed, err := database.IsFileProcessed(watch.WatchID, filePath)
		if err != nil {
			log.Printf("[%s] ERROR: Failed to check if file processed: %s: %v", watch.WatchID, filePath, err)
			errors++
			continue
		}

		if alreadyProcessed {
			log.Printf("[%s] SKIP: File already processed: %s", watch.WatchID, filepath.Base(filePath))
			continue
		}

		if dryRun {
			log.Printf("[%s] DRY-RUN: Would process file: %s", watch.WatchID, filePath)
			processed++
			continue
		}

		// Execute the processor
		log.Printf("[%s] Processing file: %s", watch.WatchID, filepath.Base(filePath))
		success, output, exitCode := executeProcessor(watch.ExecutablePath, filePath)

		if success {
			log.Printf("[%s] SUCCESS: Processor completed (exit code: %d)", watch.WatchID, exitCode)
			if output != "" {
				log.Printf("[%s] Output: %s", watch.WatchID, output)
			}

			// Move file to processed directory
			if err := moveToProcessed(filePath, watch.ProcessedPath); err != nil {
				log.Printf("[%s] ERROR: Failed to move file to processed: %v", watch.WatchID, err)
				errors++
				continue
			}

			// Record as processed
			if err := database.RecordProcessedFile(watch.WatchID, filePath); err != nil {
				log.Printf("[%s] ERROR: Failed to record processed file: %v", watch.WatchID, err)
				errors++
				continue
			}

			log.Printf("[%s] File moved to: %s", watch.WatchID, watch.ProcessedPath)
			processed++
		} else {
			log.Printf("[%s] FAILED: Processor failed (exit code: %d)", watch.WatchID, exitCode)
			if output != "" {
				log.Printf("[%s] Error output: %s", watch.WatchID, output)
			}
			log.Printf("[%s] File left in place for retry: %s", watch.WatchID, filePath)
			errors++
		}
	}

	return processed, errors
}

// executeProcessor runs the external executable with the file path as argument
func executeProcessor(executablePath, filePath string) (success bool, output string, exitCode int) {
	cmd := exec.Command(executablePath, filePath)

	// Capture both stdout and stderr
	outputBytes, err := cmd.CombinedOutput()
	output = string(outputBytes)

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return false, output, exitErr.ExitCode()
		}
		// Non-exit error (e.g., executable not found)
		return false, fmt.Sprintf("Failed to execute: %v\n%s", err, output), -1
	}

	return true, output, 0
}

// moveToProcessed moves a file from the watch path to the processed path
func moveToProcessed(filePath, processedPath string) error {
	// Ensure processed directory exists
	if err := os.MkdirAll(processedPath, 0755); err != nil {
		return fmt.Errorf("create processed directory: %w", err)
	}

	// Get the filename
	filename := filepath.Base(filePath)

	// Construct destination path
	destPath := filepath.Join(processedPath, filename)

	// Handle file name collision by adding timestamp
	if _, err := os.Stat(destPath); err == nil {
		timestamp := time.Now().Format("20060102-150405")
		ext := filepath.Ext(filename)
		nameWithoutExt := filename[:len(filename)-len(ext)]
		filename = fmt.Sprintf("%s_%s%s", nameWithoutExt, timestamp, ext)
		destPath = filepath.Join(processedPath, filename)
	}

	// Move the file
	if err := os.Rename(filePath, destPath); err != nil {
		return fmt.Errorf("move file: %w", err)
	}

	return nil
}
