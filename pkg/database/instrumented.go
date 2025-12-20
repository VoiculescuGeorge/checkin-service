package database

import (
	"database/sql"
	"fmt"

	"checkin.service/internal/config"
	"github.com/XSAM/otelsql"
	_ "github.com/jackc/pgx/v5/stdlib" // Register pgx driver
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// NewInstrumentedConnection creates a database connection with OpenTelemetry instrumentation.
func NewInstrumentedConnection(cfg config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	// Connect to database with OpenTelemetry instrumentation
	// otelsql.Open wraps the driver to intercept queries and create spans
	db, err := otelsql.Open("pgx", dsn,
		otelsql.WithAttributes(semconv.DBSystemPostgreSQL),
		otelsql.WithSQLCommenter(true),
	)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
