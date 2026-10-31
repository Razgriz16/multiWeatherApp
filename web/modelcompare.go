package web

import (
	"sort"

	"multiweather/database"
)

// chartSeries is one line in the model-comparison chart. Temps is nil
// where that source had no reading for the hour at the same index in
// chartData.Labels — Chart.js renders that as a gap, not a dip to zero.
type chartSeries struct {
	Name  string     `json:"name"`
	Temps []*float64 `json:"temps"`
}

// chartData is everything the client-side Chart.js widget needs, shaped
// so every series lines up against the same hour labels.
type chartData struct {
	Labels []string      `json:"labels"`
	Series []chartSeries `json:"series"`
}

// buildModelComparison aligns the aggregate's and each source's hourly
// temperatures to one shared set of hour labels (taken from the
// aggregate, which is already complete and chronologically ordered),
// so every line in the chart plots against the same x-axis.
func buildModelComparison(aggregateHours []database.ForecastHour, modelReadings []database.ModelReading) chartData {
	labels := make([]string, len(aggregateHours))
	indexByValidTime := make(map[string]int, len(aggregateHours))
	aggTemps := make([]*float64, len(aggregateHours))
	for i, h := range aggregateHours {
		labels[i] = hourLabel(h.ValidTime)
		indexByValidTime[h.ValidTime] = i
		aggTemps[i] = h.Temperature
	}

	tempsBySource := map[string][]*float64{}
	var sources []string
	for _, r := range modelReadings {
		temps, ok := tempsBySource[r.Source]
		if !ok {
			temps = make([]*float64, len(labels))
			tempsBySource[r.Source] = temps
			sources = append(sources, r.Source)
		}
		if i, ok := indexByValidTime[r.ValidTime]; ok {
			temps[i] = r.Temperature
		}
	}
	sort.Strings(sources)

	series := make([]chartSeries, 0, len(sources)+1)
	series = append(series, chartSeries{Name: "Aggregate", Temps: aggTemps})
	for _, name := range sources {
		series = append(series, chartSeries{Name: name, Temps: tempsBySource[name]})
	}

	return chartData{Labels: labels, Series: series}
}
