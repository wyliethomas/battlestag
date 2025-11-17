package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"tech-tip/db"
	"tech-tip/llm"

	"github.com/joho/godotenv"
)

const (
	defaultOllamaHost  = "http://localhost:11434"
	defaultOllamaModel = "dolphin3"
	defaultDBPath      = "./tech_tips.db"
)

func main() {
	// Load .env file if it exists (silently ignore if it doesn't)
	// Environment variables set in the shell take precedence
	_ = godotenv.Load()

	// CLI flags
	regenerate := flag.Bool("regenerate", false, "Force regenerate today's tip")
	dateFlag := flag.String("date", "", "Show tip from specific date (YYYY-MM-DD)")
	randomFlag := flag.Bool("random", false, "Show random tip from history")
	jsonOutput := flag.Bool("json", false, "Output in JSON format")
	flag.Parse()

	// Configuration from environment variables with defaults
	ollamaHost := getEnv("OLLAMA_HOST", defaultOllamaHost)
	ollamaModel := getEnv("OLLAMA_MODEL", defaultOllamaModel)
	dbPath := getEnv("DB_PATH", defaultDBPath)

	// Initialize database
	database, err := db.New(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Handle --random flag
	if *randomFlag {
		if *regenerate || *dateFlag != "" {
			log.Fatalf("Cannot use --random with --regenerate or --date flags")
		}

		tip, err := database.GetRandomTip()
		if err != nil {
			log.Fatalf("Failed to get random tip: %v", err)
		}
		if tip == nil {
			log.Fatalf("No tips in database. Generate your first tip by running without flags.")
		}

		displayTip(tip.Date, tip.Tip, tip.Category, tip.CreatedAt, *jsonOutput)
		return
	}

	// Determine which date to work with
	targetDate := time.Now().Format("2006-01-02")
	if *dateFlag != "" {
		// Validate date format
		_, err := time.Parse("2006-01-02", *dateFlag)
		if err != nil {
			log.Fatalf("Invalid date format. Use YYYY-MM-DD: %v", err)
		}
		targetDate = *dateFlag
	}

	// Handle regenerate flag
	if *regenerate {
		if *dateFlag != "" {
			log.Fatalf("Cannot use --regenerate with --date flag")
		}
		err := database.DeleteTipByDate(targetDate)
		if err != nil {
			log.Fatalf("Failed to delete existing tip: %v", err)
		}
	}

	// Check if tip exists for the target date
	tip, err := database.GetTipByDate(targetDate)
	if err != nil {
		log.Fatalf("Failed to query tip: %v", err)
	}

	// If tip exists and we're not regenerating, display it
	if tip != nil && !*regenerate {
		displayTip(targetDate, tip.Tip, tip.Category, tip.CreatedAt, *jsonOutput)
		return
	}

	// If we're querying a past date and it doesn't exist, show error
	if *dateFlag != "" && tip == nil {
		log.Fatalf("No tip found for date: %s", targetDate)
	}

	// Generate new tip (only for today or with --regenerate)
	if *dateFlag == "" || *regenerate {
		client := llm.NewClient(ollamaHost, ollamaModel)

		// Health check before generating
		if err := client.HealthCheck(); err != nil {
			log.Fatalf("Ollama server health check failed: %v\nMake sure Ollama is running at %s", err, ollamaHost)
		}

		fmt.Println("Generating tech tip...")
		newTip, category, err := client.GenerateTechTip()
		if err != nil {
			log.Fatalf("Failed to generate tip: %v", err)
		}

		// Save to database
		err = database.SaveTip(targetDate, newTip, category)
		if err != nil {
			log.Fatalf("Failed to save tip: %v", err)
		}

		displayTip(targetDate, newTip, category, time.Now(), *jsonOutput)
	}
}

// TipOutput represents the JSON output format
type TipOutput struct {
	Date      string `json:"date"`
	Title     string `json:"title"`
	Tip       string `json:"tip"`
	Category  string `json:"category"`
	CreatedAt string `json:"created_at"`
}

// displayTip formats and displays the tip to the user
func displayTip(date, tip, category string, createdAt time.Time, jsonFormat bool) {
	if jsonFormat {
		output := TipOutput{
			Date:      date,
			Title:     "", // Tech tip doesn't have titles in current schema
			Tip:       tip,
			Category:  category,
			CreatedAt: createdAt.Format(time.RFC3339),
		}
		jsonData, err := json.Marshal(output)
		if err != nil {
			log.Fatalf("Failed to marshal JSON: %v", err)
		}
		fmt.Println(string(jsonData))
	} else {
		fmt.Printf("\n💡 Tech Tip for %s:\n\n", date)
		fmt.Printf("%s\n\n", tip)

		if category != "" {
			fmt.Printf("Category: %s\n", category)
		}

		fmt.Printf("(Generated at %s)\n\n", createdAt.Format("2006-01-02 15:04:05"))
	}
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
