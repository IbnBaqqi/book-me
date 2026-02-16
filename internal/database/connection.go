// Package database provides database connection, queries and transaction management.
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/IbnBaqqi/book-me/internal/config"
	_ "github.com/lib/pq" // PostgreSQL driver registration
)

// DB wraps the database connection pool & queries
type DB struct {
	*sql.DB
	*Queries
}

// Tx wraps a database transaction
type Tx struct {
	*sql.Tx
}

// Connect establishes a connection to the database
func Connect(ctx context.Context, cfg *config.AppConfig) (*DB, error) {
	// Initialize database
	dbConn, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection, ping with context
	if err := dbConn.PingContext(ctx); err != nil {
		if err = dbConn.Close(); err != nil {
			slog.Error("failed to close database connection", "error", err)
		}
		slog.Error("failed to ping database", "error", err)
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{
		DB: dbConn,
		Queries: New(dbConn),
	}, nil
}

// Close closes the database connection and logs the closure.
func (db *DB) Close() error {
	slog.Info("closing database connection")
	return db.DB.Close()
}

// BeginTx starts a new database transaction with the specified isolation level
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.DB.BeginTx(ctx, opts)
	if err != nil {
		slog.Error("failed to begin transaction", "error", err)
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	slog.Debug("transaction started")
	return &Tx{
		Tx: tx,
	}, nil
}

// Commit commits the transaction
func (tx *Tx) Commit() error {
	if err := tx.Tx.Commit(); err != nil {
		slog.Error("failed to commit transaction", "error", err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.Debug("transaction committed")
	return nil
}

// Rollback rolls back the transaction
func (tx *Tx) Rollback() error {
	if err := tx.Tx.Rollback(); err != nil {
		if errors.Is(err, sql.ErrTxDone) {
			slog.Debug("transaction already closed, ignoring rollback")
			return nil
		}
		slog.Error("failed to rollback transaction", "error", err)
		return fmt.Errorf("failed to rollback transaction: %w", err)
	}

	slog.Debug("transaction rolled back")
	return nil
}