package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Setup configures the global zerolog logger.
func Setup(isLocalDev bool) {
	// Use Unix timestamps for performance and consistency
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if isLocalDev {
		// Pretty printing for local development
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		// Default to JSON output for production
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}
