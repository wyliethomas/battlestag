package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"agent-gateway/llm"
	"agent-gateway/models"
	"agent-gateway/programs"
)

// LLMHandler handles LLM chat requests
type LLMHandler struct {
	client           *llm.Client
	programsRegistry *programs.Registry
}

// NewLLMHandler creates a new LLM handler
func NewLLMHandler(client *llm.Client) *LLMHandler {
	return &LLMHandler{
		client:           client,
		programsRegistry: nil,
	}
}

// NewLLMHandlerWithPrograms creates a new LLM handler with programs registry
func NewLLMHandlerWithPrograms(client *llm.Client, registry *programs.Registry) *LLMHandler {
	return &LLMHandler{
		client:           client,
		programsRegistry: registry,
	}
}

// Chat handles POST /api/llm/chat
func (h *LLMHandler) Chat(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req models.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		models.WriteError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Validate message
	if req.Message == "" {
		models.WriteError(w, http.StatusBadRequest, "Message cannot be empty")
		return
	}

	// Build programs context if available
	var programsContext string
	if h.programsRegistry != nil {
		programsContext = h.buildProgramsContext()
	}

	// Convert chat history to LLM messages
	history := make([]llm.Message, len(req.History))
	for i, msg := range req.History {
		history[i] = llm.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Prepend programs context to first message if available
	message := req.Message
	if programsContext != "" {
		message = programsContext + "\n\n" + req.Message
	}

	// Send to LLM
	response, err := h.client.Chat(message, history)
	if err != nil {
		models.WriteError(w, http.StatusInternalServerError, "LLM error: "+err.Error())
		return
	}

	// Parse response for program invocation
	programID, programParams, cleanMessage := h.parseProgramInvocation(response)

	// Log if program was detected
	if programID != "" {
		paramsJSON, _ := json.Marshal(programParams)
		fmt.Printf("[LLM] Detected program: %s with params: %s\n", programID, string(paramsJSON))
		fmt.Printf("[LLM] Clean message: %s\n", cleanMessage)
	}

	// Extract suggested commands from response
	suggestedCommands := extractCommands(cleanMessage)

	// Build response
	chatResp := models.ChatResponse{
		Message:           cleanMessage,
		SuggestedCommands: suggestedCommands,
		Model:             h.client.Model,
		Timestamp:         time.Now(),
		ProgramID:         programID,
		ProgramParams:     programParams,
	}

	models.WriteSuccess(w, chatResp)
}

// extractCommands extracts command suggestions from LLM response
// Looks for patterns like /command or mentions of specific commands
func extractCommands(response string) []string {
	commands := []string{}
	seen := make(map[string]bool)

	// Pattern 1: Look for /command patterns
	re := regexp.MustCompile(`/[a-z]+(?:\s+[a-z]+)*`)
	matches := re.FindAllString(response, -1)
	for _, match := range matches {
		cmd := strings.TrimSpace(match)
		if !seen[cmd] {
			commands = append(commands, cmd)
			seen[cmd] = true
		}
	}

	// Pattern 2: Look for common command mentions
	commandKeywords := map[string]string{
		"dashboard":    "/dashboard",
		"assets":       "/assets",
		"liabilities":  "/liabilities",
		"net worth":    "/dashboard",
		"overview":     "/dashboard",
		"upload":       "/upload",
		"settings":     "/settings",
	}

	lowerResponse := strings.ToLower(response)
	for keyword, command := range commandKeywords {
		if strings.Contains(lowerResponse, keyword) && !seen[command] {
			commands = append(commands, command)
			seen[command] = true
		}
	}

	return commands
}

// buildProgramsContext creates a context string describing available programs
func (h *LLMHandler) buildProgramsContext() string {
	if h.programsRegistry == nil {
		return ""
	}

	programs := h.programsRegistry.List()
	if len(programs) == 0 {
		return ""
	}

	context := `You have access to the following programs that can perform tasks:

`
	for _, prog := range programs {
		context += fmt.Sprintf("- %s (%s): %s\n", prog.Name, prog.ID, prog.Description)
		if len(prog.Parameters) > 0 {
			context += "  Parameters:\n"
			for _, param := range prog.Parameters {
				required := ""
				if param.Required {
					required = " [REQUIRED]"
				}
				context += fmt.Sprintf("    • %s (%s)%s: %s\n", param.Name, param.Type, required, param.Description)
			}
		}
		context += "\n"
	}

	context += `When a user's request matches a program's capability, you MUST use this EXACT format to execute it:

EXECUTE_PROGRAM: program_id
PARAMETERS: {"param1": "value1", "param2": "value2"}
---
Your natural language explanation here

IMPORTANT RULES:
1. Use ONLY ONE EXECUTE_PROGRAM block per response
2. ALL required parameters MUST be included in the PARAMETERS line
3. The PARAMETERS must be valid JSON on a SINGLE line
4. Put your explanation AFTER the --- marker
5. Do NOT describe what you're doing before the EXECUTE_PROGRAM block
6. Do NOT say you've completed the task until you get the actual program result
7. Do NOT try to answer the question yourself - let the program provide the data
8. Your message should ONLY explain what you're doing, not provide the answer
9. NEVER show fake "Output:" or fabricate program responses - wait for the actual program
10. ONLY use programs when the user's request specifically needs that program's data
11. For general conversation, brainstorming, or questions - DO NOT use programs, just respond naturally

CRITICAL: When reading program output (especially JSON), present EXACTLY what the program returned.
- Do NOT make up, infer, or fabricate data
- Do NOT change values from what the program reported
- If a server status is "unknown", say "unknown" - do NOT guess "online" or "offline"
- If a field says 0, report 0 - do NOT make up a different number
- Read the actual values carefully and present them accurately

Example (Adding data):
User: "Add a new project called Workshop to house context with goal of organizing tools"
Your response:
EXECUTE_PROGRAM: task-manager
PARAMETERS: {"command": "add-project", "project_name": "Workshop", "context": "house", "goal": "organizing tools"}
---
I'm adding a new project called "Workshop" to your house context.

Example (Querying data):
User: "Show me all projects for property"
Your response:
EXECUTE_PROGRAM: task-manager
PARAMETERS: {"command": "list-projects", "context": "property"}
---
Let me get your property projects from the Task Manager.

DO NOT list the projects yourself - the program will provide the actual data.

Example (General conversation - NO PROGRAM):
User: "Come up with some ideas for a name for a software company"
Your response:
Here are some creative name ideas for a software company:

1. CodeCraft Solutions
2. ByteForge Technologies
3. NexGen Innovations

[Just answer naturally - DO NOT call any programs or show fake output]

For general conversation or questions that don't need a program, just respond normally without the EXECUTE_PROGRAM markers.`

	return context
}

// parseProgramInvocation parses the LLM response to extract program execution details
func (h *LLMHandler) parseProgramInvocation(response string) (string, map[string]interface{}, string) {
	// Look for the pattern: EXECUTE_PROGRAM: program_id\nPARAMETERS: {...}\n---\nmessage
	lines := strings.Split(response, "\n")

	var programID string
	var params map[string]interface{}
	var messageStart int

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for EXECUTE_PROGRAM
		if strings.HasPrefix(trimmed, "EXECUTE_PROGRAM:") {
			programID = strings.TrimSpace(strings.TrimPrefix(trimmed, "EXECUTE_PROGRAM:"))
			continue
		}

		// Check for PARAMETERS
		if strings.HasPrefix(trimmed, "PARAMETERS:") {
			jsonStr := strings.TrimSpace(strings.TrimPrefix(trimmed, "PARAMETERS:"))
			json.Unmarshal([]byte(jsonStr), &params)
			continue
		}

		// Check for separator
		if trimmed == "---" {
			messageStart = i + 1
			break
		}
	}

	// Extract clean message (everything after ---)
	cleanMessage := response
	if messageStart > 0 && messageStart < len(lines) {
		cleanMessage = strings.Join(lines[messageStart:], "\n")
		cleanMessage = strings.TrimSpace(cleanMessage)
	}

	// If no program was detected, return empty program info
	if programID == "" {
		return "", nil, response
	}

	return programID, params, cleanMessage
}

// Health checks if the LLM service is available
func (h *LLMHandler) Health(w http.ResponseWriter, r *http.Request) {
	if err := h.client.Ping(); err != nil {
		models.WriteError(w, http.StatusServiceUnavailable, "LLM service unavailable: "+err.Error())
		return
	}

	models.WriteSuccess(w, map[string]string{
		"status": "healthy",
		"model":  h.client.Model,
	})
}
