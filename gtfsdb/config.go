package gtfsdb

import "maglev.onebusaway.org/internal/appconf"

const (
	// DefaultBulkInsertBatchSize is the fallback batch size used when fieldsPerRow is
	// unknown or <= 0.
	DefaultBulkInsertBatchSize = 3000

	// sqliteMaxVariables is the hard upper bound on bound parameters per SQLite
	// statement (SQLITE_MAX_VARIABLE_NUMBER).
	sqliteMaxVariables = 32766
)

// Config holds configuration options for the Client
type Config struct {
	// Database configuration
	DBPath  string              // Path to SQLite database file
	Env     appconf.Environment // Environment name: development, test, production.
	verbose bool                // Enable verbose logging

	// Performance tuning
	// BulkInsertBatchSize is reserved for future use. Batch sizes are now computed
	// automatically via SafeBatchSize based on SQLITE_MAX_VARIABLE_NUMBER (32766)
	// and the number of fields bound per row.
	BulkInsertBatchSize int
}

func NewConfig(dbPath string, env appconf.Environment, verbose bool) Config {
	return Config{
		DBPath:              dbPath,
		Env:                 env,
		verbose:             verbose,
		BulkInsertBatchSize: DefaultBulkInsertBatchSize,
	}
}

// SafeBatchSize returns the maximum safe number of rows per multi-row INSERT given
// the number of parameters bound per row. SQLite enforces a hard upper bound of
// SQLITE_MAX_VARIABLE_NUMBER (32766) bound parameters per statement, so the safe
// maximum is 32766 / fieldsPerRow.
// If fieldsPerRow is <= 0, DefaultBulkInsertBatchSize is returned as a fallback.
func (c Config) SafeBatchSize(fieldsPerRow int) int {
	if fieldsPerRow <= 0 {
		return DefaultBulkInsertBatchSize
	}
	return sqliteMaxVariables / fieldsPerRow
}
