package database

import (
	"path/filepath"
	"testing"

	"multiweather/model"
)

func TestInsertReadings(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := AddLocation(db, model.Location{
		Name: "Berlin", Latitude: 52.52, Longitude: 13.41, Timezone: "Europe/Berlin",
	}); err != nil {
		t.Fatal(err)
	}
	sourceIDs, err := SourceIDs(db)
	if err != nil {
		t.Fatal(err)
	}

	temp := 12.5
	readings := []model.Reading{
		{SourceID: sourceIDs["ecmwf"], LocationID: 1, ValidTime: "2026-09-27T00:00", IssuedAt: "a", FetchedAt: "b", Temperature: &temp},
		{SourceID: sourceIDs["ecmwf"], LocationID: 1, ValidTime: "2026-09-27T01:00", IssuedAt: "a", FetchedAt: "b"}, // Temperature nil
	}

	n, err := InsertReadings(db, readings)
	if err != nil || n != 2 {
		t.Fatalf("first insert: n=%d err=%v", n, err)
	}

	// same rows again: should be skipped as duplicates, not double-counted
	n, err = InsertReadings(db, readings)
	if err != nil || n != 0 {
		t.Fatalf("duplicate insert: n=%d err=%v", n, err)
	}

	var gotTemp *float64
	row := db.QueryRow(`SELECT temperature FROM readings WHERE valid_time = '2026-09-27T01:00'`)
	if err := row.Scan(&gotTemp); err != nil {
		t.Fatal(err)
	}
	if gotTemp != nil {
		t.Errorf("expected NULL temperature, got %v", *gotTemp)
	}
}
