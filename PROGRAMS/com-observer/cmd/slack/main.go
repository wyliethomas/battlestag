package main

import (
	"flag"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/battlestag/com-observer/internal/config"
	"github.com/battlestag/com-observer/internal/messages"
	"github.com/battlestag/com-observer/internal/state"
	"github.com/slack-go/slack"
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

	// Validate Slack configuration
	if cfg.Slack.Token == "" {
		log.Fatal("Slack token is not configured")
	}

	// Initialize state manager
	stateManager, err := state.NewManager(*stateDir)
	if err != nil {
		log.Fatalf("Failed to create state manager: %v", err)
	}

	// Load Slack state
	slackState, err := stateManager.Load("slack")
	if err != nil {
		log.Fatalf("Failed to load Slack state: %v", err)
	}

	// Initialize message queue
	queue, err := messages.NewQueue(cfg.Queue.Path)
	if err != nil {
		log.Fatalf("Failed to create message queue: %v", err)
	}

	// Initialize Slack client
	api := slack.New(cfg.Slack.Token)

	// Test authentication
	authTest, err := api.AuthTest()
	if err != nil {
		log.Fatalf("Slack authentication failed: %v", err)
	}
	log.Printf("Connected to Slack as %s (team: %s)", authTest.User, authTest.Team)

	// Get list of channels
	channels, err := getChannels(api, cfg)
	if err != nil {
		log.Fatalf("Failed to get channels: %v", err)
	}

	log.Printf("Monitoring %d channels", len(channels))

	// Sync each channel
	messageCount := 0
	for _, channel := range channels {
		count, err := syncChannel(api, queue, slackState, channel, cfg.User.SlackID)
		if err != nil {
			log.Printf("Error syncing channel %s: %v", channel.Name, err)
			continue
		}
		messageCount += count
	}

	// Sync DMs if enabled
	if cfg.Slack.WatchDMs {
		count, err := syncDMs(api, queue, slackState, cfg.User.SlackID)
		if err != nil {
			log.Printf("Error syncing DMs: %v", err)
		} else {
			messageCount += count
		}
	}

	// Save updated state
	if err := stateManager.Save(slackState); err != nil {
		log.Fatalf("Failed to save state: %v", err)
	}

	log.Printf("Sync complete: %d messages queued", messageCount)
}

// getChannels returns the list of channels to monitor
func getChannels(api *slack.Client, cfg *config.Config) ([]slack.Channel, error) {
	// Get all channels
	allChannels, _, err := api.GetConversations(&slack.GetConversationsParameters{
		Types: []string{"public_channel", "private_channel"},
		Limit: 1000,
	})
	if err != nil {
		return nil, err
	}

	// Filter channels based on include/exclude lists
	var filtered []slack.Channel
	for _, ch := range allChannels {
		// Skip archived channels
		if ch.IsArchived {
			continue
		}

		// Check exclude list
		if contains(cfg.Slack.ChannelsExclude, ch.ID) {
			continue
		}

		// Check include list (empty = include all)
		if len(cfg.Slack.ChannelsInclude) > 0 && !contains(cfg.Slack.ChannelsInclude, ch.ID) {
			continue
		}

		filtered = append(filtered, ch)
	}

	return filtered, nil
}

// syncChannel syncs messages from a single channel
func syncChannel(api *slack.Client, queue *messages.Queue, st *state.State, channel slack.Channel, userID string) (int, error) {
	lastSync := st.GetChannelLastSync(channel.ID)

	// If this is the first sync, only get messages from the last hour
	if lastSync.IsZero() {
		lastSync = time.Now().Add(-1 * time.Hour)
	}

	// Convert to Unix timestamp
	oldest := fmt.Sprintf("%d.000000", lastSync.Unix())

	// Get messages
	params := &slack.GetConversationHistoryParameters{
		ChannelID: channel.ID,
		Oldest:    oldest,
		Limit:     1000,
	}

	history, err := api.GetConversationHistory(params)
	if err != nil {
		return 0, err
	}

	messageCount := 0
	var lastMessageTS string

	// Process messages in reverse order (oldest first)
	for i := len(history.Messages) - 1; i >= 0; i-- {
		msg := history.Messages[i]
		lastMessageTS = msg.Timestamp

		// Skip bot messages only (we now process user's own messages for delegation tracking)
		if msg.BotID != "" {
			continue
		}

		// Get thread replies if this is a thread
		var threadMessages []slack.Message
		if msg.ThreadTimestamp != "" && msg.ThreadTimestamp == msg.Timestamp {
			// This is the parent of a thread, fetch replies
			replies, _, _, err := api.GetConversationReplies(&slack.GetConversationRepliesParameters{
				ChannelID: channel.ID,
				Timestamp: msg.Timestamp,
			})
			if err != nil {
				log.Printf("Failed to get thread replies: %v", err)
			} else {
				threadMessages = replies
			}
		}

		// Convert to standard message format
		stdMsg := convertSlackMessage(api, msg, channel.Name, channel.ID, threadMessages)

		// Push to queue
		if err := queue.Push(stdMsg); err != nil {
			return messageCount, err
		}
		messageCount++
	}

	// Update state
	if lastMessageTS != "" {
		// Parse timestamp to get sync time
		ts, err := parseSlackTimestamp(lastMessageTS)
		if err != nil {
			ts = time.Now()
		}
		st.UpdateChannelSync(channel.ID, ts, lastMessageTS)
	}

	return messageCount, nil
}

