package db

import (
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	_ "github.com/mattn/go-sqlite3"

	"git.neds.sh/matty/entain/racing/proto/racing"
)

// RacesRepo provides repository access to races.
type RacesRepo interface {
	// Init will initialise our races repository.
	Init() error

	// updateStatus will update the status of races based on their advertised start time.
	updateStatus() (bool, error)

	// List will return a list of races.
	List(filter *racing.ListRacesRequestFilter) ([]*racing.Race, error)
}

type racesRepo struct {
	db   *sql.DB
	init sync.Once
}

// NewRacesRepo creates a new races repository.
func NewRacesRepo(db *sql.DB) RacesRepo {
	return &racesRepo{db: db}
}

// Init prepares the race repository dummy data.
func (r *racesRepo) Init() error {
	var err error

	r.init.Do(func() {
		// For test/example purposes, we seed the DB with some dummy races.
		err = r.seed()
	})

	_, err = r.updateStatus()
	if err != nil {
		return err
	}

	return err
}

// updateStatus will update the status of races based on their advertised start time. If the advertised start time is in the past, the status will be set to "CLOSED", otherwise it will be set to "OPEN".
func (r *racesRepo) updateStatus() (bool, error) {

	_, err := r.db.Exec(updateStatusQueries()[racesUpdateStatus])
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *racesRepo) List(filter *racing.ListRacesRequestFilter) ([]*racing.Race, error) {
	var (
		err   error
		query string
		args  []interface{}
	)

	query = getRaceQueries()[racesList]

	query, args = r.applyFilter(query, filter)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}

	_, err = r.updateStatus()
	if err != nil {
		return nil, err
	}

	return r.scanRaces(rows)
}

func (r *racesRepo) applyFilter(query string, filter *racing.ListRacesRequestFilter) (string, []interface{}) {
	var (
		clauses []string
		args    []interface{}
	)

	if filter == nil {
		return query, args
	}
	// Visible is a pointer to a bool, so we can check if it's nil (not set) or not.
	if filter.Visible != nil {
		clauses = append(clauses, "visible = ?")
		args = append(args, *filter.Visible)
	}

	if len(filter.MeetingIds) > 0 {
		clauses = append(clauses, "meeting_id IN ("+strings.Repeat("?,", len(filter.MeetingIds)-1)+"?)")

		for _, meetingID := range filter.MeetingIds {
			args = append(args, meetingID)
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
			orderBy = " ORDER BY advertised_start_time ASC"
		} else {
			orderBy = " ORDER BY advertised_start_time DESC"
		}
	}

	query += orderBy

	return query, args
}

func (m *racesRepo) scanRaces(
	rows *sql.Rows,
) ([]*racing.Race, error) {
	var races []*racing.Race

	for rows.Next() {
		var race racing.Race
		var advertisedStart time.Time

		if err := rows.Scan(&race.Id, &race.MeetingId, &race.Name, &race.Number, &race.Visible, &advertisedStart); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}

			return nil, err
		}

		ts, err := ptypes.TimestampProto(advertisedStart)
		if err != nil {
			return nil, err
		}

		race.AdvertisedStartTime = ts

		races = append(races, &race)
	}

	return races, nil
}
