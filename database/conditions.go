package database

import "database/sql"

// ConditionIDs returns every conditions row as WMO code -> id.
func ConditionIDs(db *sql.DB) (map[int64]int64, error) {
	rows, err := db.Query(`SELECT id, code FROM conditions`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := map[int64]int64{}
	for rows.Next() {
		var id, code int64
		if err := rows.Scan(&id, &code); err != nil {
			return nil, err
		}
		ids[code] = id
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}
