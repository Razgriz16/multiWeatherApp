package fetch

import (
	"testing"

	"multiweather/model"
)

func f(x float64) *float64 { return &x }

func TestNormalizeReadings(t *testing.T) {
	var forecast forecastResponse
	forecast.Hourly.Time = []string{"2026-09-27T00:00", "2026-09-27T01:00"}
	forecast.Hourly.TemperatureECMWF = []*float64{f(12.8), f(12.6)}
	forecast.Hourly.TemperatureGFS = []*float64{f(13.0), f(12.9)}
	forecast.Hourly.TemperatureICON = []*float64{f(12.9), nil} // missing hour 1
	forecast.Hourly.HumidityECMWF = []*float64{f(60), f(62)}
	forecast.Hourly.HumidityGFS = []*float64{f(61), f(63)}
	forecast.Hourly.HumidityICON = []*float64{f(59), f(64)}
	forecast.Hourly.DewPointECMWF = []*float64{f(5), f(5)}
	forecast.Hourly.DewPointGFS = []*float64{f(5), f(5)}
	forecast.Hourly.DewPointICON = []*float64{f(5), f(5)}
	forecast.Hourly.PrecipitationECMWF = []*float64{f(0), f(0)}
	forecast.Hourly.PrecipitationGFS = []*float64{f(0), f(0)}
	forecast.Hourly.PrecipitationICON = []*float64{f(0), f(0)}
	forecast.Hourly.PrecipProbabilityECMWF = []*float64{f(0), f(0)}
	forecast.Hourly.PrecipProbabilityGFS = []*float64{f(0), f(0)}
	forecast.Hourly.PrecipProbabilityICON = []*float64{f(0), f(0)}
	forecast.Hourly.WindSpeedECMWF = []*float64{f(10), f(10)}
	forecast.Hourly.WindSpeedGFS = []*float64{f(10), f(10)}
	forecast.Hourly.WindSpeedICON = []*float64{f(10), f(10)}
	forecast.Hourly.WindDirectionECMWF = []*float64{f(180), f(180)}
	forecast.Hourly.WindDirectionGFS = []*float64{f(180), f(180)}
	forecast.Hourly.WindDirectionICON = []*float64{f(180), f(180)}

	code := func(c int64) *int64 { return &c }
	forecast.Hourly.WeatherCodeECMWF = []*int64{code(0), code(61)}
	forecast.Hourly.WeatherCodeGFS = []*int64{code(0), code(61)}
	forecast.Hourly.WeatherCodeICON = []*int64{code(0), nil} // missing hour 1
	// 61 (rain) deliberately unseeded, to exercise the "unknown code" path

	loc := model.Location{ID: 1}
	sourceIDs := map[string]int64{"ecmwf": 10, "gfs": 20, "icon": 30}
	conditionIDs := map[int64]int64{0: 100}

	readings, err := NormalizeReadings(loc, sourceIDs, conditionIDs, forecast, "issued", "fetched")
	if err != nil {
		t.Fatal(err)
	}
	if len(readings) != 6 { // 3 sources x 2 hours
		t.Fatalf("expected 6 readings, got %d", len(readings))
	}

	var iconHour1 *model.Reading
	for i := range readings {
		if readings[i].SourceID == 30 && readings[i].ValidTime == "2026-09-27T01:00" {
			iconHour1 = &readings[i]
		}
	}
	if iconHour1 == nil {
		t.Fatal("missing icon reading for hour 1")
	}
	if iconHour1.Temperature != nil {
		t.Errorf("expected nil temperature for the missing value, got %v", *iconHour1.Temperature)
	}
	if iconHour1.LocationID != 1 {
		t.Errorf("expected LocationID 1, got %d", iconHour1.LocationID)
	}
	if iconHour1.ConditionID != nil {
		t.Errorf("expected nil ConditionID for the missing weather_code, got %v", *iconHour1.ConditionID)
	}

	var ecmwfHour0 *model.Reading
	for i := range readings {
		if readings[i].SourceID == 10 && readings[i].ValidTime == "2026-09-27T00:00" {
			ecmwfHour0 = &readings[i]
		}
	}
	if ecmwfHour0 == nil || ecmwfHour0.ConditionID == nil || *ecmwfHour0.ConditionID != 100 {
		t.Fatalf("expected ConditionID 100 for weather_code 0, got %+v", ecmwfHour0)
	}

	var ecmwfHour1 *model.Reading
	for i := range readings {
		if readings[i].SourceID == 10 && readings[i].ValidTime == "2026-09-27T01:00" {
			ecmwfHour1 = &readings[i]
		}
	}
	if ecmwfHour1 == nil || ecmwfHour1.ConditionID != nil {
		t.Fatalf("expected nil ConditionID for unseeded weather_code 61, got %+v", ecmwfHour1)
	}
}

func TestNormalizeReadingsMissingSource(t *testing.T) {
	var forecast forecastResponse
	forecast.Hourly.Time = []string{"2026-09-27T00:00"}

	_, err := NormalizeReadings(model.Location{}, map[string]int64{}, map[int64]int64{}, forecast, "issued", "fetched")
	if err == nil {
		t.Fatal("expected an error when sourceIDs has no entries")
	}
}
