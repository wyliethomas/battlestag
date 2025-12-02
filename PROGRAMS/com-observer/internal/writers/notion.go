package writers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/battlestag/com-observer/internal/config"
	"github.com/battlestag/com-observer/internal/messages"
)

// NotionWriter writes tasks to a Notion database
type NotionWriter struct {
	token      string
	databaseID string
	client     *http.Client
}

// NewNotionWriter creates a new Notion writer
func NewNotionWriter(cfg config.NotionConfig) (*NotionWriter, error) {
	if cfg.Token == "" {
		return nil, fmt.Errorf("notion token is required")
	}
	if cfg.DatabaseID == "" {
		return nil, fmt.Errorf("notion database ID is required")
	}

	return &NotionWriter{
		token:      cfg.Token,
		databaseID: cfg.DatabaseID,
		client:     &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// WriteTask creates a new page in the Notion database for the task
func (w *NotionWriter) WriteTask(task messages.Task) error {
	// Check if task already exists
	if w.TaskExists(task.Permalink) {
		return nil // Skip duplicate
	}

	// Build Notion page properties
	properties := map[string]interface{}{
		"Name": map[string]interface{}{
			"title": []map[string]interface{}{
				{
					"text": map[string]string{
						"content": task.Description,
					},
				},
			},
		},
		"Status": map[string]interface{}{
			"select": map[string]string{
				"name": "To Do",
			},
		},
		"Source": map[string]interface{}{
			"select": map[string]string{
				"name": task.Source,
			},
		},
		"Channel": map[string]interface{}{
			"rich_text": []map[string]interface{}{
				{
					"text": map[string]string{
						"content": task.Channel,
					},
				},
			},
		},
		"Author": map[string]interface{}{
			"rich_text": []map[string]interface{}{
				{
					"text": map[string]string{
						"content": task.Author,
					},
				},
			},
		},
		"Date": map[string]interface{}{
			"date": map[string]string{
				"start": task.Timestamp.Format(time.RFC3339),
			},
		},
		"Link": map[string]interface{}{
			"url": task.Permalink,
		},
	}

	// Build page children (content blocks)
	children := []map[string]interface{}{}
	if task.Context != "" {
		children = append(children, map[string]interface{}{
			"object": "block",
			"type":   "paragraph",
			"paragraph": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{
						"type": "text",
						"text": map[string]string{
							"content": task.Context,
						},
					},
				},
			},
		})
	}

	// Create page request
	requestBody := map[string]interface{}{
		"parent": map[string]string{
			"database_id": w.databaseID,
		},
		"properties": properties,
		"children":   children,
	}

	reqJSON, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make API request
	req, err := http.NewRequest("POST", "https://api.notion.com/v1/pages", bytes.NewBuffer(reqJSON))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", w.token))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call Notion API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Notion API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// TaskExists checks if a task with the given permalink already exists
func (w *NotionWriter) TaskExists(permalink string) bool {
	// Query database for pages with matching permalink
	filter := map[string]interface{}{
		"property": "Link",
		"url": map[string]string{
			"equals": permalink,
		},
	}

	requestBody := map[string]interface{}{
		"filter": filter,
	}

	reqJSON, err := json.Marshal(requestBody)
	if err != nil {
		return false
	}

	// Make API request
	url := fmt.Sprintf("https://api.notion.com/v1/databases/%s/query", w.databaseID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqJSON))
	if err != nil {
		return false
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", w.token))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Notion-Version", "2022-06-28")

	resp, err := w.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false
	}

	var response struct {
		Results []map[string]interface{} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false
	}

	// If we found any results, the task exists
	return len(response.Results) > 0
}
