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

	loc := model.Location{ID: 1}
	sourceIDs := map[string]int64{"ecmwf": 10, "gfs": 20, "icon": 30}

	readings, err := NormalizeReadings(loc, sourceIDs, forecast, "issued", "fetched")
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
}

func TestNormalizeReadingsMissingSource(t *testing.T) {
	var forecast forecastResponse
	forecast.Hourly.Time = []string{"2026-09-27T00:00"}

	_, err := NormalizeReadings(model.Location{}, map[string]int64{}, forecast, "issued", "fetched")
	if err == nil {
		t.Fatal("expected an error when sourceIDs has no entries")
	}
}
