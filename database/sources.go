package database

import "database/sql"

// SourceIDs returns every sources row as name -> id.
func SourceIDs(db *sql.DB) (map[string]int64, error) {
	rows, err := db.Query(`SELECT id, name FROM sources`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := map[string]int64{}
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		ids[name] = id
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}

// APIModels returns every source's api_model, ordered by id — the
// list to pass as Open-Meteo's models= parameter.
func APIModels(db *sql.DB) ([]string, error) {
	rows, err := db.Query(`SELECT api_model FROM sources ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []string
	for rows.Next() {
		var m string
		if err := rows.Scan(&m); err != nil {
			return nil, err
		}
		models = append(models, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return models, nil
}
