package extractor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/battlestag/com-observer/internal/config"
	"github.com/battlestag/com-observer/internal/messages"
)

// Extractor uses an LLM to extract action items from messages
type Extractor struct {
	cfg *config.LLMConfig
}

// NewExtractor creates a new task extractor
func NewExtractor(cfg *config.LLMConfig) *Extractor {
	return &Extractor{cfg: cfg}
}

// Extract processes a batch of messages and returns extracted tasks
func (e *Extractor) Extract(userName string, msgs []messages.Message) ([]messages.Task, error) {
	if len(msgs) == 0 {
		return nil, nil
	}

	// Build prompt
	prompt := e.buildPrompt(userName, msgs)

	// Call LLM based on provider
	var tasks []messages.Task
	var err error

	switch e.cfg.Provider {
	case "anthropic":
		tasks, err = e.callAnthropicAPI(prompt)
	case "openai":
		tasks, err = e.callOpenAIAPI(prompt)
	case "custom":
		tasks, err = e.callCustomAPI(prompt)
	default:
		return nil, fmt.Errorf("unsupported LLM provider: %s", e.cfg.Provider)
	}

	if err != nil {
		return nil, err
	}

	// Add metadata from original messages to tasks
	tasks = e.enrichTasks(tasks, msgs)

	return tasks, nil
}

// buildPrompt constructs the prompt for the LLM
func (e *Extractor) buildPrompt(userName string, msgs []messages.Message) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`You are analyzing messages for %s to extract action items.

Your task is to identify any actions, tasks, or commitments mentioned in these messages.

Rules:
- Extract ALL actionable items mentioned in the messages
- For each task, identify WHO is being asked to do it (the assignee)
- If %s is the one being asked, set assignee to "%s"
- If someone else is being asked, set assignee to their name/username
- If unclear who should do it, set assignee to "unassigned"
- Ignore completed tasks or things already done
- Include enough context to understand what needs to be done
- Be specific but concise

Return your response as a JSON array of tasks with this format:
[
  {
    "description": "Brief description of what needs to be done",
    "context": "Additional context from the conversation",
    "assignee": "Name of person who should do this task"
  }
]

If there are no action items, return an empty array: []

Messages to analyze:

`, userName, userName, userName))

	// Add messages
	for i, msg := range msgs {
		sb.WriteString(fmt.Sprintf("---\nMessage %d:\n", i+1))
		sb.WriteString(fmt.Sprintf("From: %s\n", msg.Author))
		sb.WriteString(fmt.Sprintf("Channel: %s (%s)\n", msg.Channel, msg.Source))
		sb.WriteString(fmt.Sprintf("Time: %s\n", msg.Timestamp.Format("2006-01-02 15:04:05")))
		sb.WriteString(fmt.Sprintf("Content:\n%s\n\n", msg.Content))
	}

	return sb.String()
}

// callAnthropicAPI calls the Anthropic Claude API
func (e *Extractor) callAnthropicAPI(prompt string) ([]messages.Task, error) {
	url := "https://api.anthropic.com/v1/messages"

	requestBody := map[string]interface{}{
		"model":      e.cfg.Model,
		"max_tokens": 4096,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	reqJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", e.cfg.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Content) == 0 {
		return nil, fmt.Errorf("empty response from API")
	}

	// Parse the JSON array from the response text
	return e.parseTasksJSON(response.Content[0].Text)
}

// callOpenAIAPI calls the OpenAI API or OpenAI-compatible endpoint (like Ollama)
func (e *Extractor) callOpenAIAPI(prompt string) ([]messages.Task, error) {
	// Use custom endpoint if provided, otherwise use OpenAI default
	url := "https://api.openai.com/v1/chat/completions"
	if e.cfg.Endpoint != "" {
		url = e.cfg.Endpoint
	}

	requestBody := map[string]interface{}{
		"model": e.cfg.Model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	reqJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// Only set Authorization header if API key is provided (Ollama doesn't require it)
	if e.cfg.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", e.cfg.APIKey))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("empty response from API")
	}

	// Parse the JSON array from the response text
	return e.parseTasksJSON(response.Choices[0].Message.Content)
}

// callCustomAPI calls a custom LLM endpoint
func (e *Extractor) callCustomAPI(prompt string) ([]messages.Task, error) {
	if e.cfg.Endpoint == "" {
		return nil, fmt.Errorf("custom endpoint not configured")
	}

	// Simple POST with prompt
	requestBody := map[string]string{
		"prompt": prompt,
	}

	reqJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", e.cfg.Endpoint, bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if e.cfg.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", e.cfg.APIKey))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response struct {
		Response string `json:"response"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Parse the JSON array from the response text
	return e.parseTasksJSON(response.Response)
}

// parseTasksJSON extracts the JSON array from the LLM response
func (e *Extractor) parseTasksJSON(text string) ([]messages.Task, error) {
	// Strip markdown code blocks if present (```json ... ``` or ``` ... ```)
	text = strings.ReplaceAll(text, "```json", "")
	text = strings.ReplaceAll(text, "```", "")
	text = strings.TrimSpace(text)

	// Find JSON array in the text (LLMs sometimes add extra text)
	start := strings.Index(text, "[")
	end := strings.LastIndex(text, "]")

	if start == -1 || end == -1 || start >= end {
		// No valid JSON array found, return empty list
		return []messages.Task{}, nil
	}

	jsonStr := text[start : end+1]

	// Parse JSON
	var rawTasks []struct {
		Description string `json:"description"`
		Context     string `json:"context"`
		Assignee    string `json:"assignee"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &rawTasks); err != nil {
		return nil, fmt.Errorf("failed to parse tasks JSON: %w", err)
	}

	// Convert to Task objects (metadata will be added by enrichTasks)
	tasks := make([]messages.Task, len(rawTasks))
	for i, rt := range rawTasks {
		tasks[i] = messages.Task{
			Description: rt.Description,
			Context:     rt.Context,
			Assignee:    rt.Assignee,
		}
	}

	return tasks, nil
}

// enrichTasks adds metadata from original messages to extracted tasks
func (e *Extractor) enrichTasks(tasks []messages.Task, msgs []messages.Message) []messages.Task {
	// For simplicity, use the first message's metadata for all tasks
	// In a more sophisticated implementation, you could try to match each task to its source message
	if len(msgs) == 0 {
		return tasks
	}

	firstMsg := msgs[0]
	for i := range tasks {
		tasks[i].Source = firstMsg.Source
		tasks[i].Channel = firstMsg.Channel
		tasks[i].Author = firstMsg.Author
		tasks[i].Timestamp = firstMsg.Timestamp
		tasks[i].Permalink = firstMsg.Permalink
	}

	return tasks
}
