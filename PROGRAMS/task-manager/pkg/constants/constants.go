package constants

// Database and configuration paths
const (
	// DefaultDBPath is the default location for the task manager database
	DefaultDBPath = "~/.local/share/task-manager/tasks.db"

	// DefaultConfigDir is the default location for configuration files
	DefaultConfigDir = "~/.config/task-manager/"

	// DefaultInstallDir is the default location for installed binaries
	DefaultInstallDir = "~/.local/bin"
)

// Project status values
const (
	// StatusActive indicates a project is currently being worked on
	StatusActive = "active"

	// StatusPaused indicates a project is temporarily on hold
	StatusPaused = "paused"

	// StatusBlocked indicates a project is waiting on external dependencies
	StatusBlocked = "blocked"

	// StatusComplete indicates a project has been finished
	StatusComplete = "complete"
)

// Default contexts seeded during installation
var DefaultContexts = []string{
	"property",
	"house",
	"pxp",
	"personal",
}

// Executable names (must match install script)
const (
	ExecProject   = "task_project_run"
	ExecChecklist = "task_checklist_run"
	ExecJournal   = "task_journal_run"
	ExecQuery     = "task_query_run"
)

// Exit codes (kept in exitcodes package, listed here for reference)
const (
	ExitSuccess  = 0 // Success
	ExitArgsErr  = 1 // Invalid arguments
	ExitDBError  = 2 // Database error
	ExitNotFound = 3 // Resource not found
)
