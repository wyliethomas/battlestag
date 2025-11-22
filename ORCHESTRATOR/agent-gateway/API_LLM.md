# LLM Chat API

## Overview

The LLM Chat API provides conversational AI capabilities powered by a local language model (via Ollama). The AI assistant is configured to help with financial management tasks and can suggest relevant API commands.

## Configuration

The LLM integration is configured in `config.yaml`:

```yaml
llm:
  endpoint: "http://localhost:11434"  # Ollama server endpoint
  model: "llama3.2:8b"                # Model to use
  timeout: 30                          # Request timeout in seconds
  system_prompt: "You are a helpful financial assistant..."
```

Environment variable overrides:
- `LLM_ENDPOINT` - Override Ollama server endpoint
- `LLM_MODEL` - Override model name
- `LLM_TIMEOUT` - Override timeout (in seconds)
- `LLM_SYSTEM_PROMPT` - Override system prompt

## Endpoints

### Chat with LLM

**Endpoint:** `POST /api/llm/chat`

**Authentication:** Required (X-API-Key header)

**Description:** Send a message to the LLM and receive a response with optional command suggestions.

**Request Body:**
```json
{
  "message": "What's my current net worth?",
  "history": [
    {
      "role": "user",
      "content": "Hello"
    },
    {
      "role": "assistant",
      "content": "Hi! How can I help you with your finances today?"
    }
  ]
}
```

**Request Fields:**
- `message` (string, required): The current user message
- `history` (array, optional): Previous conversation history
  - Each message has `role` ("user" or "assistant") and `content` (string)

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "I can help you check your net worth! Try running the /dashboard command to see your financial overview, or use /assets and /liabilities to see detailed breakdowns.",
    "suggested_commands": [
      "/dashboard",
      "/assets",
      "/liabilities"
    ],
    "model": "llama3.2:8b",
    "timestamp": "2025-11-21T10:30:00Z"
  }
}
```

**Response Fields:**
- `message` (string): The AI assistant's response
- `suggested_commands` (array): Commands the AI suggests based on the conversation
- `model` (string): The model used to generate the response
- `timestamp` (string): When the response was generated (ISO 8601 format)

**Example with curl:**
```bash
curl -X POST http://192.168.1.140:8080/api/llm/chat \
  -H "X-API-Key: test-api-key-12345" \
  -H "Content-Type: application/json" \
  -d '{
    "message": "How can I track my expenses?",
    "history": []
  }'
```

**Error Responses:**

400 Bad Request:
```json
{
  "success": false,
  "error": "Message cannot be empty"
}
```

500 Internal Server Error:
```json
{
  "success": false,
  "error": "LLM error: ollama server unreachable: connection refused"
}
```

### Check LLM Health

**Endpoint:** `GET /api/llm/health`

**Authentication:** Required (X-API-Key header)

**Description:** Check if the LLM service (Ollama) is available and responding.

**Response:**
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "model": "llama3.2:8b"
  }
}
```

**Example with curl:**
```bash
curl http://192.168.1.140:8080/api/llm/health \
  -H "X-API-Key: test-api-key-12345"
```

**Error Response (Service Unavailable):**
```json
{
  "success": false,
  "error": "LLM service unavailable: ollama server unreachable: connection refused"
}
```

## Command Suggestions

The AI assistant can suggest relevant API commands based on user questions. It uses two methods:

1. **Pattern Matching**: Detects `/command` patterns in the AI's response
2. **Keyword Detection**: Maps keywords to relevant commands
   - "dashboard", "overview", "net worth" → `/dashboard`
   - "assets" → `/assets`
   - "liabilities" → `/liabilities`
   - etc.

**Available Commands the AI Knows About:**
- `/dashboard` - View financial overview
- `/assets` - List all assets
- `/liabilities` - List all liabilities
- `/upload` - Upload bank statement
- `/settings` - Configure application
- `/asset add`, `/asset edit`, `/asset delete` - Asset management
- `/liability add`, `/liability edit`, `/liability delete` - Liability management

## Conversation History

The client is responsible for managing conversation history. The API is stateless - each request should include relevant conversation history in the `history` array.

**Best Practices:**
- Keep history to last 5-10 messages for context
- Trim old messages to avoid exceeding token limits
- Include both user and assistant messages in order

**Example Conversation Flow:**

Request 1:
```json
{
  "message": "Hello",
  "history": []
}
```

Request 2 (includes previous exchange):
```json
{
  "message": "What's my net worth?",
  "history": [
    {"role": "user", "content": "Hello"},
    {"role": "assistant", "content": "Hi! How can I help?"}
  ]
}
```

## Error Handling

Common error scenarios:

1. **Ollama Not Running:**
   - Error: "LLM error: ollama server unreachable: connection refused"
   - Solution: Start Ollama service (`ollama serve`)

2. **Model Not Available:**
   - Error: "ollama returned status 404"
   - Solution: Pull the model (`ollama pull llama3.2:8b`)

3. **Timeout:**
   - Error: "LLM error: ... timeout"
   - Solution: Increase `timeout` in config.yaml or use faster model

4. **Invalid Request:**
   - Error: "Invalid request body" or "Message cannot be empty"
   - Solution: Check request JSON format and required fields

## Performance Considerations

- **Response Time:** Depends on model size and hardware (8B model: 2-10 seconds typical)
- **Timeout:** Default 30 seconds, configurable in config.yaml
- **Concurrent Requests:** Ollama handles queueing, but expect delays under load
- **Resource Usage:** 8B model requires ~8GB RAM

## Security

- **API Key Required:** All LLM endpoints require authentication
- **Local Only:** Default configuration uses localhost Ollama server
- **No Data Logging:** Conversations are not stored server-side
- **Client-Side History:** Full conversation history never leaves client

## Integration with TUI

The TUI client uses this API to provide:
- Natural language interface for financial queries
- AI-powered command suggestions
- Contextual help and guidance
- Conversational financial assistant

See the battlestag-tui documentation for details on the chat interface.
