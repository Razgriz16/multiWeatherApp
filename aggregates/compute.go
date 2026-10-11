package aggregates

import (
	"sort"

	"multiweather/model"
)

// Below this many sources with a valid critical field (temperature,
// humidity), the hour's forecast isn't trustworthy enough to publish —
// skip the aggregate row entirely rather than write a one-source "median"
// (docs/SCHEMA-ANALYSIS.md, aggregates).
const minCriticalSources = 2

// Placeholder until the calibration study (docs/PROJECT-RESEARCH.md §6.2)
// picks a location-specific value: a source counts as "precipitating" if
// it forecasts at least this much.
const precipThresholdMM = 0.1

type groupKey struct {
	locationID int64
	validTime  string
	issuedAt   string
}

// ComputeAggregates collapses one ingest run's per-source readings into one
// cross-source consensus row per (location, valid_time, issued_at). Pure
// function over already-fetched readings — no I/O, no randomness.
//
// conditionCodes maps a conditions.id (Reading.ConditionID) back to its WMO
// code, needed only to break a condition vote tie by severity — the id
// itself is an arbitrary row number, not an indicator of severity.
func ComputeAggregates(readings []model.Reading, conditionCodes map[int64]int64, computedAt string) []model.Aggregate {
	groups := make(map[groupKey][]model.Reading)
	var order []groupKey
	for _, r := range readings {
		k := groupKey{r.LocationID, r.ValidTime, r.IssuedAt}
		if _, ok := groups[k]; !ok {
			order = append(order, k)
		}
		groups[k] = append(groups[k], r)
	}
	sort.Slice(order, func(i, j int) bool {
		if order[i].locationID != order[j].locationID {
			return order[i].locationID < order[j].locationID
		}
		if order[i].issuedAt != order[j].issuedAt {
			return order[i].issuedAt < order[j].issuedAt
		}
		return order[i].validTime < order[j].validTime
	})

	var out []model.Aggregate
	for _, k := range order {
		group := groups[k]

		temps := nonNil(pluck(group, func(r model.Reading) *float64 { return r.Temperature }))
		hums := nonNil(pluck(group, func(r model.Reading) *float64 { return r.Humidity }))
		if len(temps) < minCriticalSources || len(hums) < minCriticalSources {
			continue
		}

		precips := nonNil(pluck(group, func(r model.Reading) *float64 { return r.PrecipitationMM }))

		out = append(out, model.Aggregate{
			LocationID:               k.locationID,
			ValidTime:                k.validTime,
			IssuedAt:                 k.issuedAt,
			Temperature:              medianPtr(temps),
			Humidity:                 medianPtr(hums),
			DewPoint:                 medianPtr(nonNil(pluck(group, func(r model.Reading) *float64 { return r.DewPoint }))),
			PrecipitationMM:          maxPtr(precips),
			PrecipitationProbability: precipProbability(precips),
			ConditionID:              consensusCondition(group, conditionCodes),
			SourceCount:              int64(len(temps)),
			ComputedAt:               computedAt,
		})
	}
	return out
}

func pluck(group []model.Reading, field func(model.Reading) *float64) []*float64 {
	vals := make([]*float64, len(group))
	for i, r := range group {
		vals[i] = field(r)
	}
	return vals
}

func nonNil(vals []*float64) []float64 {
	out := make([]float64, 0, len(vals))
	for _, v := range vals {
		if v != nil {
			out = append(out, *v)
		}
	}
	return out
}

func medianPtr(vals []float64) *float64 {
	if len(vals) == 0 {
		return nil
	}
	sorted := append([]float64(nil), vals...)
	sort.Float64s(sorted)
	n := len(sorted)
	var m float64
	if n%2 == 1 {
		m = sorted[n/2]
	} else {
		m = (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return &m
}

func maxPtr(vals []float64) *float64 {
	if len(vals) == 0 {
		return nil
	}
	m := vals[0]
	for _, v := range vals[1:] {
		if v > m {
			m = v
		}
	}
	return &m
}

// precipProbability is the ensemble-threshold fraction of sources
// forecasting at least precipThresholdMM — not a median, and not an
// average of the sources' own probability fields (docs/SCHEMA-ANALYSIS.md).
func precipProbability(precips []float64) *float64 {
	if len(precips) == 0 {
		return nil
	}
	var wet int
	for _, p := range precips {
		if p >= precipThresholdMM {
			wet++
		}
	}
	prob := float64(wet) / float64(len(precips))
	return &prob
}

// consensusCondition majority-votes the sources' condition ids. A tie
// breaks toward the numerically highest WMO code — Open-Meteo's table
// escalates roughly by severity within each weather family (thunderstorms
// 95-99 highest, clear sky 0 lowest).
func consensusCondition(group []model.Reading, conditionCodes map[int64]int64) *int64 {
	counts := make(map[int64]int)
	for _, r := range group {
		if r.ConditionID != nil {
			counts[*r.ConditionID]++
		}
	}
	if len(counts) == 0 {
		return nil
	}

	var best int64
	bestCount := -1
	for id, count := range counts {
		if count > bestCount || (count == bestCount && conditionCodes[id] > conditionCodes[best]) {
			best, bestCount = id, count
		}
	}
	return &best
}
