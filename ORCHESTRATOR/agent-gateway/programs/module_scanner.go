package programs

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ModuleMetadata holds parsed metadata from a module script
type ModuleMetadata struct {
	ID          string
	Name        string
	Description string
	Category    string
	Parameters  []Parameter
}

// ModuleScanner discovers and loads module scripts
type ModuleScanner struct {
	modulesDir string
}

// NewModuleScanner creates a new module scanner
func NewModuleScanner(modulesDir string) *ModuleScanner {
	return &ModuleScanner{
		modulesDir: modulesDir,
	}
}

// DiscoverModules scans the modules directory and returns ModuleProgram instances
func (s *ModuleScanner) DiscoverModules() ([]*ModuleProgram, error) {
	modules := make([]*ModuleProgram, 0)

	// Walk the modules directory tree
	err := filepath.Walk(s.modulesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process .sh files
		if !strings.HasSuffix(path, ".sh") {
			return nil
		}

		// Skip test files
		if strings.HasPrefix(info.Name(), "test_") {
			return nil
		}

		// Parse module metadata
		metadata, err := s.parseModuleMetadata(path)
		if err != nil {
			// Log warning but continue scanning
			fmt.Fprintf(os.Stderr, "Warning: failed to parse module %s: %v\n", path, err)
			return nil
		}

		// Create ModuleProgram
		module := NewModuleProgram(
			metadata.ID,
			metadata.Name,
			metadata.Description,
			metadata.Category,
			metadata.Parameters,
			path,
		)

		modules = append(modules, module)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan modules directory: %w", err)
	}

	return modules, nil
}

// parseModuleMetadata extracts metadata from module script comments
// Expected format:
// # MODULE: csv_insight
// # NAME: CSV Intelligence
// # CATEGORY: data
// # DESCRIPTION: Analyzes CSV files and provides AI-powered insights
// # PARAM: csv_file (required, string) - Path to CSV file
// # PARAM: output_report (optional, string) - Path to output report
func (s *ModuleScanner) parseModuleMetadata(scriptPath string) (*ModuleMetadata, error) {
	file, err := os.Open(scriptPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open script: %w", err)
	}
	defer file.Close()

	metadata := &ModuleMetadata{
		Parameters: make([]Parameter, 0),
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Read first 50 lines for metadata (should be at top of file)
	for scanner.Scan() && lineNum < 50 {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Stop at first non-comment line
		if !strings.HasPrefix(line, "#") {
			break
		}

		// Parse metadata comments
		if strings.HasPrefix(line, "# MODULE:") {
			metadata.ID = strings.TrimSpace(strings.TrimPrefix(line, "# MODULE:"))
		} else if strings.HasPrefix(line, "# NAME:") {
			metadata.Name = strings.TrimSpace(strings.TrimPrefix(line, "# NAME:"))
		} else if strings.HasPrefix(line, "# CATEGORY:") {
			metadata.Category = strings.TrimSpace(strings.TrimPrefix(line, "# CATEGORY:"))
		} else if strings.HasPrefix(line, "# DESCRIPTION:") {
			metadata.Description = strings.TrimSpace(strings.TrimPrefix(line, "# DESCRIPTION:"))
		} else if strings.HasPrefix(line, "# PARAM:") {
			param := parseParameterLine(strings.TrimPrefix(line, "# PARAM:"))
			if param != nil {
				metadata.Parameters = append(metadata.Parameters, *param)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading script: %w", err)
	}

	// Use defaults if metadata is missing
	if metadata.ID == "" {
		// Use filename without extension as ID
		metadata.ID = strings.TrimSuffix(filepath.Base(scriptPath), ".sh")
	}

	if metadata.Name == "" {
		// Use ID as name
		metadata.Name = metadata.ID
	}

	if metadata.Category == "" {
		// Use parent directory as category
		metadata.Category = filepath.Base(filepath.Dir(scriptPath))
	}

	if metadata.Description == "" {
		metadata.Description = fmt.Sprintf("Module: %s", metadata.Name)
	}

	return metadata, nil
}

// parseParameterLine parses a parameter definition line
// Format: name (required|optional, type) - description
// Example: csv_file (required, string) - Path to CSV file
func parseParameterLine(line string) *Parameter {
	line = strings.TrimSpace(line)

	// Split on " - " to separate name/spec from description
	parts := strings.SplitN(line, " - ", 2)
	if len(parts) < 1 {
		return nil
	}

	nameSpec := strings.TrimSpace(parts[0])
	description := ""
	if len(parts) > 1 {
		description = strings.TrimSpace(parts[1])
	}

	// Parse name and spec (name (required, type))
	// Find the opening parenthesis
	parenStart := strings.Index(nameSpec, "(")
	parenEnd := strings.Index(nameSpec, ")")

	if parenStart == -1 || parenEnd == -1 {
		// No spec provided, just name
		return &Parameter{
			Name:        strings.TrimSpace(nameSpec),
			Type:        "string",
			Description: description,
			Required:    false,
		}
	}

	name := strings.TrimSpace(nameSpec[:parenStart])
	spec := strings.TrimSpace(nameSpec[parenStart+1 : parenEnd])

	// Parse spec (required|optional, type)
	specParts := strings.Split(spec, ",")

	required := false
	paramType := "string"

	if len(specParts) > 0 {
		req := strings.TrimSpace(specParts[0])
		required = (req == "required")
	}

	if len(specParts) > 1 {
		paramType = strings.TrimSpace(specParts[1])
	}

	return &Parameter{
		Name:        name,
		Type:        paramType,
		Description: description,
		Required:    required,
	}
}
