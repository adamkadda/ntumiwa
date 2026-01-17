package db

import (
	"context"
	"fmt"
	"time"

	"github.com/adamkadda/ntumiwa/internal/models"
	"github.com/jackc/pgx/v5"
)

func (db *DB) ListPerformances(
	ctx context.Context,
	timeframe string,
) ([]models.Performance, error) {
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
	FROM events
	WHERE status = 'published'
	`

	switch timeframe {
	case models.TimeframeUpcoming:
		query += " AND event_date >= CURRENT_DATE"
	case models.TimeframePast:
		query += " AND event_date < CURRENT_DATE"
	default:
		return nil, fmt.Errorf("invalid timeframe filter: %s", timeframe)
	}

	eventRows, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByName[]())
}
