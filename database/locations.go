package database

import (
	"database/sql"
	"errors"
	"math"
	"time"

	"multiweather/model"
)

func round4(x float64) float64 {
	return math.Round(x*1e4) / 1e4
}

// inserted is false with a nil error when the coordinates are already tracked.
func AddLocation(db *sql.DB, p model.Location) (inserted bool, err error) {
	if p.Name == "" || p.Timezone == "" {
		return false, errors.New("nothing to insert: location is empty")
	}

	res, err := db.Exec(
		`INSERT INTO locations (name, latitude, longitude, timezone, created_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT (latitude, longitude) DO NOTHING`,
		p.Name, round4(p.Latitude), round4(p.Longitude), p.Timezone,
		time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return false, err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n == 1, nil
}
