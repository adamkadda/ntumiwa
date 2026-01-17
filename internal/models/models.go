package models

import (
	"time"
)

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusArchived  = "archived"
)

const (
	TimeframePast     = "past"
	TimeframeUpcoming = "upcoming"
)

const htmlFormat = "2006-01-02"
const textFormat = "2 January, 2006"

type EventRow struct {
	ID          int        `db:"event_id"`
	Title       string     `db:"event_title"`
	Date        *time.Time `db:"event_date"`
	TicketLink  *string    `db:"ticket_link"`
	VenueID     *int       `db:"venue_id"`
	ProgrammeID *int       `db:"programme_id"`
	Status      string     `db:"status"`
	Notes       *string    `db:"notes"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

type EventListRow struct {
	ID        int        `db:"event_id"`
	Title     string     `db:"event_title"`
	Date      *time.Time `db:"event_date"`
	Status    string     `db:"status"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
}

type EventRequest struct {
	Title       *string    `json:"title"`
	Date        *time.Time `json:"date"`
	TicketLink  *string    `json:"ticket_link"`
	VenueID     *int       `json:"venue_id"`
	ProgrammeID *int       `json:"programme_id"`
	Notes       *string    `json:"notes"`
}

type EventFullResponse struct {
	ID         int                    `json:"id"`
	Title      string                 `json:"title"`
	Venue      *VenueResponse         `json:"venue"`
	Date       *time.Time             `json:"date"`
	TextDate   *string                `json:"text_date"`
	TicketLink *string                `json:"ticket_link"`
	Programme  *ProgrammeFullResponse `json:"programme"`
	Status     string                 `json:"status"`
	Notes      *string                `json:"notes"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

func (r *EventRow) ToFullResponse(
	venue *VenueResponse,
	programme *ProgrammeFullResponse,
) EventFullResponse {
	var textDate *string
	if r.Date != nil {
		textDate = ptr(r.Date.Format(textFormat))
	}

	return EventFullResponse{
		ID:         r.ID,
		Title:      r.Title,
		Venue:      venue,
		Date:       r.Date,
		TextDate:   textDate,
		TicketLink: r.TicketLink,
		Programme:  programme,
		Status:     r.Status,
		Notes:      r.Notes,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}

type EventListResponse struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	TextDate  *string   `json:"text_date"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (r *EventListRow) ToListResponse() EventListResponse {
	var textDate *string
	if r.Date != nil {
		textDate = ptr(r.Date.Format(textFormat))
	}

	return EventListResponse{
		ID:        r.ID,
		Title:     r.Title,
		TextDate:  textDate,
		Status:    r.Status,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
}

type PerformanceRow struct {
	Title       string    `db:"event_title"`
	Date        time.Time `db:"event_date"`
	TicketLink  string    `db:"ticket_link"`
	VenueID     int       `db:"venue_id"`
	ProgrammeID int       `db:"programme_id"`
}

// To be used for public API responses. Must not contain
// any internal information e.g. event_id
type PerformanceResponse struct {
	Title      string  `json:"title"`
	ExactDate  string  `json:"exact_date"`
	Date       string  `json:"date"`
	Venue      string  `json:"venue"`
	Programme  []Piece `json:"programme"`
	TicketLink string  `json:"ticket_link"`
}

func (r *PerformanceRow) ToResponse(
	venue string,
	programme string,
) PerformanceResponse {
	return PerformanceResponse{}
}

type ProgrammePieceRow struct {
	PieceID  int `db:"piece_id"`
	Sequence int `db:"sequence"`
}

type ProgrammePieceRequest struct {
	PieceID  int `json:"piece_id"`
	Sequence int `json:"sequence"`
}

type ProgrammePieceResponse struct {
	PieceID  int `json:"piece_id"`
	Sequence int `json:"sequence"`
}

func (r *ProgrammePieceRow) ToResponse() ProgrammePieceResponse {
	return ProgrammePieceResponse{
		PieceID:  r.PieceID,
		Sequence: r.Sequence,
	}
}

type ProgrammeRow struct {
	ID    int    `db:"programme_id"`
	Title string `db:"programme_title"`
}

type ProgrammeListRow struct {
	ID         int    `db:"programme_id"`
	Title      string `db:"programme_title"`
	PieceCount int    `db:"piece_count"`
	EventCount int    `db:"event_count"`
}

type ProgrammeRequest struct {
	Title  string                  `json:"title"`
	Pieces []ProgrammePieceRequest `json:"pieces"`
}

type ProgrammeFullResponse struct {
	ID     int                      `json:"id"`
	Title  string                   `json:"title"`
	Pieces []ProgrammePieceResponse `json:"pieces"`
}

type ProgrammeListResponse struct {
	ID         int    `json:"id"`
	Title      string `json:"title"`
	PieceCount int    `json:"piece_count"`
	EventCount int    `json:"event_count"`
}

func (r *ProgrammeRow) ToFullResponse(pieces []ProgrammePieceResponse) ProgrammeFullResponse {
	return ProgrammeFullResponse{
		ID:     r.ID,
		Title:  r.Title,
		Pieces: pieces,
	}
}

func (r *ProgrammeListRow) ToListResponse() ProgrammeListResponse {
	return ProgrammeListResponse{
		ID:         r.ID,
		Title:      r.Title,
		PieceCount: r.PieceCount,
		EventCount: r.EventCount,
	}
}

type PieceRow struct {
	ID           int    `db:"piece_id"`
	Title        string `db:"piece_title"`
	ComposerID   int    `db:"composer_id"`
	ComposerName string `db:"composer_name"`
}

type PieceRequest struct {
	Title      string `json:"title"`
	ComposerID int    `json:"composer_id"`
}

type PieceResponse struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	ComposerID   int    `json:"composer_id"`
	ComposerName string `json:"composer_name"`
}

// Used for public API JSON responses. Should not contain
// any internal information e.g. piece_id or composer_id.
type Piece struct {
	Composer string
	Title    string
}

func (r *PieceRow) ToResponse() PieceResponse {
	return PieceResponse{
		ID:           r.ID,
		Title:        r.Title,
		ComposerID:   r.ComposerID,
		ComposerName: r.ComposerName,
	}
}

type VenueRow struct {
	ID      int    `db:"venue_id"`
	Address string `db:"address"`
}

type VenueResponse struct {
	ID      int    `json:"venue_id"`
	Address string `json:"address"`
}

type VenueRequest struct {
	Address string `json:"address"`
}

func (r *VenueRow) ToResponse() VenueResponse {
	return VenueResponse{
		ID:      r.ID,
		Address: r.Address,
	}
}

type ComposerRow struct {
	ID        int    `db:"composer_id"`
	ShortName string `db:"short_name"`
	FullName  string `db:"full_name"`
}

type ComposerResponse struct {
	ID        int    `json:"id"`
	ShortName string `json:"short_name"`
	FullName  string `json:"full_name"`
}

type ComposerRequest struct {
	ShortName string `json:"short_name"`
	FullName  string `json:"full_name"`
}

func (r *ComposerRow) ToResponse() ComposerResponse {
	return ComposerResponse{
		ID:        r.ID,
		ShortName: r.ShortName,
		FullName:  r.FullName,
	}
}

// TODO: Video (id, title, extended title, embed URL)

// TODO: Contact details

// TODO: Biography
