# TUI Migration Guide: Gateway v1 → v2

This document outlines the changes needed in `APPS/battlestag-tui` to work with agent-gateway-v2.

## API Differences Summary

### Old Gateway (v1)
- **Base Path**: `/api/*`
- **Authentication**: Required (API key via `X-API-Key` header)
- **Chat Endpoint**: `POST /api/llm/chat`
- **Request Format**:
  ```json
  {
    "message": "user message",
    "history": [
      {"role": "user", "content": "..."},
      {"role": "assistant", "content": "..."}
    ]
  }
  ```
- **Response Format**:
  ```json
  {
    "success": true,
    "data": {
      "message": "assistant response",
      "suggested_commands": ["cmd1", "cmd2"],
      "model": "llama3.2",
      "timestamp": "2025-11-26T...",
      "program_id": "task-manager",
      "program_params": {...}
    }
  }
  ```

### New Gateway (v2)
- **Base Path**: `/`
- **Authentication**: None (open API)
- **Chat Endpoint**: `POST /chat`
- **Request Format**:
  ```json
  {
    "messages": [
      {"role": "user", "content": "user message"},
      {"role": "assistant", "content": "..."}
    ]
  }
  ```
- **Response Format**:
  ```json
  {
    "response": "assistant response",
    "error": "error message (if failed)"
  }
  ```

## Files to Update

### 1. `client/llm.go`

**Current Code:**
```go
// ChatRequest represents the request to the LLM chat endpoint
type ChatRequest struct {
    Message string        `json:"message"`
    History []ChatHistory `json:"history"`
}

func (c *Client) Chat(message string, history []ChatHistory) (*ChatResponse, error) {
    req := ChatRequest{
        Message: message,
        History: history,
    }

    resp, err := c.doRequest("POST", "/api/llm/chat", req)
    // ...
}
```

**New Code:**
```go
// Message represents a chat message for v2 API
type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

// ChatRequestV2 for agent-gateway-v2
type ChatRequestV2 struct {
    Messages []Message `json:"messages"`
}

// ChatResponseV2 from agent-gateway-v2
type ChatResponseV2 struct {
    Response string `json:"response"`
    Error    string `json:"error,omitempty"`
}

func (c *Client) Chat(message string, history []ChatHistory) (*ChatResponse, error) {
    // Convert history + new message to v2 format
    messages := make([]Message, 0, len(history)+1)
    for _, h := range history {
        messages = append(messages, Message{
            Role:    h.Role,
            Content: h.Content,
        })
    }
    messages = append(messages, Message{
        Role:    "user",
        Content: message,
    })

    req := ChatRequestV2{
        Messages: messages,
    }

    resp, err := c.doRequest("POST", "/chat", req)
    if err != nil {
        return nil, err
    }

    // Parse v2 response
    var v2Resp ChatResponseV2
    if err := parseResponse(resp, &v2Resp); err != nil {
        return nil, err
    }

    if v2Resp.Error != "" {
        return nil, fmt.Errorf("gateway error: %s", v2Resp.Error)
    }

    // Convert to old format for backward compatibility
    return &ChatResponse{
        Message:           v2Resp.Response,
        SuggestedCommands: []string{}, // v2 doesn't provide suggestions
        Model:             "",          // v2 doesn't expose model in response
        Timestamp:         time.Now(),
    }, nil
}
```

### 2. `client/api.go`

**Update authentication:**

The v2 gateway doesn't require authentication. Update `doRequest` to skip adding API key header:

```go
func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
    // ... existing code to prepare request ...

    // REMOVE or make conditional:
    // req.Header.Set("X-API-Key", c.APIKey)

    // Or add version detection:
    if c.GatewayVersion == "v1" {
        req.Header.Set("X-API-Key", c.APIKey)
    }
    // v2 doesn't need API key

    return c.HTTPClient.Do(req)
}
```

### 3. Health Check Update

**Current:**
```go
func (c *Client) HealthLLM() (string, error) {
    resp, err := c.doRequest("GET", "/api/llm/health", nil)
    // Expects wrapped response with success field
}
```

**New:**
```go
func (c *Client) HealthLLM() (string, error) {
    resp, err := c.doRequest("GET", "/health", nil)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("health check failed (status %d)", resp.StatusCode)
    }

    var health struct {
        Status         string `json:"status"`
        ProgramsLoaded int    `json:"programs_loaded"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
        return "", err
    }

    return fmt.Sprintf("%s (%d programs loaded)", health.Status, health.ProgramsLoaded), nil
}
```

## Configuration Changes

### Environment Variables

Update default URL in config:

**Old:**
```go
baseURL := os.Getenv("AGENT_GATEWAY_URL")
if baseURL == "" {
    baseURL = "http://192.168.1.140:8080"  // Old v1 gateway
}
```

**New:**
```go
baseURL := os.Getenv("AGENT_GATEWAY_URL")
if baseURL == "" {
    baseURL = "http://localhost:8080"  // v2 gateway default
}

// API key no longer needed for v2
// apiKey := os.Getenv("AGENT_GATEWAY_API_KEY") // Remove or make optional
```

## Testing

After updates, test:

1. **Health Check:**
   ```bash
   curl http://localhost:8080/health
   ```

2. **Chat:**
   ```bash
   curl -X POST http://localhost:8080/chat \
     -H "Content-Type: application/json" \
     -d '{
       "messages": [
         {"role": "user", "content": "Hello!"}
       ]
     }'
   ```

3. **Programs List:**
   ```bash
   curl http://localhost:8080/programs
   ```

## Migration Strategy

### Option 1: Clean Break (Recommended)
Update TUI to only work with v2 gateway. Simplest approach.

### Option 2: Support Both Versions
Add version detection and maintain compatibility:

```go
type Client struct {
    BaseURL        string
    APIKey         string  // Only for v1
    GatewayVersion string  // "v1" or "v2"
    HTTPClient     *http.Client
}

func (c *Client) detectVersion() error {
    // Try v2 health endpoint first
    resp, err := c.HTTPClient.Get(c.BaseURL + "/health")
    if err == nil && resp.StatusCode == 200 {
        c.GatewayVersion = "v2"
        return nil
    }

    // Fall back to v1
    c.GatewayVersion = "v1"
    return nil
}
```

## Breaking Changes

1. ✅ **No more suggested commands** - v2 doesn't provide command suggestions
2. ✅ **No more program metadata** - v2 response is just the message text
3. ✅ **No authentication** - API key no longer needed
4. ✅ **Simpler response format** - Direct response, no nested success/data wrapper

## Summary Checklist

- [ ] Update `ChatRequest` struct to v2 format (`messages` array)
- [ ] Update `Chat()` method to use `/chat` endpoint
- [ ] Update `ChatResponse` handling for simpler v2 format
- [ ] Update `HealthLLM()` to use `/health` endpoint
- [ ] Remove or make API key authentication optional
- [ ] Update default gateway URL configuration
- [ ] Test all chat functionality
- [ ] Update UI to handle missing suggested_commands gracefully
- [ ] Update documentation

## Notes

- The v2 gateway is **simpler** and more **focused** on core LLM chat
- Tool/program execution happens transparently within the gateway
- The TUI just needs to send messages and display responses
- No need to worry about program discovery or parameter validation - gateway handles it
