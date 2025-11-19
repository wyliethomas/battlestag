package app

import (
	"financial-asset-tracker/config"
	"financial-asset-tracker/db"
)

// InitDatabase loads config and opens database connection
// Returns the database connection or an error
// Caller is responsible for calling db.Close()
func InitDatabase() (*db.DB, error) {
	// Load configuration
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
