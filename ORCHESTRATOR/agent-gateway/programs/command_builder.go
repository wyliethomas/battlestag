package programs

import (
	"context"
	"os/exec"
	"strings"
)

// CommandBuilder provides a fluent API for building CLI commands.
// Simplifies command construction with optional arguments.
//
// Example usage:
//
//	cmd := NewCommand("task_project_run").
//	    Arg("add").
//	    Flag("name", "Sawmill").
//	    Flag("context", "property").
//	    OptionalFlag("goal", goal, goal != "").
//	    Build(ctx)
type CommandBuilder struct {
	bin  string
	args []string
}

// NewCommand creates a new command builder for the given binary
func NewCommand(bin string) *CommandBuilder {
	return &CommandBuilder{
		bin:  bin,
		args: []string{},
	}
}

// Arg adds a positional argument to the command
//
// Example:
//
//	NewCommand("task_project_run").Arg("add").Arg("list")
//	// Produces: task_project_run add list
func (b *CommandBuilder) Arg(arg string) *CommandBuilder {
	b.args = append(b.args, arg)
	return b
}

// Flag adds a flag with value to the command (--flag value)
//
// Example:
//
//	NewCommand("task_project_run").Flag("name", "Test")
//	// Produces: task_project_run --name Test
func (b *CommandBuilder) Flag(name, value string) *CommandBuilder {
	b.args = append(b.args, "--"+name, value)
	return b
}

// OptionalFlag adds a flag only if the condition is true.
// Useful for optional parameters.
//
// Example:
//
//	NewCommand("task_project_run").
//	    Flag("name", name).
//	    OptionalFlag("goal", goal, goal != "")
//	// If goal is empty, --goal flag is not added
func (b *CommandBuilder) OptionalFlag(name, value string, condition bool) *CommandBuilder {
	if condition {
		b.args = append(b.args, "--"+name, value)
	}
	return b
}

// BoolFlag adds a boolean flag (--flag) without a value
//
// Example:
//
//	NewCommand("task_checklist_run").Arg("list").BoolFlag("pending")
//	// Produces: task_checklist_run list --pending
func (b *CommandBuilder) BoolFlag(name string) *CommandBuilder {
	b.args = append(b.args, "--"+name)
	return b
}

// ConditionalBoolFlag adds a boolean flag only if the condition is true
//
// Example:
//
//	NewCommand("task_checklist_run").
//	    Arg("list").
//	    ConditionalBoolFlag("pending", pendingOnly)
//	// Adds --pending only if pendingOnly is true
func (b *CommandBuilder) ConditionalBoolFlag(name string, condition bool) *CommandBuilder {
	if condition {
		b.args = append(b.args, "--"+name)
	}
	return b
}

// IntFlag adds a flag with an integer value
//
// Example:
//
//	NewCommand("task_checklist_run").Arg("check").IntFlag("id", 5)
//	// Produces: task_checklist_run check --id 5
func (b *CommandBuilder) IntFlag(name string, value int) *CommandBuilder {
	b.args = append(b.args, "--"+name, intToString(value))
	return b
}

// OptionalIntFlag adds an integer flag only if the condition is true
//
// Example:
//
//	NewCommand("task_journal_run").
//	    Arg("list").
//	    OptionalIntFlag("limit", limit, limit > 0)
//	// Adds --limit only if limit > 0
func (b *CommandBuilder) OptionalIntFlag(name string, value int, condition bool) *CommandBuilder {
	if condition {
		b.args = append(b.args, "--"+name, intToString(value))
	}
	return b
}

// Args returns the current argument list (for debugging)
func (b *CommandBuilder) Args() []string {
	return b.args
}

// String returns the full command as a string (for debugging)
func (b *CommandBuilder) String() string {
	parts := append([]string{b.bin}, b.args...)
	return strings.Join(parts, " ")
}

// Build creates an exec.Cmd with context for execution
//
// Example:
//
//	cmd := NewCommand("task_project_run").
//	    Arg("list").
//	    Flag("context", "property").
//	    Build(ctx)
//	output, err := cmd.CombinedOutput()
func (b *CommandBuilder) Build(ctx context.Context) *exec.Cmd {
	return exec.CommandContext(ctx, b.bin, b.args...)
}

// Execute is a convenience method that builds and executes the command,
// returning the combined output and any error.
//
// Example:
//
//	output, err := NewCommand("task_project_run").
//	    Arg("list").
//	    Execute(ctx)
func (b *CommandBuilder) Execute(ctx context.Context) ([]byte, error) {
	cmd := b.Build(ctx)
	return cmd.CombinedOutput()
}

// ==================== Helper Functions ====================

// intToString converts an integer to string for command arguments
func intToString(i int) string {
	// Simple conversion without importing strconv
	if i == 0 {
		return "0"
	}

	negative := i < 0
	if negative {
		i = -i
	}

	result := ""
	for i > 0 {
		digit := i % 10
		result = string(rune('0'+digit)) + result
		i /= 10
	}

	if negative {
		result = "-" + result
	}

	return result
}
