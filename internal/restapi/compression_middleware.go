package restapi

import (
	"net/http"

	"github.com/klauspost/compress/gzhttp"
)

// CompressionConfig holds configuration options for response compression
type CompressionConfig struct {
	// MinSize is the minimum response size in bytes to compress (default: 1024)
	MinSize int
	// Level is the compression level 1-9 (default: 6 for balanced speed/compression)
	Level int
}

// DefaultCompressionConfig returns sensible defaults for compression
func DefaultCompressionConfig() CompressionConfig {
	return CompressionConfig{
		MinSize: 1024, // 1KB minimum
		Level:   6,    // Balanced compression
	}
}

// NewCompressionMiddleware creates a compression middleware with the given configuration
func NewCompressionMiddleware(config CompressionConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// Configure gzhttp with our settings
		wrapper, err := gzhttp.NewWrapper(
			gzhttp.MinSize(config.MinSize),
			gzhttp.CompressionLevel(config.Level),
		)
		if err != nil {
			// Fallback to default if configuration fails
			return gzhttp.GzipHandler(next)
		}
		return wrapper(next)
	}
}

// CompressionMiddleware applies gzip compression with default settings
func CompressionMiddleware(next http.Handler) http.Handler {
	config := DefaultCompressionConfig()
	return NewCompressionMiddleware(config)(next)
}
