package db

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	_ "github.com/mattn/go-sqlite3"

	"git.neds.sh/matty/entain/sports/proto/sports"
)

// SportsRepo provides repository access to sports.
type SportsRepo interface {
	// Init will initialise our sports repository.
	Init() error

	// List will return a list of sports.
	List(filter *sports.ListEventsRequestFilter) ([]*sports.Sport, error)

	// GetEvent will return a single event based on the provided filter.
	GetEvent(filter *sports.GetEventRequestFilter) (*sports.Sport, error)
}

// sportsRepo is a concrete implementation of the SportsRepo interface.
type sportsRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewSportsRepo creates a new sports repository.
func NewSportsRepo(db *sql.DB) SportsRepo {
	return &sportsRepo{db: db}
}

// Init prepares the sports repository dummy data.
func (r *sportsRepo) Init() error {
	var err error

	if _, err = r.db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return err
	}

	r.init.Do(func() {
		// For test/example purposes, we seed the DB with some dummy sports.
		err = r.seed()
	})
	// We update the status of sports based on their advertised start time every time we initialise the repository.
	_, err = r.updateStatus()
	if err != nil {
		return err
	}

	return err
}

// updateStatus will update the status of sports based on their advertised start time.
func (r *sportsRepo) updateStatus() (bool, error) {

	_, err := r.db.Exec(updateStatusQueries()[sportsUpdateStatus])
	if err != nil {
		return false, err
	}
	return true, nil
}

// List will return a list of events.
func (r *sportsRepo) List(filter *sports.ListEventsRequestFilter) ([]*sports.Sport, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	query = getSportQueries()[sportsList]

	query, args = r.applyFilter(query, filter)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	_, err = r.updateStatus()
	if err != nil {
		return nil, err
	}

	return r.scanSports(rows)
}

// GetEvent will return a single event based on the provided filter.
func (r *sportsRepo) GetEvent(filter *sports.GetEventRequestFilter) (*sports.Sport, error) {
	query := getSportQueries()[sportsList] + " WHERE e.id = ?"

	rows, err := r.db.Query(query, filter.EventId)
	if err != nil {
		return nil, err
	}

	_, err = r.updateStatus()
	if err != nil {
		return nil, err
	}

	sportResults, err := r.scanSports(rows)
	if err != nil {
		return nil, err
	}
	if len(sportResults) == 0 {
		return nil, nil
	}
	return sportResults[0], nil
}

// applyFilter will apply the provided filter to the query and return the modified query and its arguments.
func (r *sportsRepo) applyFilter(query string, filter *sports.ListEventsRequestFilter) (string, []interface{}) {
	var (
		clauses []string
		args    []interface{}
	)

	if filter == nil {
		return query, args
	}

	if filter.Visible != nil {
		clauses = append(clauses, "e.visible = ?")
		args = append(args, *filter.Visible)
	}

	if len(filter.EventIds) > 0 {
		clauses = append(clauses, "e.id IN ("+strings.Repeat("?,", len(filter.EventIds)-1)+"?)")

		for _, eventID := range filter.EventIds {
			args = append(args, eventID)
		}
	}

	if len(clauses) != 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}

	// AdvertisedStartTimeOrdering is a pointer to a bool, so we can check if it's nil (not set) or not.
	// If it's set, we order by advertised_start_time in the specified order. If it's not set, we don't apply any ordering.
	var orderBy string
	if filter.AdvertisedStartTimeOrdering != nil {
		if *filter.AdvertisedStartTimeOrdering {
			orderBy = " ORDER BY e.advertised_start_time ASC"
		} else {
			orderBy = " ORDER BY e.advertised_start_time DESC"
		}
	}

	query += orderBy

	return query, args
}

// scanSports will scan the provided rows and return a list of sports.
func (r *sportsRepo) scanSports(rows *sql.Rows) ([]*sports.Sport, error) {
	var results []*sports.Sport

	for rows.Next() {
		var sport sports.Sport
		var advertisedStart time.Time

		if err := rows.Scan(&sport.Id, &sport.Name, &sport.EventId, &sport.Visible, &advertisedStart, &sport.Status); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, err
		}

		ts, err := ptypes.TimestampProto(advertisedStart)
		if err != nil {
			return nil, err
		}

		sport.AdvertisedStartTime = ts

		results = append(results, &sport)
	}

	return results, nil
}
