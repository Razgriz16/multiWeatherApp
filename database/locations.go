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

func GetActiveLocations(db *sql.DB) ([]model.Location, error) {
	rows, err := db.Query(
		`SELECT id, name, latitude, longitude, timezone, active, notify, created_at
		 FROM locations WHERE active = 1`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []model.Location
	for rows.Next() {
		var loc model.Location
		if err := rows.Scan(
			&loc.ID, &loc.Name, &loc.Latitude, &loc.Longitude,
			&loc.Timezone, &loc.Active, &loc.Notify, &loc.CreatedAt,
		); err != nil {
			return nil, err
		}
		locations = append(locations, loc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return locations, nil
}
