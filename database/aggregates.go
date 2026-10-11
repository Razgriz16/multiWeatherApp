package database

import (
	"database/sql"

	"multiweather/model"
)

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
