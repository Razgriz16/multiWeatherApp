package fetch

import (
	"fmt"
	"multiweather/model"
)

// one view per model, so the per-hour loop below is written once
// instead of once per model.
type modelReadings struct {
	sourceName  string
	temperature []*float64
	humidity    []*float64
	dewPoint    []*float64
	precip      []*float64
	precipProb  []*float64
	windSpeed   []*float64
	windDir     []*float64
}

// NormalizeReadings turns one forecastResponse into one model.Reading per
// (source, hour). sourceIDs maps a short source name ("ecmwf", "gfs",
// "icon") to its row id in the sources table.
func NormalizeReadings(
	loc model.Location,
	sourceIDs map[string]int64,
	forecast forecastResponse,
	issuedAt, fetchedAt string,
) ([]model.Reading, error) {
	h := forecast.Hourly
	views := []modelReadings{
		{"ecmwf", h.TemperatureECMWF, h.HumidityECMWF, h.DewPointECMWF, h.PrecipitationECMWF, h.PrecipProbabilityECMWF, h.WindSpeedECMWF, h.WindDirectionECMWF},
		{"gfs", h.TemperatureGFS, h.HumidityGFS, h.DewPointGFS, h.PrecipitationGFS, h.PrecipProbabilityGFS, h.WindSpeedGFS, h.WindDirectionGFS},
		{"icon", h.TemperatureICON, h.HumidityICON, h.DewPointICON, h.PrecipitationICON, h.PrecipProbabilityICON, h.WindSpeedICON, h.WindDirectionICON},
	}

	var readings []model.Reading
	for _, v := range views {
		sourceID, ok := sourceIDs[v.sourceName]
		if !ok {
			return nil, fmt.Errorf("no sources row for %q — seed the sources table first", v.sourceName)
		}

		for i, t := range h.Time {
			readings = append(readings, model.Reading{
				SourceID:                 sourceID,
				LocationID:               loc.ID,
				ValidTime:                t,
				IssuedAt:                 issuedAt,
				FetchedAt:                fetchedAt,
				Temperature:              v.temperature[i],
				Humidity:                 v.humidity[i],
				DewPoint:                 v.dewPoint[i],
				PrecipitationMM:          v.precip[i],
				PrecipitationProbability: v.precipProb[i],
				WindSpeed:                v.windSpeed[i],
				WindDirection:            v.windDir[i],
			})
		}
	}
	return readings, nil
}
