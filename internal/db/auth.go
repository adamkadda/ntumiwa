package db

import (
	"context"
	"time"
)

// TODO: Complete implementation
func (db *DB) UserExists(
	ctx context.Context,
	username string,
) (bool, error) {
	// NOTE: Light transaction
	deadline := time.Now().Add(db.timeout)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	query := `
	SELECT EXISTS(
		SELECT 1
		FROM users
		WHERE username = $1
	)
	`

	var exists bool
	if err := db.pool.QueryRow(ctx, query, username).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}
