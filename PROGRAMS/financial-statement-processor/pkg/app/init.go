package app

import (
	"financial-statement-processor/config"
	"financial-statement-processor/db"

	"github.com/joho/godotenv"
)

// InitDatabase loads environment variables, config, and opens database connection
// Returns the database connection or an error
// Caller is responsible for calling db.Close()
func InitDatabase() (*db.DB, error) {
	// Load .env file if it exists (silently ignore if it doesn't)
	_ = godotenv.Load()

	// Load database configuration
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return nil, err
	}

	// Connect to database
	database, err := db.New(cfg.DatabasePath())
	if err != nil {
		return nil, err
	}

	return database, nil
}

// InitConfig loads environment variables and config
// Returns the config or an error
func InitConfig() (*config.Config, error) {
	// Load .env file if it exists (silently ignore if it doesn't)
	_ = godotenv.Load()

	// Load configuration
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
