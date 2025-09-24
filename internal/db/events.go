package db

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/adamkadda/ntumiwa-site/shared/models"
	"github.com/jackc/pgx/v5"
)

func (db *DB) GetEvent(
	ctx context.Context,
	eventID int,
) (models.EventFullResponse, error) {
	// NOTE: Light transaction
	deadline := time.Now().Add(db.timeout)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	query := `
	SELECT
		event_id,
		event_title,
		event_date,
		ticket_link,
		venue_id,	
		programme_id,
		status,
		notes,
		created_at,
		updated_at
	FROM events
	WHERE event_id = $1
	`

	rows, err := db.pool.Query(ctx, query, eventID)
	if err != nil {
		return models.EventFullResponse{}, fmt.Errorf("failed to get event with id %d: %w", eventID, err)
	}

	eventRow, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[models.EventRow])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.EventFullResponse{}, fmt.Errorf("event with id %d not found: %w", eventID, ErrResourceNotFound)
		}

		return models.EventFullResponse{}, fmt.Errorf("failed to collect event: %w", err)
	}

	venue, err := db.GetVenue(ctx, eventRow.VenueID)
	if err != nil {
		return models.EventFullResponse{}, err
	}

	ctxProgramme, cancel := context.WithTimeout(ctx, db.timeout)
	defer cancel()

	programme, err := db.GetProgramme(ctxProgramme, eventRow.ProgrammeID)
	if err != nil {
		return models.EventFullResponse{}, err
	}

	return eventRow.ToFullResponse(programme, venue), nil
}

func (db *DB) GetEvents(
	ctx context.Context,
	timeframe string,
	status string,
) ([]models.EventListResponse, error) {
	// NOTE: Moderately heavy transaction
	deadline := time.Now().Add(db.timeout * 2)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	query := `
	SELECT
		event_id,
		event_date,
		status,
		createdAt,
		updatedAt
	FROM events
	`

	// NOTE: Modify query on filters
	var conditions []string

	if timeframe != "" {
		switch timeframe {
		case models.TimeframeUpcoming:
			conditions = append(conditions, "event_date >= CURRENT_DATE")
		case models.TimeframePast:
			conditions = append(conditions, "event_date < CURRENT_DATE")
		default:
			return nil, fmt.Errorf("invalid timeframe filter: %s", timeframe)
		}
	}

	if status != "" {
		switch status {
		case models.StatusDraft, models.StatusPublished, models.StatusArchived:
			conditions = append(conditions, fmt.Sprintf("status = '%s'", status))
		default:
			return nil, fmt.Errorf("invalid status filter: %s", status)
		}
	}

	var where string
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	// NOTE: The WHERE clause is optional
	if where != "" {
		query += where
	}

	rows, err := db.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	eventRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.EventRow])
	if err != nil {
		return nil, fmt.Errorf("failed to collect event list rows: %w", err)
	}

	eventList := make([]models.EventListResponse, 0, len(eventRows))
	for _, row := range eventRows {
		eventList = append(eventList, row.ToListResponse())
	}

	return eventList, nil
}

