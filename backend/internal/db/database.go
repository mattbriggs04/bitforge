package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

const DefaultDatabaseURL = "file:bitforge.db?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)"

func OpenDatabase(ctx context.Context, databaseURL string) (*sql.DB, error) {
	if databaseURL == "" {
		databaseURL = DefaultDatabaseURL
	}

	db, err := sql.Open("sqlite", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(30 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}
	return db, nil
}
