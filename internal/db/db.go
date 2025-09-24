package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrResourceNotFound    = errors.New("resource not found")
	ErrIncompleteResource  = errors.New("required fields missing")
	ErrStatusUnchanged     = errors.New("no status change")
	ErrForeignKeyViolation = errors.New("foreign key constraint violation")
	ErrTitleNotFound       = errors.New("title is a required field")
)

type DB struct {
	pool    *pgxpool.Pool
	timeout time.Duration
}

// Instantiates a new DB instance.
//
// Requires a connection string that follows the format
// described in pgx's documentation:
//
// pkg.go.dev/github.com/jackc/pgx/v5/pgxpool#ParseConfig
func New(connString string, timeout time.Duration) *DB {
	cfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		panic(fmt.Errorf("Failed to parse DB config: %w", err))
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		panic(fmt.Errorf("Failed to initialize DB pool: %w", err))
	}

	return &DB{
		pool:    pool,
		timeout: timeout,
	}
}
