package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/adamkadda/ntumiwa/internal/models"
	"github.com/jackc/pgx/v5"
)

func (db *DB) GetPiece(
	ctx context.Context,
	pieceID int,
) (*models.PieceResponse, error) {
	// NOTE: Light transaction
	deadline := time.Now().Add(db.timeout)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	query := `
	SELECT
		p.piece_id,
		p.piece_title,
		p.composer_id,
		c.short_name AS composer_name
	FROM pieces p
	LEFT JOIN composers c ON p.composer_id = c.composer_id
	WHERE p.piece_id = $1
	`

	rows, err := db.pool.Query(ctx, query, pieceID)
	if err != nil {
		return nil, fmt.Errorf("query failed for piece with id %d: %w", pieceID, err)
	}

	pieceRow, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[models.PieceRow])
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, fmt.Errorf("piece with id %d not found: %w", pieceID, ErrResourceNotFound)
		case errors.Is(err, pgx.ErrTooManyRows):
			return nil, fmt.Errorf("multiple pieces with id %d found: %w", pieceID, err)
		default:
			return nil, fmt.Errorf("failed to collect piece with id %d: %w", pieceID, err)
		}
	}

	pieceResponse := pieceRow.ToResponse()

	return &pieceResponse, nil
}

// TODO: Consider implementing get by composer (id)
func (db *DB) ListPieces(
	ctx context.Context,
) ([]models.PieceResponse, error) {
	// NOTE: Moderately heavy transaction
	deadline := time.Now().Add(db.timeout * 2)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	query := `
	SELECT
		p.piece_id,
		p.piece_title,
		p.composer_id,
		c.short_name AS composer_name
	FROM pieces p
	LEFT JOIN composers c ON p.composer_id = c.composer_id
	`

	rows, err := db.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	pieceRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.PieceRow])
	if err != nil {
		return nil, fmt.Errorf("failed to collect pieces: %w", err)
	}

	pieceResponses := make([]models.PieceResponse, 0, len(pieceRows))
	for _, row := range pieceRows {
		pieceResponses = append(pieceResponses, row.ToResponse())
	}

	return pieceResponses, nil
}

func (db *DB) CreatePiece(
	ctx context.Context,
	req models.PieceRequest,
) (*models.PieceResponse, error) {
	// NOTE: Light transaction
	deadline := time.Now().Add(db.timeout)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	query := `
	WITH ins AS (
		INSERT INTO pieces (piece_title, composer_id)
		VALUES ($1, $2)
		RETURNING piece_id, piece_title, composer_id
	)
	SELECT
		ins.piece_id,
		ins.piece_title,
		ins.composer_id,
		c.short_name AS composer_name
	FROM ins 
	LEFT JOIN composers c ON ins.composer_id = c.composer_id;
	`

	// Validate request
	if len(req.Title) < 1 {
		return nil, ErrTitleNotFound
	}

	rows, err := db.pool.Query(ctx, query, req.Title, req.ComposerID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	pieceRow, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[models.PieceRow])
	if err != nil {
		return nil, fmt.Errorf("failed to collect inserted piece: %w", err)
	}

	pieceResponse := pieceRow.ToResponse()

	return &pieceResponse, nil
}

func (db *DB) UpdatePiece(
	ctx context.Context,
	pieceID int,
	req models.PieceRequest,
) (*models.PieceResponse, error) {
	// NOTE: Light transaction
	deadline := time.Now().Add(db.timeout)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	query := `
	WITH upd AS (
		UPDATE pieces
		SET piece_title = $2, composer_id = $3
		WHERE piece_id = $1
		RETURNING piece_id, piece_title, composer_id
	)
	SELECT
		upd.piece_id,
		upd.piece_title,
		upd.composer_id,
		c.short_name AS composer_name
	FROM upd
	LEFT JOIN composers c ON upd.composer_id = c.composer_id
	`

	rows, err := db.pool.Query(ctx, query, pieceID, req.Title, req.ComposerID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	pieceRow, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[models.PieceRow])
	if err != nil {
		return nil, fmt.Errorf("failed to collect updated piece row: %w", err)
	}

	pieceResponse := pieceRow.ToResponse()

	return &pieceResponse, nil
}

func (db *DB) DeletePiece(
	ctx context.Context,
	pieceID int,
) error {
	// NOTE: Light transaction
	deadline := time.Now().Add(db.timeout)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	query := "DELETE FROM pieces WHERE piece_id = $1"

	cmdTag, err := db.pool.Exec(ctx, query, pieceID)
	if err != nil {
		return fmt.Errorf("failed to delete piece with id %d: %w", pieceID, err)
	}

	rowsAffected := cmdTag.RowsAffected()

	if rowsAffected == 0 {
		return ErrResourceNotFound
	}

	return nil
}
