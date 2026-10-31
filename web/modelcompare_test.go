package web

import (
	"testing"

	"multiweather/database"
)

func TestBuildHourlyChart_AlignsLabelsTempsAndPrecip(t *testing.T) {
	hours := []database.ForecastHour{
		{ValidTime: "2026-10-17T00:00", Temperature: new(10.0), PrecipitationMM: new(0.0)},
		{ValidTime: "2026-10-17T01:00", Temperature: new(11.0)}, // no precipitation reading
	}

	data := buildHourlyChart(hours)

	if len(data.Labels) != 2 || data.Labels[0] != "00:00" || data.Labels[1] != "01:00" {
		t.Fatalf("unexpected labels: %+v", data.Labels)
	}
	if data.Temps[0] == nil || *data.Temps[0] != 10.0 || data.Temps[1] == nil || *data.Temps[1] != 11.0 {
		t.Errorf("unexpected temps: %+v", data.Temps)
	}
	if data.Precip[0] == nil || *data.Precip[0] != 0.0 {
		t.Errorf("expected first hour's precipitation to be 0.0, got %+v", data.Precip[0])
	}
	if data.Precip[1] != nil {
		t.Errorf("expected nil for the missing precipitation reading, got %v", *data.Precip[1])
	}
}

func TestBuildModelComparison_AlignsSourcesToAggregateHours(t *testing.T) {
	aggregateHours := []database.ForecastHour{
		{ValidTime: "2026-10-17T00:00", Temperature: new(10.0)},
		{ValidTime: "2026-10-17T01:00", Temperature: new(11.0)},
	}
	modelReadings := []database.ModelReading{
		{Source: "gfs", ValidTime: "2026-10-17T00:00", Temperature: new(9.0)},
		{Source: "gfs", ValidTime: "2026-10-17T01:00", Temperature: new(10.0)},
		{Source: "ecmwf", ValidTime: "2026-10-17T00:00", Temperature: new(11.0)},
		// ecmwf has no 01:00 reading — should come through as nil, not 0.
	}

	data := buildModelComparison(aggregateHours, modelReadings)

	if len(data.Labels) != 2 || data.Labels[0] != "00:00" || data.Labels[1] != "01:00" {
		t.Fatalf("unexpected labels: %+v", data.Labels)
	}
	if len(data.Series) != 3 {
		t.Fatalf("expected 3 series (aggregate + 2 sources), got %d", len(data.Series))
	}

	if data.Series[0].Name != "Aggregate" || *data.Series[0].Temps[0] != 10.0 {
		t.Errorf("expected aggregate series first, got %+v", data.Series[0])
	}

	// Sources come after the aggregate, alphabetically.
	if data.Series[1].Name != "ecmwf" || *data.Series[1].Temps[0] != 11.0 {
		t.Errorf("unexpected ecmwf series: %+v", data.Series[1])
	}
	if data.Series[1].Temps[1] != nil {
		t.Errorf("expected nil for ecmwf's missing 01:00 reading, got %v", *data.Series[1].Temps[1])
	}

	if data.Series[2].Name != "gfs" || *data.Series[2].Temps[1] != 10.0 {
		t.Errorf("unexpected gfs series: %+v", data.Series[2])
	}
}

func TestBuildModelComparison_NoReadingsStillReturnsAggregate(t *testing.T) {
	aggregateHours := []database.ForecastHour{
		{ValidTime: "2026-10-17T00:00", Temperature: new(10.0)},
	}

	data := buildModelComparison(aggregateHours, nil)

	if len(data.Series) != 1 || data.Series[0].Name != "Aggregate" {
		t.Errorf("expected only the aggregate series, got %+v", data.Series)
	}
}
