package gtfsdb

// Config holds configuration options for the Client
type Config struct {
	// Database configuration
	DBPath  string // Path to SQLite database file
	verbose bool   // Verbose logging
}

func NewConfig(dbPath string, verbose bool) Config {
	config := Config{
		DBPath:  dbPath,
		verbose: verbose,
	}

	return config
}
