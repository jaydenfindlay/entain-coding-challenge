package db

const (
	sportsList         = "list"
	sportsUpdateStatus = "update_status"
)

// getSportQueries returns a map of queries for the sports repository.
func getSportQueries() map[string]string {
	return map[string]string{
		sportsList: `
			SELECT
				s.id,
				s.name,
				e.id AS event_id,
				e.visible,
				e.advertised_start_time,
				e.status
			FROM events e
			INNER JOIN sports s ON s.id = e.sport_id
		`,
	}
}

// updateStatusQueries checks the advertised start time of events and updates their status to 'CLOSED' if the start time has passed, otherwise it sets the status to 'OPEN'.
func updateStatusQueries() map[string]string {
	return map[string]string{
		sportsUpdateStatus: `
			UPDATE events
			SET status = CASE
				WHEN advertised_start_time < datetime('now') THEN 'CLOSED'
				ELSE 'OPEN'
			END
		`,
	}
}
