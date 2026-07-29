package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"

	"github.com/ametis70/hellbot/internal/domain"
)

const schema = `
CREATE TABLE IF NOT EXISTS campaign (
	id      INTEGER PRIMARY KEY CHECK (id = 1),
	payload TEXT    NOT NULL
);

CREATE TABLE IF NOT EXISTS ongoing_events (
	id   INTEGER NOT NULL,
	kind TEXT    NOT NULL,
	PRIMARY KEY (id, kind)
);
`

// Store implements port.CampaignStore and port.EventStore using a SQLite database.
type Store struct {
	db *sql.DB
}

// Options holds the configuration for the SQLite store.
type Options struct {
	// Path is the file path for the SQLite database (e.g. "./hellbot.db").
	// Use ":memory:" for an in-process, non-persistent database.
	Path string
}

// New opens (or creates) the SQLite database at opts.Path and applies the schema.
func New(opts Options) (*Store, error) {
	db, err := sql.Open("sqlite", opts.Path)
	if err != nil {
		return nil, fmt.Errorf("sqlite: open %q: %w", opts.Path, err)
	}

	// Single writer to avoid SQLITE_BUSY on concurrent access.
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("sqlite: apply schema: %w", err)
	}

	return &Store{db: db}, nil
}

// Close releases the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}

// ── CampaignStore ────────────────────────────────────────────────────────────

func (s *Store) SaveCampaign(c *domain.CampaignStatus) error {
	data, err := json.Marshal(c)
	if err != nil {
		return fmt.Errorf("sqlite: marshal campaign: %w", err)
	}
	_, err = s.db.Exec(
		`INSERT INTO campaign (id, payload) VALUES (1, ?) ON CONFLICT(id) DO UPDATE SET payload = excluded.payload`,
		string(data),
	)
	if err != nil {
		return fmt.Errorf("sqlite: save campaign: %w", err)
	}
	return nil
}

func (s *Store) LatestCampaign() (*domain.CampaignStatus, error) {
	var payload string
	err := s.db.QueryRow(`SELECT payload FROM campaign WHERE id = 1`).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("no campaign stored")
	}
	if err != nil {
		return nil, fmt.Errorf("sqlite: get campaign: %w", err)
	}

	var c domain.CampaignStatus
	if err := json.Unmarshal([]byte(payload), &c); err != nil {
		return nil, fmt.Errorf("sqlite: unmarshal campaign: %w", err)
	}
	return &c, nil
}

// ── EventStore ───────────────────────────────────────────────────────────────

func (s *Store) SaveOngoingEvent(id int, kind domain.EventKind) error {
	_, err := s.db.Exec(
		`INSERT OR IGNORE INTO ongoing_events (id, kind) VALUES (?, ?)`,
		id, string(kind),
	)
	if err != nil {
		return fmt.Errorf("sqlite: save event: %w", err)
	}
	return nil
}

func (s *Store) RemoveOngoingEvent(id int, kind domain.EventKind) error {
	res, err := s.db.Exec(
		`DELETE FROM ongoing_events WHERE id = ? AND kind = ?`,
		id, string(kind),
	)
	if err != nil {
		return fmt.Errorf("sqlite: remove event: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("sqlite: remove event rows affected: %w", err)
	}
	if n == 0 {
		return errors.New("event not found")
	}
	return nil
}

func (s *Store) GetOngoingEvent(id int, kind domain.EventKind) (*domain.OngoingEvent, error) {
	var evID int
	var evKind string
	err := s.db.QueryRow(
		`SELECT id, kind FROM ongoing_events WHERE id = ? AND kind = ?`,
		id, string(kind),
	).Scan(&evID, &evKind)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("event not found")
	}
	if err != nil {
		return nil, fmt.Errorf("sqlite: get event: %w", err)
	}
	return &domain.OngoingEvent{ID: evID, Kind: domain.EventKind(evKind)}, nil
}

func (s *Store) ListOngoingEvents(kind domain.EventKind) (_ []*domain.OngoingEvent, err error) {
	rows, err := s.db.Query(
		`SELECT id, kind FROM ongoing_events WHERE kind = ?`,
		string(kind),
	)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list events: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("sqlite: close rows: %w", cerr)
		}
	}()

	result := make([]*domain.OngoingEvent, 0)
	for rows.Next() {
		var evID int
		var evKind string
		if err := rows.Scan(&evID, &evKind); err != nil {
			return nil, fmt.Errorf("sqlite: scan event: %w", err)
		}
		result = append(result, &domain.OngoingEvent{ID: evID, Kind: domain.EventKind(evKind)})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: list events rows: %w", err)
	}
	return result, nil
}
