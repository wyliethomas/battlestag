package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"battlestag/client"
)

// ChatMessage represents a single message in the conversation
type ChatMessage struct {
	Role      string    // "user" or "assistant"
	Content   string    // Message content
	Timestamp time.Time // When the message was sent
	Commands  []string  // Suggested commands (for assistant messages)
}

// ChatModel represents the chat conversation view
type ChatModel struct {
	messages     []ChatMessage
	width        int
	height       int
	loading      bool
	err          error
	apiClient    *client.Client
	scrollOffset int // Scroll position (lines from bottom)
}

// NewChatModel creates a new chat model
func NewChatModel(apiClient *client.Client) *ChatModel {
	return &ChatModel{
		messages:  []ChatMessage{},
		apiClient: apiClient,
	}
}

// AddUserMessage adds a user message to the conversation
func (m *ChatModel) AddUserMessage(content string) {
	m.messages = append(m.messages, ChatMessage{
		Role:      "user",
		Content:   content,
		Timestamp: time.Now(),
	})
	// Auto-scroll to bottom when new message is added
	m.scrollOffset = 0
}

// AddAssistantMessage adds an assistant message to the conversation
func (m *ChatModel) AddAssistantMessage(content string, commands []string) {
	m.messages = append(m.messages, ChatMessage{
		Role:      "assistant",
		Content:   content,
		Timestamp: time.Now(),
		Commands:  commands,
	})
	// Auto-scroll to bottom when new message is added
	m.scrollOffset = 0
}

// AddErrorMessage adds an error message to the conversation
func (m *ChatModel) AddErrorMessage(err error) {
	m.messages = append(m.messages, ChatMessage{
		Role:      "assistant",
		Content:   fmt.Sprintf("Error: %v", err),
		Timestamp: time.Now(),
	})
	// Auto-scroll to bottom when new message is added
	m.scrollOffset = 0
}

// SetLoading sets the loading state
func (m *ChatModel) SetLoading(loading bool) {
	m.loading = loading
}

// SetError sets the error state
func (m *ChatModel) SetError(err error) {
	m.err = err
}

