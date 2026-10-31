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
			wind_speed, wind_direction, condition_id
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
			r.WindSpeed, r.WindDirection, r.ConditionID,
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

// ModelReading is one source's temperature for one hour — the raw input
// behind an aggregate row, kept separate from ForecastHour since this is
// per-source, not rolled up.
type ModelReading struct {
	Source      string
	ValidTime   string
	Temperature *float64
}

// ModelReadingsForDay returns every source's hourly temperature for one
// location's calendar day, from the most recently issued ingest run —
// same "most recent" rule as LatestForecast, scoped to a single date.
func ModelReadingsForDay(db *sql.DB, locationID int64, date string) ([]ModelReading, error) {
	rows, err := db.Query(
		`SELECT s.name, r.valid_time, r.temperature
		 FROM readings r
		 JOIN sources s ON s.id = r.source_id
		 WHERE r.location_id = ?
		   AND r.valid_time LIKE ?
		   AND r.issued_at = (SELECT MAX(issued_at) FROM readings WHERE location_id = ?)
		 ORDER BY r.valid_time, s.name`,
		locationID, date+"%", locationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readings []ModelReading
	for rows.Next() {
		var r ModelReading
		if err := rows.Scan(&r.Source, &r.ValidTime, &r.Temperature); err != nil {
			return nil, err
		}
		readings = append(readings, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return readings, nil
}
