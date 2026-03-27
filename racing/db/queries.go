package db

const (
	racesList         = "list"
	racesUpdateStatus = "update_status"
)

func getRaceQueries() map[string]string {
	return map[string]string{
		racesList: `
			SELECT 
				id, 
				meeting_id, 
				name, 
				number, 
				visible, 
				advertised_start_time,
				status
			FROM races
		`,
	}
}

func updateStatusQueries() map[string]string {
	return map[string]string{
		racesUpdateStatus: `
			UPDATE races
			SET status = CASE
				WHEN advertised_start_time < datetime('now') THEN 'CLOSED'
				ELSE 'OPEN'
			END
		`,
	}
}
