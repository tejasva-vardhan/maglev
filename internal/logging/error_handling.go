package logging

import (
	"fmt"
	"io"
	"log/slog"
)

// SafeCloseWithLogging closes a resource and logs any errors that occur
func SafeCloseWithLogging(closer io.Closer, logger *slog.Logger, operation string) {
	if closer == nil {
		return
	}

	if err := closer.Close(); err != nil {
		LogError(logger, "failed to close resource", err,
			slog.String("operation", operation),
			slog.String("component", "resource_management"))
	}
}

// SafeRollbackWithLogging rolls back a transaction and logs any errors that occur
// It ignores "already committed/rolled back" errors as these are expected when using defer
func SafeRollbackWithLogging(tx interface{ Rollback() error }, logger *slog.Logger, operation string) {
	if tx == nil {
		return
	}

	if err := tx.Rollback(); err != nil {
		// Ignore the common "already committed or rolled back" error
		// This happens when defer rollback is called after successful commit
		errStr := err.Error()
		if errStr == "sql: transaction has already been committed or rolled back" {
			return
		}

		LogError(logger, "failed to rollback transaction", err,
			slog.String("operation", operation),
			slog.String("component", "database"))
	}
}

// HandleDeferredError handles errors from deferred operations
// It modifies the original error to include deferred operation failures
func HandleDeferredError(originalErr *error, deferredOp func() error, logger *slog.Logger, operation string) {
	if deferredOp == nil {
		return
	}

	if err := deferredOp(); err != nil {
		// Log the deferred error
		LogError(logger, "deferred operation failed", err,
			slog.String("operation", operation),
			slog.String("component", "deferred_cleanup"))

		// If there was no original error, set this as the error
		if *originalErr == nil {
			*originalErr = fmt.Errorf("%s failed: %w", operation, err)
		}
		// If there was an original error, we keep it but log the deferred error
		// The original error takes precedence as it's usually more important
	}
}