func (db *DB) CreateEvent(
	ctx context.Context,
	req models.EventRequest,
) (models.EventFullResponse, error) {
	// NOTE: Moderately heavy transaction
	deadline := time.Now().Add(db.timeout * 2)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return models.EventFullResponse{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// NOTE: Titles are required
	if req.Title == "" {
		return models.EventFullResponse{}, ErrTitleNotFound
	}

	// NOTE: Dynamically build query

	// Contains the column titles, and values to be inserted
	var cols []string = []string{"event_title"}
	var vals []any = []any{req.Title}

	// NOTE: Contains PostgreSQL placeholders e.g. $1, $2, $3 ...
	var placeholders []string = []string{"$1"}

	if !req.Date.IsZero() {
		cols = append(cols, "event_date")
		vals = append(vals, req.Date)
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(cols)))
	}

	if req.TicketLink != "" {
		cols = append(cols, "ticket_link")
		vals = append(vals, req.TicketLink)
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(cols)))
	}

	if req.VenueID != 0 {
		cols = append(cols, "venue_id")
		vals = append(vals, req.VenueID)
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(cols)))
	}

	if req.ProgrammeID != 0 {
		cols = append(cols, "programme_id")
		vals = append(vals, req.ProgrammeID)
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(cols)))
	}

	if req.Notes != "" {
		cols = append(cols, "notes")
		vals = append(vals, req.Notes)
		placeholders = append(placeholders, fmt.Sprintf("$%d", len(cols)))
	}

	// Build query
	query := fmt.Sprintf(`
		INSERT INTO events (%s)
		VALUES (%s)
		RETURNING event_id
		`, strings.Join(cols, ", "), strings.Join(placeholders, ", "))

	var eventID int

	// the spread ... operator is cool
	err = tx.QueryRow(ctx, query, vals...).Scan(&eventID)
	if err != nil {
		return models.EventFullResponse{}, fmt.Errorf("failed to insert event: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return models.EventFullResponse{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return db.GetEvent(ctx, eventID)
}

func (db *DB) UpdateEvent(
	ctx context.Context,
	eventID int,
	req models.EventRequest,
) (models.EventFullResponse, error) {
	// NOTE: Moderately heavy transaction
	deadline := time.Now().Add(db.timeout * 2)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return models.EventFullResponse{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// NOTE: Dynamically build query
	var clauses []string
	var vals []any

	// NOTE: Blank req.Title field -> no changes to title
	if req.Title != "" {
		clauses = append(clauses, fmt.Sprintf("event_title = $%d", len(clauses)+1))
		vals = append(vals, req.Title)
	}
	if !req.Date.IsZero() {
		clauses = append(clauses, fmt.Sprintf("event_date = $%d", len(clauses)+1))
		vals = append(vals, req.Date)
	}
	if req.TicketLink != "" {
		clauses = append(clauses, fmt.Sprintf("ticket_link = $%d", len(clauses)+1))
		vals = append(vals, req.TicketLink)
	}
	if req.VenueID != 0 {
		clauses = append(clauses, fmt.Sprintf("venue_id = $%d", len(clauses)+1))
		vals = append(vals, req.VenueID)
	}
	if req.ProgrammeID != 0 {
		clauses = append(clauses, fmt.Sprintf("programme_id = $%d", len(clauses)+1))
		vals = append(vals, req.ProgrammeID)
	}
	if req.Notes != "" {
		clauses = append(clauses, fmt.Sprintf("notes = $%d", len(clauses)+1))
		vals = append(vals, req.Notes)
	}

	// Check if there's anything to update
	if len(clauses) == 0 {
		return models.EventFullResponse{}, fmt.Errorf("empty event request")
	}

	// Add eventID as the final parameter
	vals = append(vals, eventID)

	query := fmt.Sprintf(`
		UPDATE events 
		SET %s 
		WHERE event_id = $%d
		`,
		strings.Join(clauses, ", "),
		len(vals))

	cmdTag, err := tx.Exec(ctx, query, vals...)
	if err != nil {
		return models.EventFullResponse{}, fmt.Errorf("failed to update event: %w", err)
	}

	rowsAffected := cmdTag.RowsAffected()

	if rowsAffected == 0 {
		return models.EventFullResponse{}, ErrResourceNotFound
	}

	if rowsAffected > 1 {
		return models.EventFullResponse{}, fmt.Errorf("unexpected: %d rows affected, multiple events with id %d", rowsAffected, eventID)
	}

	if err = tx.Commit(ctx); err != nil {
		return models.EventFullResponse{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return db.GetEvent(ctx, eventID)
}

func (db *DB) DraftEvent(
	ctx context.Context,
	eventID int,
) (models.EventFullResponse, error) {
	// NOTE: Light transaction
	deadline := time.Now().Add(db.timeout)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	var exists bool

	existsQuery := `SELECT EXISTS(SELECT 1 FROM events WHERE event_id = $1)`
	err := db.pool.QueryRow(ctx, existsQuery, eventID).Scan(&exists)
	if err != nil {
		return models.EventFullResponse{}, fmt.Errorf("existence check failed: %w", err)
	}

	if !exists {
		return models.EventFullResponse{}, ErrResourceNotFound
	}

	draftQuery := `
		UPDATE events
		SET status = 'draft'
		WHERE event_id = $1 AND status <> 'draft'
	`

	cmdTag, err := db.pool.Exec(ctx, draftQuery, eventID)
	if err != nil {
		return models.EventFullResponse{}, fmt.Errorf("failed to draft event: %w", err)
	}

	rowsAffected := cmdTag.RowsAffected()

	if rowsAffected == 0 {
		return models.EventFullResponse{}, ErrStatusUnchanged
	}

	if rowsAffected > 1 {
		return models.EventFullResponse{}, fmt.Errorf("unexpected: %d rows affected, multiple events with id %d", rowsAffected, eventID)
	}

	return db.GetEvent(ctx, eventID)
}

func (db *DB) PublishEvent(
	ctx context.Context,
	eventID int,
) (models.EventFullResponse, error) {
	// NOTE: Moderately heavy transaction
	deadline := time.Now().Add(db.timeout * 2)
	ctx, cancel := context.WithDeadline(ctx, deadline)

	defer cancel()
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return models.EventFullResponse{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// NOTE: Validate event completeness before publishing
	query := `
	SELECT EXISTS (
		SELECT 1
		FROM events
		WHERE 
			event_id = $1
			AND event_title IS NOT NULL
			AND event_date IS NOT NULL
			AND ticket_link IS NOT NULL
			AND venue_id IS NOT NULL
			AND programme_id IS NOT NULL
	)
	`

	var exists bool
	err = tx.QueryRow(ctx, query, eventID).Scan(&exists)
	if err != nil {
		return models.EventFullResponse{}, fmt.Errorf("query failed: %w", err)
	}

	if !exists {
		return models.EventFullResponse{}, ErrIncompleteResource
	}

	// The actual publishing
	cmdTag, err := tx.Exec(ctx, `UPDATE events SET status = 'published' WHERE event_id = $1 AND status <> 'published'`, eventID)
	if err != nil {
		return models.EventFullResponse{}, fmt.Errorf("failed to publish event: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return models.EventFullResponse{}, fmt.Errorf("event with id %d is already published: %w", eventID, ErrResourceNotFound)
	}

	if err = tx.Commit(ctx); err != nil {
		return models.EventFullResponse{}, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return db.GetEvent(ctx, eventID)
}

func (db *DB) ArchiveEvent(
	ctx context.Context,
	eventID int,
) (models.EventFullResponse, error) {
	// NOTE: Light transaction
	deadline := time.Now().Add(db.timeout)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	var exists bool

	existsQuery := `SELECT EXISTS(SELECT 1 FROM events WHERE event_id = $1)`
	err := db.pool.QueryRow(ctx, existsQuery, eventID).Scan(&exists)
	if err != nil {
		return models.EventFullResponse{}, fmt.Errorf("existence check failed: %w", err)
	}

	if !exists {
		return models.EventFullResponse{}, ErrResourceNotFound
	}

	archiveQuery := `
		UPDATE events
		SET status = 'draft'
		WHERE event_id = $1 AND status <> 'archive'
	`

	cmdTag, err := db.pool.Exec(ctx, archiveQuery, eventID)
	if err != nil {
		return models.EventFullResponse{}, fmt.Errorf("failed to archive event: %w", err)
	}

	rowsAffected := cmdTag.RowsAffected()

	if rowsAffected == 0 {
		return models.EventFullResponse{}, ErrStatusUnchanged
	}

	if rowsAffected > 1 {
		return models.EventFullResponse{}, fmt.Errorf("unexpected: %d rows affected, multiple events with id %d", rowsAffected, eventID)
	}

	return db.GetEvent(ctx, eventID)
}

func (db *DB) DeleteEvent(
	ctx context.Context,
	eventID int,
) error {
	// NOTE: Light transaction
	deadline := time.Now().Add(db.timeout)
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	query := `DELETE FROM events WHERE event_id = $1`

	cmdTag, err := db.pool.Exec(ctx, query, eventID)
	if err != nil {
		return fmt.Errorf("failed to delete event with id %d: %w", eventID, err)
	}

	rowsAffected := cmdTag.RowsAffected()

	if rowsAffected == 0 {
		return ErrResourceNotFound
	}

	if rowsAffected > 1 {
		return fmt.Errorf("unexpected: %d rows affected, multiple events with id %d", rowsAffected, eventID)
	}

	return nil
}
