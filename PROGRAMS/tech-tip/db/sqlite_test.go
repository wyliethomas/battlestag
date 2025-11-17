package db

import (
	"os"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	// Create a temporary database file
	dbPath := "./test_tips.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if db.conn == nil {
		t.Error("Database connection is nil")
	}
}

func TestSaveAndGetTip(t *testing.T) {
	dbPath := "./test_tips.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Test data
	date := "2025-11-16"
	tipText := "Use 'git log --oneline' for a compact commit history view."
	category := "git"

	// Save a tip
	err = db.SaveTip(date, tipText, category)
	if err != nil {
		t.Fatalf("Failed to save tip: %v", err)
	}

	// Retrieve the tip
	tip, err := db.GetTipByDate(date)
	if err != nil {
		t.Fatalf("Failed to get tip: %v", err)
	}

	if tip == nil {
		t.Fatal("Expected tip to be found, got nil")
	}

	if tip.Date != date {
		t.Errorf("Expected date %s, got %s", date, tip.Date)
	}

	if tip.Tip != tipText {
		t.Errorf("Expected tip %s, got %s", tipText, tip.Tip)
	}

	if tip.Category != category {
		t.Errorf("Expected category %s, got %s", category, tip.Category)
	}
}

func TestGetTipByDate_NotFound(t *testing.T) {
	dbPath := "./test_tips.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Try to get a tip that doesn't exist
	tip, err := db.GetTipByDate("2025-01-01")
	if err != nil {
		t.Fatalf("Failed to query tip: %v", err)
	}

	if tip != nil {
		t.Error("Expected nil tip for non-existent date")
	}
}

func TestDeleteTipByDate(t *testing.T) {
	dbPath := "./test_tips.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Save a tip
	date := "2025-11-16"
	err = db.SaveTip(date, "Test tip", "test")
	if err != nil {
		t.Fatalf("Failed to save tip: %v", err)
	}

	// Delete the tip
	err = db.DeleteTipByDate(date)
	if err != nil {
		t.Fatalf("Failed to delete tip: %v", err)
	}

	// Verify it's deleted
	tip, err := db.GetTipByDate(date)
	if err != nil {
		t.Fatalf("Failed to query tip: %v", err)
	}

	if tip != nil {
		t.Error("Expected tip to be deleted")
	}
}

func TestGetRandomTip(t *testing.T) {
	dbPath := "./test_tips.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Save multiple tips
	dates := []string{"2025-11-14", "2025-11-15", "2025-11-16"}
	for i, date := range dates {
		err = db.SaveTip(date, "Tip "+string(rune('A'+i)), "test")
		if err != nil {
			t.Fatalf("Failed to save tip: %v", err)
		}
	}

	// Get a random tip
	tip, err := db.GetRandomTip()
	if err != nil {
		t.Fatalf("Failed to get random tip: %v", err)
	}

	if tip == nil {
		t.Fatal("Expected random tip, got nil")
	}

	// Verify the tip is one of the saved ones
	found := false
	for _, date := range dates {
		if tip.Date == date {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Random tip date %s not in saved dates", tip.Date)
	}
}

func TestGetRandomTip_EmptyDatabase(t *testing.T) {
	dbPath := "./test_tips.db"
	defer os.Remove(dbPath)

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Try to get random tip from empty database
	tip, err := db.GetRandomTip()
	if err != nil {
		t.Fatalf("Failed to query random tip: %v", err)
	}

	if tip != nil {
		t.Error("Expected nil tip for empty database")
	}
}

func TestDateFormatting(t *testing.T) {
	// Test date formatting
	now := time.Now()
	dateStr := now.Format("2006-01-02")

	// Parse it back
	parsed, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		t.Fatalf("Failed to parse date: %v", err)
	}

	if parsed.Year() != now.Year() || parsed.Month() != now.Month() || parsed.Day() != now.Day() {
		t.Error("Date formatting/parsing mismatch")
	}
}
