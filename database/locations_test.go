package database

import (
	"path/filepath"
	"testing"

	"multiweather/model"
)

func TestGetActiveLocations(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	for _, name := range []string{"Berlin", "Santiago"} {
		p := model.Location{Name: name, Latitude: 1, Longitude: 1, Timezone: "UTC"}
		if name == "Santiago" {
			p.Longitude = 2
		}
		if _, err := AddLocation(db, p); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`UPDATE locations SET active = 0 WHERE name = 'Berlin'`); err != nil {
		t.Fatal(err)
	}

	got, err := GetActiveLocations(db)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "Santiago" {
		t.Fatalf("expected only Santiago, got %+v", got)
	}
}
