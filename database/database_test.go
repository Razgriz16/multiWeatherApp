package database

import (
	"path/filepath"
	"testing"

	"multiweather/model"
)

func TestOpenCreatesSchema(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "sub", "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	for _, table := range []string{"sources", "conditions", "locations", "readings", "aggregates"} {
		var n int
		if err := db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n != 1 {
			t.Errorf("table %s missing", table)
		}
	}
}

func TestAddLocation(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	p := model.Location{Name: "Santiago", Latitude: -33.456941, Longitude: -70.648273, Timezone: "America/Santiago"}

	inserted, err := AddLocation(db, p)
	if err != nil || !inserted {
		t.Fatalf("first insert: inserted=%v err=%v", inserted, err)
	}

	// same point after rounding to 4 decimals must count as a duplicate
	p.Latitude, p.Longitude = -33.456939, -70.648274
	inserted, err = AddLocation(db, p)
	if err != nil || inserted {
		t.Fatalf("duplicate insert: inserted=%v err=%v", inserted, err)
	}

	if _, err := AddLocation(db, model.Location{}); err == nil {
		t.Error("empty place should return an error")
	}
}
