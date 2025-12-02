package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/battlestag/com-observer/internal/config"
	"github.com/battlestag/com-observer/internal/messages"
	"github.com/battlestag/com-observer/internal/state"
)

// TrelloAction represents a Trello action (comment, card update, etc.)
type TrelloAction struct {
	ID              string    `json:"id"`
	Type            string    `json:"type"`
	Date            time.Time `json:"date"`
	MemberCreator   Member    `json:"memberCreator"`
	Data            ActionData `json:"data"`
}

type Member struct {
	FullName string `json:"fullName"`
	Username string `json:"username"`
}

type ActionData struct {
	Text  string `json:"text"`
	Card  Card   `json:"card"`
	Board Board  `json:"board"`
}

type Card struct {
	Name string `json:"name"`
}

type Board struct {
	Name string `json:"name"`
}

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "Path to configuration file")
	stateDir := flag.String("state-dir", "/tmp/com-observer/state", "Path to state directory")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate Trello configuration
	if cfg.Trello.APIKey == "" || cfg.Trello.Token == "" {
		log.Fatal("Trello API key and token are required")
	}

	// Initialize state manager
	stateManager, err := state.NewManager(*stateDir)
	if err != nil {
		log.Fatalf("Failed to create state manager: %v", err)
	}

	// Load Trello state
	trelloState, err := stateManager.Load("trello")
	if err != nil {
		log.Fatalf("Failed to load Trello state: %v", err)
	}

	// Initialize message queue
	queue, err := messages.NewQueue(cfg.Queue.Path)
	if err != nil {
		log.Fatalf("Failed to create message queue: %v", err)
	}

	log.Println("Connected to Trello API")

	// Sync each board
	messageCount := 0
	for _, boardID := range cfg.Trello.Boards {
		count, err := syncBoard(cfg, queue, trelloState, boardID)
		if err != nil {
			log.Printf("Error syncing board %s: %v", boardID, err)
			continue
		}
		messageCount += count
	}

	// Save updated state
	if err := stateManager.Save(trelloState); err != nil {
		log.Fatalf("Failed to save state: %v", err)
	}

	log.Printf("Sync complete: %d messages queued", messageCount)
}

// syncBoard syncs actions from a Trello board
func syncBoard(cfg *config.Config, queue *messages.Queue, st *state.State, boardID string) (int, error) {
	lastSync := st.GetChannelLastSync(boardID)

	// If this is the first sync, only get actions from the last day
	if lastSync.IsZero() {
		lastSync = time.Now().Add(-24 * time.Hour)
	}

	// Get board info
	boardName, err := getBoardName(cfg, boardID)
	if err != nil {
		return 0, fmt.Errorf("failed to get board name: %w", err)
	}

	log.Printf("Syncing board: %s (%s)", boardName, boardID)

	// Get actions since last sync
	actions, err := getBoardActions(cfg, boardID, lastSync)
	if err != nil {
		return 0, err
	}

	log.Printf("Found %d actions in board %s", len(actions), boardName)

	messageCount := 0
	var latestTimestamp time.Time

	// Process actions
	for _, action := range actions {
		// Only process comment actions
		if action.Type != "commentCard" {
			continue
		}

		// Track latest timestamp
		if action.Date.After(latestTimestamp) {
			latestTimestamp = action.Date
		}

		// Convert to standard message format
		stdMsg := convertTrelloAction(action, boardName, boardID)

		// Push to queue
		if err := queue.Push(stdMsg); err != nil {
			return messageCount, err
		}
		messageCount++
	}

	// Update state
	if !latestTimestamp.IsZero() {
		st.UpdateChannelSync(boardID, latestTimestamp, "")
	}

	return messageCount, nil
}

// getBoardName fetches the board name from Trello
func getBoardName(cfg *config.Config, boardID string) (string, error) {
	apiURL := fmt.Sprintf("https://api.trello.com/1/boards/%s?key=%s&token=%s",
		boardID, cfg.Trello.APIKey, cfg.Trello.Token)

	resp, err := http.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var board Board
	if err := json.NewDecoder(resp.Body).Decode(&board); err != nil {
		return "", err
	}

	return board.Name, nil
}

// getBoardActions fetches actions from a Trello board since a given time
func getBoardActions(cfg *config.Config, boardID string, since time.Time) ([]TrelloAction, error) {
	// Build URL with parameters
	params := url.Values{}
	params.Set("key", cfg.Trello.APIKey)
	params.Set("token", cfg.Trello.Token)
	params.Set("since", since.Format(time.RFC3339))
	params.Set("filter", "commentCard")
	params.Set("limit", "1000")

	apiURL := fmt.Sprintf("https://api.trello.com/1/boards/%s/actions?%s",
		boardID, params.Encode())

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var actions []TrelloAction
	if err := json.NewDecoder(resp.Body).Decode(&actions); err != nil {
		return nil, err
	}

	return actions, nil
}

// convertTrelloAction converts a Trello action to standard message format
func convertTrelloAction(action TrelloAction, boardName, boardID string) messages.Message {
	// Build card name/context
	cardContext := action.Data.Card.Name
	if cardContext == "" {
		cardContext = "Unknown card"
	}

	// Build permalink
	permalink := fmt.Sprintf("https://trello.com/c/%s", action.Data.Card.Name)

	return messages.Message{
		ID:        action.ID,
		Source:    "trello",
		Channel:   boardName,
		Author:    action.MemberCreator.FullName,
		Content:   fmt.Sprintf("Card: %s\n\n%s", cardContext, action.Data.Text),
		Timestamp: action.Date,
		Permalink: permalink,
		Metadata: map[string]string{
			"board_id": boardID,
			"card_name": cardContext,
			"username": action.MemberCreator.Username,
		},
	}
}
