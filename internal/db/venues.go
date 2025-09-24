package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/adamkadda/ntumiwa-site/shared/models"
	"github.com/jackc/pgx/v5"
)

func (db *DB) GetVenue(
	ctx context.Context,
	venueID int,
) (models.VenueResponse, error) {
	// NOTE: Light transaction
	deadline := time.Now().Add(db.timeout)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	query := `SELECT venue_id, address FROM venues WHERE venue_id = $1`

	rows, err := db.pool.Query(ctx, query, venueID)
	if err != nil {
		return models.VenueResponse{}, fmt.Errorf("query failed: %w", err)
	}

	venueRow, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[models.VenueRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.VenueResponse{}, fmt.Errorf("venue with id %d not found: %w", venueID, ErrResourceNotFound)
		}

		return models.VenueResponse{}, fmt.Errorf("failed to collect venue with id %d: %w", venueID, err)
	}

	return venueRow.ToResponse(), nil
}

func (db *DB) GetVenues(
	ctx context.Context,
) ([]models.VenueResponse, error) {
	// NOTE: Moderately heavy transaction
	deadline := time.Now().Add(db.timeout * 2)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	query := `SELECT venue_id, address FROM venues ORDER BY address`

	rows, err := db.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	venueRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.VenueRow])
	if err != nil {
		return nil, fmt.Errorf("failed to collect venue rows: %w", err)
	}

	venueResponses := make([]models.VenueResponse, 0, len(venueRows))
	for _, row := range venueRows {
		venueResponses = append(venueResponses, row.ToResponse())
	}

	return venueResponses, nil
}

func (db *DB) CreateVenue(
	ctx context.Context,
	req models.VenueRequest,
) (models.VenueResponse, error) {
	// NOTE: Light transaction
	deadline := time.Now().Add(db.timeout)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	if len(req.Address) < 1 {
		return models.VenueResponse{}, fmt.Errorf("address field required: %w", ErrIncompleteResource)
	}

	query := `
	INSERT into venues (address)
	VALUES ($1)
	RETURNING venue_id, address
	`

	rows, err := db.pool.Query(ctx, query, req.Address)
	if err != nil {
		return models.VenueResponse{}, fmt.Errorf("query failed: %w", err)
	}

	venueRow, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[models.VenueRow])
	if err != nil {
		return models.VenueResponse{}, fmt.Errorf("failed to collect inserted venue row: %w", err)
	}

	return venueRow.ToResponse(), nil
}

func (db *DB) UpdateVenue(
	ctx context.Context,
	venueID int,
	req models.VenueRequest,
) (models.VenueResponse, error) {
	// NOTE: Light transaction
	deadline := time.Now().Add(db.timeout)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	query := `
	UPDATE venues
	SET address = $2
	WHERE venue_id = $1
	RETURNING venue_id, address
	`

	rows, err := db.pool.Query(ctx, query, venueID)
	if err != nil {
		return models.VenueResponse{}, fmt.Errorf("query failed: %w", err)
	}

	venueRow, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[models.VenueRow])
	if err != nil {
		return models.VenueResponse{}, fmt.Errorf("failed to collect updated venue row: %w", err)
	}

	return venueRow.ToResponse(), nil
}

func (db *DB) DeleteVenue(
	ctx context.Context,
	venueID int,
) error {
	// NOTE: Light transaction
	deadline := time.Now().Add(db.timeout)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	query := `DELETE FROM venues WHERE venue_id = $1`

	cmdTag, err := db.pool.Exec(ctx, query, venueID)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	rowsAffected := cmdTag.RowsAffected()

	if rowsAffected == 0 {
		return ErrResourceNotFound
	}

	return nil
}
