package exitcodes

// Standard exit codes for financial-statement-processor executables
const (
	Success = 0

	// Processor-specific exit codes
	ParseError  = 1
	DBError     = 2
	ConfigError = 3

	// Query-specific exit codes
	ArgsError = 1
	// DBError = 2 (shared with processor)
)
