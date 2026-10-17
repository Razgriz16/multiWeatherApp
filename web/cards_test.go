package web

import (
	"testing"

	"multiweather/database"
)

func hour(validTime string, temp float64, code int64, desc string) database.ForecastHour {
	return database.ForecastHour{
		ValidTime:                validTime,
		Temperature:              new(temp),
		PrecipitationProbability: new(20.0),
		ConditionCode:            new(code),
		Condition:                new(desc),
	}
}

func TestRollupDaily_GroupsByDateAndTracksMinMax(t *testing.T) {
	hours := []database.ForecastHour{
		hour("2026-10-12T00:00", 10, 0, "Clear sky"),
		hour("2026-10-12T12:00", 20, 0, "Clear sky"),
		hour("2026-10-13T00:00", 5, 61, "Slight rain"),
	}

	days := rollupDaily(hours)
	if len(days) != 2 {
		t.Fatalf("expected 2 days, got %d", len(days))
	}

	if days[0].Date != "2026-10-12" || days[0].MaxTemp != "20°C" || days[0].MinTemp != "10°C" {
		t.Errorf("unexpected day 0: %+v", days[0])
	}
	if days[0].Label != "Today" {
		t.Errorf("expected day 0 labeled Today, got %q", days[0].Label)
	}
	if days[1].Label != "Tomorrow" {
		t.Errorf("expected day 1 labeled Tomorrow, got %q", days[1].Label)
	}
}

func TestRollupDaily_AveragesHumidityAndTotalsPrecipitation(t *testing.T) {
	hours := []database.ForecastHour{
		{
			ValidTime: "2026-10-12T00:00", Temperature: new(10.0),
			Humidity: new(80.0), PrecipitationMM: new(1.0),
			ConditionCode: new(int64(0)), Condition: new("Clear sky"),
		},
		{
			ValidTime: "2026-10-12T12:00", Temperature: new(20.0),
			Humidity: new(60.0), PrecipitationMM: new(2.0),
			ConditionCode: new(int64(0)), Condition: new("Clear sky"),
		},
	}

	days := rollupDaily(hours)
	if days[0].Humidity != "70%" {
		t.Errorf("expected average humidity 70%%, got %q", days[0].Humidity)
	}
	if days[0].PrecipitationMM != "3.0mm" {
		t.Errorf("expected total precipitation 3.0mm, got %q", days[0].PrecipitationMM)
	}
}

func TestRollupDaily_MissingPrecipitationShowsNoData(t *testing.T) {
	hours := []database.ForecastHour{
		hour("2026-10-12T00:00", 10, 0, "Clear sky"),
	}

	days := rollupDaily(hours)
	if days[0].PrecipitationMM != "—" {
		t.Errorf("expected no-data dash for missing precipitation, got %q", days[0].PrecipitationMM)
	}
}

func TestRollupDaily_ConditionTieBreaksBySeverity(t *testing.T) {
	hours := []database.ForecastHour{
		hour("2026-10-12T00:00", 10, 0, "Clear sky"),
		hour("2026-10-12T12:00", 20, 95, "Thunderstorm"),
	}

	days := rollupDaily(hours)
	if days[0].Condition != "Thunderstorm" {
		t.Errorf("expected tie broken toward the higher code (95), got %q", days[0].Condition)
	}
}

func TestSelectHero_DefaultsToTodayAndTomorrow(t *testing.T) {
	days := []dayCard{{Date: "2026-10-12"}, {Date: "2026-10-13"}, {Date: "2026-10-14"}}

	hero := selectHero(days, "")
	if len(hero) != 2 || hero[0].Date != "2026-10-12" || hero[1].Date != "2026-10-13" {
		t.Errorf("expected today+tomorrow, got %+v", hero)
	}
}

func TestSelectHero_SingleDayWhenSelected(t *testing.T) {
	days := []dayCard{{Date: "2026-10-12"}, {Date: "2026-10-13"}, {Date: "2026-10-14"}}

	hero := selectHero(days, "2026-10-14")
	if len(hero) != 1 || hero[0].Date != "2026-10-14" {
		t.Errorf("expected single selected day, got %+v", hero)
	}
}

func TestSelectHero_FallsBackWhenSelectedDayNotFound(t *testing.T) {
	days := []dayCard{{Date: "2026-10-12"}, {Date: "2026-10-13"}}

	hero := selectHero(days, "2099-01-01")
	if len(hero) != 2 {
		t.Errorf("expected fallback to today+tomorrow, got %+v", hero)
	}
}

func TestMarkSelected_FlagsOnlyTheMatchingDay(t *testing.T) {
	days := []dayCard{{Date: "2026-10-12"}, {Date: "2026-10-13"}}

	marked := markSelected(days, "2026-10-13")
	if marked[0].Selected {
		t.Error("expected day 0 unselected")
	}
	if !marked[1].Selected {
		t.Error("expected day 1 selected")
	}
}
