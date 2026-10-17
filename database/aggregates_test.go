package database

import (
	"path/filepath"
	"testing"

	"multiweather/model"
)

func TestLatestForecast(t *testing.T) {
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

	clearID, err := ConditionIDs(db)
	if err != nil {
		t.Fatal(err)
	}

	older := []model.Aggregate{
		{LocationID: 1, ValidTime: "2026-09-27T00:00", IssuedAt: "run1", Temperature: new(5.0), Humidity: new(80.0), SourceCount: 3, ComputedAt: "c1"},
	}
	newer := []model.Aggregate{
		{LocationID: 1, ValidTime: "2026-09-27T00:00", IssuedAt: "run2", Temperature: new(12.5), Humidity: new(60.0), ConditionID: new(clearID[0]), SourceCount: 3, ComputedAt: "c2"},
	}
	if _, err := InsertAggregates(db, older); err != nil {
		t.Fatal(err)
	}
	if _, err := InsertAggregates(db, newer); err != nil {
		t.Fatal(err)
	}

	got, err := LatestForecast(db, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 hour from the latest run only, got %d", len(got))
	}
	if got[0].Temperature == nil || *got[0].Temperature != 12.5 {
		t.Errorf("expected the newer run's temperature 12.5, got %v", got[0].Temperature)
	}
	if got[0].Condition == nil || *got[0].Condition != "Clear sky" {
		t.Errorf("expected condition description 'Clear sky', got %v", got[0].Condition)
	}
}
