package database

import (
	"database/sql"

	"multiweather/model"
)

// InsertReadings writes every reading in one transaction and returns how
// many rows were actually new (duplicates are silently skipped).
func InsertReadings(db *sql.DB, readings []model.Reading) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() // no-op once committed

	stmt, err := tx.Prepare(
		`INSERT INTO readings (
			source_id, location_id, valid_time, issued_at, fetched_at,
			temperature, humidity, dew_point,
			precipitation_mm, precipitation_probability,
			wind_speed, wind_direction
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (source_id, location_id, valid_time, issued_at) DO NOTHING`,
	)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var inserted int64
	for _, r := range readings {
		res, err := stmt.Exec(
			r.SourceID, r.LocationID, r.ValidTime, r.IssuedAt, r.FetchedAt,
			r.Temperature, r.Humidity, r.DewPoint,
			r.PrecipitationMM, r.PrecipitationProbability,
			r.WindSpeed, r.WindDirection,
		)
		if err != nil {
			return inserted, err
		}
		n, err := res.RowsAffected()
		if err != nil {
			return inserted, err
		}
		inserted += n
	}

	if err := tx.Commit(); err != nil {
		return inserted, err
	}
	return inserted, nil
}