// SetSize updates the chat view size
func (m *ChatModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// GetHistory returns the conversation history for API requests
func (m *ChatModel) GetHistory() []client.ChatHistory {
	history := make([]client.ChatHistory, 0, len(m.messages))
	for _, msg := range m.messages {
		history = append(history, client.ChatHistory{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}
	return history
}

// Clear clears the conversation history
func (m *ChatModel) Clear() {
	m.messages = []ChatMessage{}
	m.err = nil
	m.scrollOffset = 0
}

// ScrollUp scrolls the chat view up
func (m *ChatModel) ScrollUp(lines int) {
	m.scrollOffset += lines
	// Cap at reasonable maximum based on content
	maxScroll := m.calculateMaxScroll()
	if m.scrollOffset > maxScroll {
		m.scrollOffset = maxScroll
	}
}

// ScrollDown scrolls the chat view down
func (m *ChatModel) ScrollDown(lines int) {
	m.scrollOffset -= lines
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
}

// ScrollToBottom scrolls to the bottom of the chat
func (m *ChatModel) ScrollToBottom() {
	m.scrollOffset = 0
}

// calculateMaxScroll calculates the maximum scroll offset based on content
func (m *ChatModel) calculateMaxScroll() int {
	// Count total lines in all messages
	totalLines := 0
	for _, msg := range m.messages {
		// Each message has header (1 line) + content lines + spacing
		contentLines := len(strings.Split(msg.Content, "\n"))
		totalLines += 1 + contentLines + 1 // header + content + spacing

		// Add command suggestions if present
		if len(msg.Commands) > 0 {
			totalLines += 1
		}
	}

	// Max scroll is total lines minus viewport height
	maxScroll := totalLines - m.height
	if maxScroll < 0 {
		maxScroll = 0
	}
	return maxScroll
}

// View renders the chat conversation
func (m *ChatModel) View() string {
	if len(m.messages) == 0 && !m.loading {
		return m.renderWelcome()
	}

	var renderedMessages []string

	// Render each message
	for _, msg := range m.messages {
		renderedMessages = append(renderedMessages, m.renderMessage(msg))
	}

	// Show loading indicator if waiting for response
	if m.loading {
		renderedMessages = append(renderedMessages, m.renderLoading())
	}

	// Join all messages
	content := lipgloss.JoinVertical(lipgloss.Left, renderedMessages...)

	// Apply viewport limiting based on height
	content = m.applyViewport(content)

	return content
}

// applyViewport limits content to viewport height and handles scrolling
func (m *ChatModel) applyViewport(content string) string {
	if m.height <= 0 {
		return content
	}

	lines := strings.Split(content, "\n")
	totalLines := len(lines)

	// If content fits in viewport, show it all
	if totalLines <= m.height {
		return content
	}

	// Calculate visible range with scroll offset
	endIdx := totalLines - m.scrollOffset
	startIdx := endIdx - m.height
	if startIdx < 0 {
		startIdx = 0
	}
	if endIdx > totalLines {
		endIdx = totalLines
	}

	// Extract visible lines
	visibleLines := lines[startIdx:endIdx]

	// Add scroll indicators
	var result []string

	// Top indicator if scrolled up
	if startIdx > 0 {
		result = append(result, scrollIndicatorStyle.Render("▲ Scrolled up - more messages above"))
	}

	result = append(result, visibleLines...)

	// Bottom indicator if not at bottom
	if endIdx < totalLines {
		result = append(result, scrollIndicatorStyle.Render("▼ More messages below - scroll down"))
	}

	return strings.Join(result, "\n")
}

// renderWelcome renders the welcome message
func (m *ChatModel) renderWelcome() string {
	welcome := `
░█▀▄░█▀█░▀█▀░▀█▀░█░░░█▀▀░█▀▀░▀█▀░█▀█░█▀▀
░█▀▄░█▀█░░█░░░█░░█░░░█▀▀░▀▀█░░█░░█▀█░█░█
░▀▀░░▀░▀░░▀░░░▀░░▀▀▀░▀▀▀░▀▀▀░░▀░░▀░▀░▀▀▀
`
	return chatWelcomeStyle.Render(welcome)
}

// renderMessage renders a single chat message
func (m *ChatModel) renderMessage(msg ChatMessage) string {
	// Format timestamp
	timeStr := msg.Timestamp.Format("15:04")

	var content string

	if msg.Role == "user" {
		// User message - right aligned, blue
		header := userHeaderStyle.Render(fmt.Sprintf("You • %s", timeStr))
		body := userMessageStyle.Width(m.width - 10).Render(msg.Content)
		content = lipgloss.JoinVertical(lipgloss.Right, header, body)
	} else {
		// Assistant message - left aligned, green
		header := assistantHeaderStyle.Render(fmt.Sprintf("Assistant • %s", timeStr))
		body := assistantMessageStyle.Width(m.width - 10).Render(msg.Content)

		// Add suggested commands if available
		var commandsSection string
		if len(msg.Commands) > 0 {
			commandsList := make([]string, len(msg.Commands))
			for i, cmd := range msg.Commands {
				commandsList[i] = suggestedCommandStyle.Render(cmd)
			}
			commandsText := "Suggested: " + strings.Join(commandsList, " ")
			commandsSection = "\n" + commandsHintStyle.Render(commandsText)
		}

		content = lipgloss.JoinVertical(lipgloss.Left, header, body, commandsSection)
	}

	// Add spacing between messages
	return content + "\n"
}

// renderLoading renders the loading indicator
func (m *ChatModel) renderLoading() string {
	return loadingStyle.Render("● Thinking...")
}

// Chat message styles
var (
	// User message styles
	userHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")). // Bright blue
			Bold(true).
			Align(lipgloss.Right)

	userMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")). // White
				Background(lipgloss.Color("24")). // Dark blue
				Padding(0, 1).
				Align(lipgloss.Right)

	// Assistant message styles
	assistantHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")). // Bright green
				Bold(true)

	assistantMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")). // White
				Background(lipgloss.Color("22")). // Dark green
				Padding(0, 1)

	// Suggested commands styles
	commandsHintStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("242")). // Gray
				Italic(true)

	suggestedCommandStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("14")). // Cyan
				Bold(true)

	// Loading indicator style
	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")). // Yellow
			Italic(true)

	// Welcome message style
	chatWelcomeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("242")). // Gray
				Padding(1, 2)

	// Scroll indicator style
	scrollIndicatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("242")). // Gray
				Italic(true).
				Align(lipgloss.Center)
)
