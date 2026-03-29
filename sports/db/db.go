package db

import (
	"time"

	"syreclabs.com/go/faker"
)

// seed will prepare our database with some dummy data.
func (r *sportsRepo) seed() error {
	statement, err := r.db.Prepare(`CREATE TABLE IF NOT EXISTS sports (id INTEGER PRIMARY KEY, name TEXT NOT NULL UNIQUE)`)
	if err == nil {
		_, err = statement.Exec()
	}

	if err == nil {
		statement, err = r.db.Prepare(`CREATE TABLE IF NOT EXISTS events (id INTEGER PRIMARY KEY, sport_id INTEGER NOT NULL, visible INTEGER, advertised_start_time DATETIME, status TEXT, FOREIGN KEY(sport_id) REFERENCES sports(id))`)
		if err == nil {
			_, err = statement.Exec()
		}
	}

	for i := 1; i <= 10; i++ {
		statement, err = r.db.Prepare(`INSERT OR IGNORE INTO sports(id, name) VALUES (?,?)`)
		if err == nil {
			_, err = statement.Exec(
				i,
				faker.Team().Name(),
			)
		}
	}

	for i := 1; i <= 100; i++ {
		statement, err = r.db.Prepare(`INSERT OR IGNORE INTO events(id, sport_id, visible, advertised_start_time, status) VALUES (?,?,?,?,?)`)
		if err == nil {
			_, err = statement.Exec(
				i,
				faker.Number().Between(1, 10),
				faker.Number().Between(0, 1),
				faker.Time().Between(time.Now().AddDate(0, 0, -1), time.Now().AddDate(0, 0, 2)).Format(time.RFC3339),
				"OPEN",
			)
		}
	}

	return err
}
