package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/battlestag/battlestag-tui/client"
)

// UploadModel represents the PDF upload screen
type UploadModel struct {
	client       *client.Client
	filePath     string
	uploading    bool
	uploadStatus string
	err          error
	cursorPos    int
}

// NewUploadModel creates a new upload model
func NewUploadModel(apiClient *client.Client) *UploadModel {
	return &UploadModel{
		client: apiClient,
	}
}

// Update handles upload screen messages
func (m UploadModel) Update(msg tea.Msg) (UploadModel, tea.Cmd) {
	switch msg := msg.(type) {
	case UploadCompleteMsg:
		m.uploading = false
		m.uploadStatus = msg.status
		m.err = nil
		return m, nil

	case ErrorMsg:
		m.err = msg.err
		m.uploading = false
		return m, nil

	case tea.KeyMsg:
		if m.uploading {
			return m, nil
		}

		switch msg.String() {
		case "enter":
			if m.filePath != "" {
				return m, m.uploadFile()
			}
		case "backspace":
			if len(m.filePath) > 0 {
				m.filePath = m.filePath[:len(m.filePath)-1]
			}
		case "ctrl+u":
			m.filePath = ""
		default:
			// Add character to file path
			if len(msg.String()) == 1 {
				m.filePath += msg.String()
			}
		}
	}

	return m, nil
}

// uploadFile uploads the PDF file
func (m *UploadModel) uploadFile() tea.Cmd {
	m.uploading = true
	filePath := m.filePath

	// Expand ~ to home directory
	if strings.HasPrefix(filePath, "~") {
		home, _ := os.UserHomeDir()
		filePath = filepath.Join(home, filePath[1:])
	}

	return func() tea.Msg {
		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return ErrorMsg{err: fmt.Errorf("file does not exist: %s", filePath)}
		}

		// Check if it's a PDF
		if !strings.HasSuffix(strings.ToLower(filePath), ".pdf") {
			return ErrorMsg{err: fmt.Errorf("file must be a PDF")}
		}

		// Upload the file
		err := m.client.UploadStatement(filePath)
		if err != nil {
			return ErrorMsg{err: err}
		}

		return UploadCompleteMsg{
			status: fmt.Sprintf("Successfully uploaded: %s", filepath.Base(filePath)),
		}
	}
}

// View renders the upload screen
func (m *UploadModel) View() string {
	title := titleStyle.Render("📤 Upload Bank Statement")

	var content string

	if m.uploading {
		content = infoStyle.Render("\n\nUploading file, please wait...")
	} else if m.uploadStatus != "" {
		content = successStyle.Render("\n\n"+m.uploadStatus) + "\n\n" +
			helpStyle.Render("Enter a new file path to upload another statement")
	} else if m.err != nil {
		content = errorStyle.Render(fmt.Sprintf("\n\nError: %v", m.err))
	} else {
		content = "\n\nEnter the path to your PDF bank statement:"
	}

	// File path input
	inputLabel := labelStyle.Render("File Path:")
	inputValue := m.filePath
	if !m.uploading {
		inputValue += "█" // Cursor
	}
	input := focusedInputStyle.Render(inputValue)

	fileInput := fmt.Sprintf("\n\n%s\n%s", inputLabel, input)

	// Help text
	help := helpStyle.Render(`
Shortcuts:
  [Enter]  - Upload file
  [Ctrl+U] - Clear input
  [Esc]    - Cancel

Supported file types: PDF
Example: ~/Documents/bank-statement-2024-01.pdf
`)

	fullContent := title + content + fileInput + "\n" + help

	return boxStyle.Render(fullContent)
}

// UploadCompleteMsg indicates upload completion
type UploadCompleteMsg struct {
	status string
}
