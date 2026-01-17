package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrResourceNotFound    = errors.New("resource not found")
	ErrIncompleteResource  = errors.New("required fields missing")
	ErrStatusUnchanged     = errors.New("no status change")
	ErrForeignKeyViolation = errors.New("foreign key constraint violation")
	ErrTitleNotFound       = errors.New("title is a required field")
	ErrEmptyRequest        = errors.New("request body is empty")
	ErrBlankField          = errors.New("blank field forbidden")
	ErrImmutableState      = errors.New("resource cannot be modified")
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

	// WARN: Exerpt from pgxpool documentation:
	// A pool returns [from New or NewWithConfig] without waiting
	// for any connections to be established.
	// Acquire a connection immediately after creating the pool
	// to check if a connection can successfully be established.

	deadline := time.Now().Add(5 * time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()

	err = pool.Ping(ctx)
	if err != nil {
		log.Fatalf("[DB] Ping failed: %v", err)
	}

	return &DB{
		pool:    pool,
		timeout: timeout,
	}
}
