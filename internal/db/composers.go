package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/adamkadda/ntumiwa/shared/models"
	"github.com/jackc/pgx/v5"
)

func (db *DB) GetComposer(
	ctx context.Context,
	composerID int,
) (models.ComposerResponse, error) {
	// NOTE: Light transaction
	deadline := time.Now().Add(db.timeout)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	query := `
	SELECT
		composer_id,
		short_name,
		full_name
	FROM composers
	WHERE composer_id = $1
	`

	rows, err := db.pool.Query(ctx, query, composerID)
	if err != nil {
		return models.ComposerResponse{}, fmt.Errorf("query failed for composer with id %d: %w", composerID, err)
	}

	composerRow, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[models.ComposerRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.ComposerResponse{}, fmt.Errorf("composer with id %d not found: %w", composerID, ErrResourceNotFound)
		}

		return models.ComposerResponse{}, fmt.Errorf("failed to collect composer with id %d: %w", composerID, err)
	}

	return composerRow.ToResponse(), nil
}

// Strings are heap allocated, and since copying struct headers is inexpensive
// I would think that the allocation overhead of doing:
// rows -> composerRows -> composerResponses
// is inexpensive.
func (db *DB) GetComposers(
	ctx context.Context,
) ([]models.ComposerResponse, error) {
	// NOTE: Moderately heavy transaction
	deadline := time.Now().Add(db.timeout * 2)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	query := "SELECT composer_id, short_name, full_name FROM composers ORDER BY short_name"

	rows, err := db.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	composerRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.ComposerRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("empty table: %w", ErrResourceNotFound)
		}
		return nil, fmt.Errorf("failed to collect composers: %w", err)
	}

	composerResponses := make([]models.ComposerResponse, 0, len(composerRows))
	for _, row := range composerRows {
		composerResponses = append(composerResponses, row.ToResponse())
	}

	return composerResponses, nil
}

func (db *DB) CreateComposer(
	ctx context.Context,
	req models.ComposerRequest,
) (models.ComposerResponse, error) {
	// NOTE: Light transaction
	deadline := time.Now().Add(db.timeout)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	// Validate request
	if req.ShortName == "" {
		return models.ComposerResponse{}, fmt.Errorf("short name field required: %w", ErrIncompleteResource)
	}

	if req.FullName == "" {
		return models.ComposerResponse{}, fmt.Errorf("full name field required: %w", ErrIncompleteResource)
	}

	query := `
	INSERT INTO composers (short_name, full_name)
	VALUES ($1, $2)
	RETURNING composer_id, short_name, full_name
	`

	rows, err := db.pool.Query(ctx, query, req.ShortName, req.FullName)
	if err != nil {
		return models.ComposerResponse{}, fmt.Errorf("query failed: %w", err)
	}

	composerRow, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[models.ComposerRow])
	if err != nil {
		return models.ComposerResponse{}, fmt.Errorf("failed to collect composer: %w", err)
	}

	return composerRow.ToResponse(), nil
}

func (db *DB) UpdateComposer(
	ctx context.Context,
	composerID int,
	composerRequest models.ComposerRequest,
) (models.ComposerResponse, error) {
	// NOTE: Light transaction
	deadline := time.Now().Add(db.timeout)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	query := `
	UPDATE composers
	SET short_name = $2, full_name = $3
	WHERE composer_id = $1
	RETURNING composer_id, short_name, full_name
	`

	rows, err := db.pool.Query(ctx, query, composerID, composerRequest.ShortName, composerRequest.FullName)
	if err != nil {
		return models.ComposerResponse{}, fmt.Errorf("query failed: %w", err)
	}

	composerRow, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[models.ComposerRow])
	if err != nil {
		return models.ComposerResponse{}, fmt.Errorf("failed to update composer with id %d: %w", composerID, err)
	}

	return composerRow.ToResponse(), nil
}

func (db *DB) DeleteComposer(
	ctx context.Context,
	composerID int,
) error {
	// NOTE: Light transaction
	deadline := time.Now().Add(db.timeout)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	query := "DELETE FROM composers WHERE composer_id = $1"

	cmdTag, err := db.pool.Exec(ctx, query, composerID)
	if err != nil {
		return fmt.Errorf("failed to delete composer with id %d: %w", composerID, err)
	}

	rowsAffected := cmdTag.RowsAffected()

	if rowsAffected == 0 {
		return ErrResourceNotFound
	}

	return nil
}
