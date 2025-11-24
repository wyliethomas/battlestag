package exitcodes

// Standard exit codes for task-manager executables
const (
	Success   = 0
	ArgsError = 1
	DBError   = 2
	NotFound  = 3 // Project, context, item, or entry not found
)
