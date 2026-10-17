package database

import (
	"database/sql"

	"multiweather/model"
)

// ForecastHour is one aggregate row joined with its human-readable
// condition — shaped for display, not for storage (hence it isn't
// model.Aggregate).
type ForecastHour struct {
	ValidTime                string
	Temperature              *float64
	Humidity                 *float64
	PrecipitationMM          *float64
	PrecipitationProbability *float64
	ConditionCode            *int64
	Condition                *string
}

// LatestForecast returns every hour from the most recently issued ingest
// run for a location, ordered by valid_time. "Most recent" is a
// MAX(issued_at) query, not a stored flag (docs/SCHEMA-ANALYSIS.md).
func LatestForecast(db *sql.DB, locationID int64) ([]ForecastHour, error) {
	rows, err := db.Query(
		`SELECT a.valid_time, a.temperature, a.humidity,
		        a.precipitation_mm, a.precipitation_probability, c.code, c.description
		 FROM aggregates a
		 LEFT JOIN conditions c ON c.id = a.condition_id
		 WHERE a.location_id = ?
		   AND a.issued_at = (SELECT MAX(issued_at) FROM aggregates WHERE location_id = ?)
		 ORDER BY a.valid_time`,
		locationID, locationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hours []ForecastHour
	for rows.Next() {
		var h ForecastHour
		if err := rows.Scan(
			&h.ValidTime, &h.Temperature, &h.Humidity,
			&h.PrecipitationMM, &h.PrecipitationProbability, &h.ConditionCode, &h.Condition,
		); err != nil {
			return nil, err
		}
		hours = append(hours, h)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return hours, nil
}

// InsertAggregates writes every aggregate in one transaction and returns
// how many rows were actually new (duplicates are silently skipped).
func InsertAggregates(db *sql.DB, aggregates []model.Aggregate) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() // no-op once committed

	stmt, err := tx.Prepare(
		`INSERT INTO aggregates (
			location_id, valid_time, issued_at,
			temperature, humidity, dew_point,
			precipitation_mm, precipitation_probability,
			condition_id, source_count, computed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (location_id, valid_time, issued_at) DO NOTHING`,
	)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var inserted int64
	for _, a := range aggregates {
		res, err := stmt.Exec(
			a.LocationID, a.ValidTime, a.IssuedAt,
			a.Temperature, a.Humidity, a.DewPoint,
			a.PrecipitationMM, a.PrecipitationProbability,
			a.ConditionID, a.SourceCount, a.ComputedAt,
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
