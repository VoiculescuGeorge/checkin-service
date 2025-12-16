package database

import (
	"database/sql"
	"fmt"

	"checkin.service/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib" 
)

// NewConnection creates and verifies a new database connection pool.
func NewConnection(cfg config.Config) (*sql.DB, error) {
	// Construct connection string from config
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
	)

	db, err := sql.Open("pgx", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Ping the database to verify the connection is alive
	return db, db.Ping()
}
