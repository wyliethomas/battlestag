package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"stoic-thought/db"
	"stoic-thought/llm"

	"github.com/joho/godotenv"
)

const (
	defaultOllamaHost  = "http://localhost:11434"
	defaultOllamaModel = "dolphin3"
	defaultDBPath      = "./stoic_thoughts.db"
)

func main() {
	// Load .env file if it exists (silently ignore if it doesn't)
	// Environment variables set in the shell take precedence
	_ = godotenv.Load()

	// CLI flags
	regenerate := flag.Bool("regenerate", false, "Force regenerate today's thought")
	dateFlag := flag.String("date", "", "Show thought from specific date (YYYY-MM-DD)")
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
		err := database.DeleteThoughtByDate(targetDate)
		if err != nil {
			log.Fatalf("Failed to delete existing thought: %v", err)
		}
	}

	// Check if thought exists for the target date
	thought, err := database.GetThoughtByDate(targetDate)
	if err != nil {
		log.Fatalf("Failed to query thought: %v", err)
	}

	// If thought exists and we're not regenerating, display it
	if thought != nil && !*regenerate {
		displayThought(targetDate, thought.Thought, thought.CreatedAt, *jsonOutput)
		return
	}

	// If we're querying a past date and it doesn't exist, show error
	if *dateFlag != "" && thought == nil {
		log.Fatalf("No thought found for date: %s", targetDate)
	}

	// Generate new thought (only for today or with --regenerate)
	if *dateFlag == "" || *regenerate {
		client := llm.NewClient(ollamaHost, ollamaModel)

		// Health check before generating
		if err := client.HealthCheck(); err != nil {
			log.Fatalf("Ollama server health check failed: %v\nMake sure Ollama is running at %s", err, ollamaHost)
		}

		fmt.Println("Generating stoic thought...")
		newThought, err := client.GenerateStoicThought()
		if err != nil {
			log.Fatalf("Failed to generate thought: %v", err)
		}

		// Save to database
		err = database.SaveThought(targetDate, newThought)
		if err != nil {
			log.Fatalf("Failed to save thought: %v", err)
		}

		displayThought(targetDate, newThought, time.Now(), *jsonOutput)
	}
}

// ThoughtOutput represents the JSON output format
type ThoughtOutput struct {
	Date      string `json:"date"`
	Thought   string `json:"thought"`
	CreatedAt string `json:"created_at"`
}

// displayThought formats and displays the thought to the user
func displayThought(date, thought string, createdAt time.Time, jsonFormat bool) {
	if jsonFormat {
		output := ThoughtOutput{
			Date:      date,
			Thought:   thought,
			CreatedAt: createdAt.Format(time.RFC3339),
		}
		jsonData, err := json.Marshal(output)
		if err != nil {
			log.Fatalf("Failed to marshal JSON: %v", err)
		}
		fmt.Println(string(jsonData))
	} else {
		fmt.Printf("\n📜 Stoic Thought for %s:\n\n", date)
		fmt.Printf("\"%s\"\n\n", thought)
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
