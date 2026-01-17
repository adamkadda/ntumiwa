package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/adamkadda/ntumiwa/internal/models"
	"github.com/jackc/pgx/v5"
)

func (db *DB) GetProgramme(
	ctx context.Context,
	programmeID int,
) (*models.ProgrammeFullResponse, error) {
	// NOTE: Light transaction
	deadline := time.Now().Add(db.timeout)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	// NOTE: Query for programme metadata
	// Does not obtain the piece_count nor event_count
	// as those fields are not required for the response
	q := `
	SELECT
		programme_id,
		programme_title
	FROM programmes
	WHERE programme_id = $1
	`

	rows, err := db.pool.Query(ctx, q, programmeID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	programmeRow, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[models.ProgrammeRow])
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, fmt.Errorf("programme with id %d not found: %w", programmeID, ErrResourceNotFound)
		case errors.Is(err, pgx.ErrTooManyRows):
			return nil, fmt.Errorf("multiple programmes with id %d found: %w", programmeID, err)
		default:
			return nil, fmt.Errorf("failed to collect programme with id %d: %w", programmeID, err)
		}
	}

	// Query for programme pieces
	q2 := `
	SELECT
		piece_id,
		sequence
	FROM programme_pieces
	WHERE programme_id = $1
	ORDER BY sequence
	`

	rows, err = db.pool.Query(ctx, q2, programmeID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	programmePieceRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.ProgrammePieceRow])
	if err != nil {
		return nil, fmt.Errorf("failed to collect programme id %d's pieces: %w", programmeID, err)
	}

	programmePieces := make([]models.ProgrammePieceResponse, 0, len(programmePieceRows))
	for _, row := range programmePieceRows {
		programmePieces = append(programmePieces, row.ToResponse())
	}

	programmeResponse := programmeRow.ToFullResponse(programmePieces)

	return &programmeResponse, nil
}

// Returns a list of brief programme items, meta data only, pieces are omitted
func (db *DB) ListProgrammes(
	ctx context.Context,
) ([]models.ProgrammeListResponse, error) {
	// NOTE: Moderately heavy transaction
	deadline := time.Now().Add(db.timeout * 2)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	query := `
	SELECT
		p.programme_id,
		p.programme_title, 
		COALESCE(pp.piece_count, 0) AS piece_count,
		COALESCE(e.event_count, 0) AS event_count
	FROM programmes p
	LEFT JOIN (
		SELECT programme_id, COUNT(*) AS piece_count
		FROM programme_pieces
		GROUP BY programme_id
	) pp ON pp.programme_id = p.programme_id
	LEFT JOIN (
		SELECT programme_id, COUNT(*) AS event_count
		FROM events
		GROUP BY programme_id
	) e ON e.programme_id = p.programme_id
	`

	rows, err := db.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	programmeRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.ProgrammeListRow])
	if err != nil {
		return nil, fmt.Errorf("failed to collect programme list rows: %w", err)
	}

	programmeList := make([]models.ProgrammeListResponse, 0, len(programmeRows))
	for _, row := range programmeRows {
		programmeList = append(programmeList, row.ToListResponse())
	}

	return programmeList, nil
}

func (db *DB) CreateProgramme(
	ctx context.Context,
	req models.ProgrammeRequest,
) (*models.ProgrammeFullResponse, error) {
	// NOTE: Moderately heavy transaction
	deadline := time.Now().Add(db.timeout * 3)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	// Begin transaction
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx) // safe to call after committing

	// Validate request
	if len(req.Title) < 1 {
		return nil, ErrTitleNotFound
	}

	// Create programme (title/metadata only), return id
	var programmeID int
	query := `INSERT INTO programmes (programme_title) VALUES ($1) RETURNING programme_id`

	err = tx.QueryRow(ctx, query, req.Title).Scan(&programmeID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert programme: %w", err)
	}

	if len(req.Pieces) > 0 {
		// Init batch, queue INSERT queries
		batch := &pgx.Batch{}

		for _, piece := range req.Pieces {
			batch.Queue(
				`INSERT INTO programme_pieces (programme_id, piece_id, sequence) VALUES ($1, $2, $3)`,
				programmeID, piece.PieceID, piece.Sequence,
			)
		}

		// Collect batch results, defer Close()
		br := tx.SendBatch(ctx, batch)
		defer br.Close()

		// Read each batch result
		for i := 0; i < len(req.Pieces); i++ {
			_, err = br.Exec()
			if err != nil {
				return nil, fmt.Errorf("failed to insert programme piece: %w", err)
			}
		}

		// NOTE: Occupies connection before commit,
		// therefore closing here is necessary in addition
		// to a defer.
		br.Close()
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return db.GetProgramme(ctx, programmeID)
}

func (db *DB) UpdateProgramme(
	ctx context.Context,
	programmeID int,
	req models.ProgrammeRequest,
) (*models.ProgrammeFullResponse, error) {
	// NOTE: Moderately heavy transaction
	deadline := time.Now().Add(db.timeout * 2)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	// Begin transaction, defer rollback
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Update metadata (title)
	if req.Title != "" {
		query := `UPDATE programmes SET programme_title = $2 WHERE programme_id = $1`

		cmdTag, err := tx.Exec(ctx, query, programmeID, req.Title)
		if err != nil {
			return nil, fmt.Errorf("programme metadata update query failed: %w", err)
		}

		if cmdTag.RowsAffected() == 0 {
			return nil, fmt.Errorf("programme with id %d not found: %w", programmeID, ErrResourceNotFound)
		}
	}

	// Update programme pieces
	if len(req.Pieces) > 0 {

		// Delete all related programme_pieces rows
		query := `DELETE FROM programme_pieces WHERE programme_id = $1`

		/*
				 NOTE: 0 rows affected here is a valid outcome
				as it is possible that we are updating a programme
				with no programme pieces associated with it.

			 	Hence, we aren't checking cmdTag.RowsAffected().
		*/

		_, err := tx.Exec(ctx, query, programmeID)
		if err != nil {
			return nil, fmt.Errorf("delete query failed: %w", err)
		}

		// Create new programme pieces
		batch := &pgx.Batch{}

		for _, piece := range req.Pieces {
			batch.Queue(
				`INSERT INTO programme_pieces (programme_id, piece_id, sequence) VALUES ($1, $2, $3)`,
				programmeID, piece.PieceID, piece.Sequence,
			)
		}

		// Collect batch results, defer Close()
		br := tx.SendBatch(ctx, batch)
		defer br.Close()

		// Read each batch result
		for i := 0; i < len(req.Pieces); i++ {
			_, err = br.Exec()
			if err != nil {
				return nil, fmt.Errorf("failed to insert programme piece: %w", err)
			}
		}

		br.Close()

	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return db.GetProgramme(ctx, programmeID)
}

func (db *DB) DeleteProgramme(
	ctx context.Context,
	programmeID int,
) error {
	// NOTE: Moderately heavy transaction
	deadline := time.Now().Add(db.timeout * 2)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	query := `DELETE FROM programmes WHERE programme_id = $1`

	cmdTag, err := db.pool.Exec(ctx, query, programmeID)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	rowsAffected := cmdTag.RowsAffected()

	if rowsAffected == 0 {
		return ErrResourceNotFound
	}

	return nil
}
