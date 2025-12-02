package main

import (
	"context"
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/battlestag/com-observer/internal/config"
	"github.com/battlestag/com-observer/internal/messages"
	"github.com/battlestag/com-observer/internal/state"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

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

	// Validate Gmail configuration
	if cfg.Gmail.CredentialsFile == "" {
		log.Fatal("Gmail credentials file is not configured")
	}

	// Initialize state manager
	stateManager, err := state.NewManager(*stateDir)
	if err != nil {
		log.Fatalf("Failed to create state manager: %v", err)
	}

	// Load Gmail state
	gmailState, err := stateManager.Load("gmail")
	if err != nil {
		log.Fatalf("Failed to load Gmail state: %v", err)
	}

	// Initialize message queue
	queue, err := messages.NewQueue(cfg.Queue.Path)
	if err != nil {
		log.Fatalf("Failed to create message queue: %v", err)
	}

	// Initialize Gmail service
	ctx := context.Background()
	srv, err := gmail.NewService(ctx, option.WithCredentialsFile(cfg.Gmail.CredentialsFile))
	if err != nil {
		log.Fatalf("Failed to create Gmail client: %v", err)
	}

	// Get user profile
	profile, err := srv.Users.GetProfile("me").Do()
	if err != nil {
		log.Fatalf("Failed to get user profile: %v", err)
	}
	log.Printf("Connected to Gmail as %s", profile.EmailAddress)

	// Sync messages for each label
	messageCount := 0
	for _, label := range cfg.Gmail.Labels {
		count, err := syncLabel(srv, queue, gmailState, label, profile.EmailAddress)
		if err != nil {
			log.Printf("Error syncing label %s: %v", label, err)
			continue
		}
		messageCount += count
	}

	// Save updated state
	if err := stateManager.Save(gmailState); err != nil {
		log.Fatalf("Failed to save state: %v", err)
	}

	log.Printf("Sync complete: %d messages queued", messageCount)
}

// syncLabel syncs messages from a specific Gmail label
func syncLabel(srv *gmail.Service, queue *messages.Queue, st *state.State, label, userEmail string) (int, error) {
	lastSync := st.GetChannelLastSync(label)

	// If this is the first sync, only get messages from the last day
	if lastSync.IsZero() {
		lastSync = time.Now().Add(-24 * time.Hour)
	}

	// Build query
	query := fmt.Sprintf("label:%s after:%s", label, lastSync.Format("2006/01/02"))

	// List messages
	req := srv.Users.Messages.List("me").Q(query).MaxResults(100)
	resp, err := req.Do()
	if err != nil {
		return 0, err
	}

	log.Printf("Found %d messages in label %s", len(resp.Messages), label)

	messageCount := 0
	var latestTimestamp time.Time

	// Fetch full message details for each
	for _, m := range resp.Messages {
		msg, err := srv.Users.Messages.Get("me", m.Id).Format("full").Do()
		if err != nil {
			log.Printf("Failed to get message %s: %v", m.Id, err)
			continue
		}

		// Parse message
		stdMsg, err := parseGmailMessage(msg, label, userEmail)
		if err != nil {
			log.Printf("Failed to parse message %s: %v", m.Id, err)
			continue
		}

		// Skip messages from the user themselves
		if strings.Contains(strings.ToLower(stdMsg.Author), strings.ToLower(userEmail)) {
			continue
		}

		// Track latest timestamp
		if stdMsg.Timestamp.After(latestTimestamp) {
			latestTimestamp = stdMsg.Timestamp
		}

		// Push to queue
		if err := queue.Push(stdMsg); err != nil {
			return messageCount, err
		}
		messageCount++
	}

	// Update state
	if !latestTimestamp.IsZero() {
		st.UpdateChannelSync(label, latestTimestamp, "")
	}

	return messageCount, nil
}

// parseGmailMessage converts a Gmail message to standard format
func parseGmailMessage(msg *gmail.Message, label, userEmail string) (messages.Message, error) {
	// Extract headers
	var subject, from string
	for _, header := range msg.Payload.Headers {
		switch header.Name {
		case "Subject":
			subject = header.Value
		case "From":
			from = header.Value
		}
	}

	// Extract body
	body := extractBody(msg.Payload)

	// Parse timestamp
	timestamp := time.Unix(msg.InternalDate/1000, 0)

	// Build permalink
	permalink := fmt.Sprintf("https://mail.google.com/mail/u/0/#inbox/%s", msg.Id)

	return messages.Message{
		ID:        msg.Id,
		Source:    "gmail",
		Channel:   subject,
		Author:    from,
		Content:   body,
		Timestamp: timestamp,
		Permalink: permalink,
		Metadata: map[string]string{
			"label":    label,
			"thread_id": msg.ThreadId,
		},
	}, nil
}

// extractBody extracts the text body from a Gmail message payload
func extractBody(payload *gmail.MessagePart) string {
	// Try to get plain text body
	if payload.MimeType == "text/plain" && payload.Body.Data != "" {
		decoded, err := base64.URLEncoding.DecodeString(payload.Body.Data)
		if err != nil {
			return ""
		}
		return string(decoded)
	}

	// If multipart, recursively search parts
	if strings.HasPrefix(payload.MimeType, "multipart/") {
		for _, part := range payload.Parts {
			if body := extractBody(part); body != "" {
				return body
			}
		}
	}

	return ""
}
