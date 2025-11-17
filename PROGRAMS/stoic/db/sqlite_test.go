package db

import (
	"os"
	"testing"
	"time"
)

func TestNewDB(t *testing.T) {
	// Use temporary file for testing
	dbPath := "./test_thoughts.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if db.conn == nil {
		t.Fatal("Database connection is nil")
	}
}

func TestSaveAndGetThought(t *testing.T) {
	dbPath := "./test_thoughts.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	date := time.Now().Format("2006-01-02")
	thoughtText := "You have power over your mind, not outside events."

	// Save thought
	err = db.SaveThought(date, thoughtText)
	if err != nil {
		t.Fatalf("Failed to save thought: %v", err)
	}

	// Retrieve thought
	thought, err := db.GetThoughtByDate(date)
	if err != nil {
		t.Fatalf("Failed to get thought: %v", err)
	}

	if thought == nil {
		t.Fatal("Expected thought but got nil")
	}

	if thought.Date != date {
		t.Errorf("Expected date %s, got %s", date, thought.Date)
	}

	if thought.Thought != thoughtText {
		t.Errorf("Expected thought %s, got %s", thoughtText, thought.Thought)
	}
}

func TestGetNonExistentThought(t *testing.T) {
	dbPath := "./test_thoughts.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Try to get thought for a date that doesn't exist
	thought, err := db.GetThoughtByDate("2000-01-01")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if thought != nil {
		t.Error("Expected nil thought for non-existent date")
	}
}

func TestDeleteThought(t *testing.T) {
	dbPath := "./test_thoughts.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	date := "2025-11-16"
	thoughtText := "Test thought"

	// Save thought
	err = db.SaveThought(date, thoughtText)
	if err != nil {
		t.Fatalf("Failed to save thought: %v", err)
	}

	// Verify it exists
	thought, err := db.GetThoughtByDate(date)
	if err != nil {
		t.Fatalf("Failed to get thought: %v", err)
	}
	if thought == nil {
		t.Fatal("Expected thought to exist")
	}

	// Delete thought
	err = db.DeleteThoughtByDate(date)
	if err != nil {
		t.Fatalf("Failed to delete thought: %v", err)
	}

	// Verify it's gone
	thought, err = db.GetThoughtByDate(date)
	if err != nil {
		t.Fatalf("Unexpected error after deletion: %v", err)
	}
	if thought != nil {
		t.Error("Expected thought to be deleted")
	}
}

func TestDateFormatting(t *testing.T) {
	now := time.Now()
	formatted := now.Format("2006-01-02")

	// Parse it back
	parsed, err := time.Parse("2006-01-02", formatted)
	if err != nil {
		t.Fatalf("Failed to parse date: %v", err)
	}

	// Verify year, month, day match
	if parsed.Year() != now.Year() || parsed.Month() != now.Month() || parsed.Day() != now.Day() {
		t.Errorf("Date formatting mismatch: expected %v, got %v", now, parsed)
	}
}