// syncDMs syncs direct messages
func syncDMs(api *slack.Client, queue *messages.Queue, st *state.State, userID string) (int, error) {
	// Get DM conversations
	conversations, _, err := api.GetConversations(&slack.GetConversationsParameters{
		Types: []string{"im"},
		Limit: 1000,
	})
	if err != nil {
		return 0, err
	}

	totalCount := 0
	for _, conv := range conversations {
		// Treat each DM as a channel
		count, err := syncChannel(api, queue, st, conv, userID)
		if err != nil {
			log.Printf("Error syncing DM %s: %v", conv.ID, err)
			continue
		}
		totalCount += count
	}

	return totalCount, nil
}

// replaceMentions replaces Slack user mentions (<@USERID>) with actual names
func replaceMentions(api *slack.Client, text string) string {
	// Find all <@USERID> patterns
	re := regexp.MustCompile(`<@([A-Z0-9]+)>`)

	return re.ReplaceAllStringFunc(text, func(match string) string {
		// Extract user ID from <@USERID>
		userID := match[2 : len(match)-1]

		// Look up user name
		user, err := api.GetUserInfo(userID)
		if err != nil {
			return match // Keep original if lookup fails
		}

		// Return @name format
		if user.RealName != "" {
			return "@" + strings.ToLower(strings.Fields(user.RealName)[0])
		} else if user.Name != "" {
			return "@" + user.Name
		}
		return match
	})
}

// convertSlackMessage converts a Slack message to standard format
func convertSlackMessage(api *slack.Client, msg slack.Message, channelName, channelID string, threadMessages []slack.Message) messages.Message {
	// Get author name from user ID
	authorName := msg.User // Fallback to ID if lookup fails
	if msg.User != "" {
		user, err := api.GetUserInfo(msg.User)
		if err != nil {
			log.Printf("Warning: Failed to get user info for %s: %v", msg.User, err)
		} else {
			if user.RealName != "" {
				authorName = user.RealName
			} else if user.Name != "" {
				authorName = user.Name
			}
		}
	}

	// Build content - include thread messages if present
	content := replaceMentions(api, msg.Text) // Replace mentions with names
	if len(threadMessages) > 1 {
		var threadTexts []string
		for _, reply := range threadMessages {
			if reply.User != "" {
				userName := reply.User
				user, err := api.GetUserInfo(reply.User)
				if err == nil {
					if user.RealName != "" {
						userName = user.RealName
					} else if user.Name != "" {
						userName = user.Name
					}
				}
				replyText := replaceMentions(api, reply.Text) // Replace mentions in replies too
				threadTexts = append(threadTexts, fmt.Sprintf("%s: %s", userName, replyText))
			}
		}
		content = strings.Join(threadTexts, "\n")
	}

	// Parse timestamp
	ts, _ := parseSlackTimestamp(msg.Timestamp)

	// Build permalink
	permalink := fmt.Sprintf("https://slack.com/archives/%s/p%s", channelID, strings.Replace(msg.Timestamp, ".", "", 1))

	return messages.Message{
		ID:        msg.Timestamp,
		Source:    "slack",
		Channel:   channelName,
		Author:    authorName,
		Content:   content,
		Timestamp: ts,
		ThreadID:  msg.ThreadTimestamp,
		Permalink: permalink,
		Metadata: map[string]string{
			"channel_id": channelID,
		},
	}
}

// parseSlackTimestamp converts Slack timestamp to time.Time
func parseSlackTimestamp(ts string) (time.Time, error) {
	parts := strings.Split(ts, ".")
	if len(parts) != 2 {
		return time.Time{}, fmt.Errorf("invalid timestamp format")
	}

	var seconds, microseconds int64
	fmt.Sscanf(parts[0], "%d", &seconds)
	fmt.Sscanf(parts[1], "%d", &microseconds)

	return time.Unix(seconds, microseconds*1000), nil
}

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
