package aggregates

import (
	"testing"

	"multiweather/model"
)

func f(x float64) *float64 { return &x }
func c(x int64) *int64     { return &x }

func reading(sourceID int64, validTime string, temp, hum, precip *float64, condition *int64) model.Reading {
	return model.Reading{
		SourceID:        sourceID,
		LocationID:      1,
		ValidTime:       validTime,
		IssuedAt:        "issued",
		Temperature:     temp,
		Humidity:        hum,
		PrecipitationMM: precip,
		ConditionID:     condition,
	}
}

func TestComputeAggregates_MedianAndMax(t *testing.T) {
	readings := []model.Reading{
		reading(10, "h0", f(10), f(50), f(0), c(1)),
		reading(20, "h0", f(12), f(55), f(2), c(1)),
		reading(30, "h0", f(14), f(60), f(0.05), c(2)),
	}
	conditionCodes := map[int64]int64{1: 0, 2: 61} // id 1 = clear, id 2 = rain

	out := ComputeAggregates(readings, conditionCodes, "computed")
	if len(out) != 1 {
		t.Fatalf("expected 1 aggregate row, got %d", len(out))
	}
	agg := out[0]

	if agg.Temperature == nil || *agg.Temperature != 12 {
		t.Errorf("expected median temperature 12, got %v", agg.Temperature)
	}
	if agg.Humidity == nil || *agg.Humidity != 55 {
		t.Errorf("expected median humidity 55, got %v", agg.Humidity)
	}
	if agg.PrecipitationMM == nil || *agg.PrecipitationMM != 2 {
		t.Errorf("expected max precipitation 2, got %v", agg.PrecipitationMM)
	}
	// 2 of 3 sources are at/above the 0.1mm threshold (0 and 2; 0.05 doesn't count)
	if agg.PrecipitationProbability == nil || *agg.PrecipitationProbability != 1.0/3.0 {
		t.Errorf("expected precipitation probability 1/3, got %v", agg.PrecipitationProbability)
	}
	if agg.SourceCount != 3 {
		t.Errorf("expected source count 3, got %d", agg.SourceCount)
	}
	// id 1 (clear) has 2 votes vs id 2 (rain) with 1 vote
	if agg.ConditionID == nil || *agg.ConditionID != 1 {
		t.Errorf("expected majority condition id 1, got %v", agg.ConditionID)
	}
}

func TestComputeAggregates_SkipsBelowMinCriticalSources(t *testing.T) {
	readings := []model.Reading{
		reading(10, "h0", f(10), f(50), nil, nil),
		reading(20, "h0", nil, f(55), nil, nil), // missing temperature
		reading(30, "h0", nil, nil, nil, nil),   // missing both
	}

	out := ComputeAggregates(readings, nil, "computed")
	if len(out) != 0 {
		t.Fatalf("expected the hour to be skipped (only 1 valid temperature), got %d rows", len(out))
	}
}

func TestComputeAggregates_ConditionTieBreaksBySeverity(t *testing.T) {
	readings := []model.Reading{
		reading(10, "h0", f(10), f(50), nil, c(1)), // clear
		reading(20, "h0", f(11), f(51), nil, c(2)), // rain
		reading(30, "h0", f(12), f(52), nil, c(3)), // thunderstorm
	}
	conditionCodes := map[int64]int64{1: 0, 2: 61, 3: 95}

	out := ComputeAggregates(readings, conditionCodes, "computed")
	if len(out) != 1 {
		t.Fatalf("expected 1 aggregate row, got %d", len(out))
	}
	// 3-way tie (1 vote each) breaks toward the highest WMO code: id 3 (95, thunderstorm)
	if out[0].ConditionID == nil || *out[0].ConditionID != 3 {
		t.Errorf("expected tie-break to pick id 3 (most severe), got %v", out[0].ConditionID)
	}
}

func TestComputeAggregates_GroupsByValidTimeIndependently(t *testing.T) {
	readings := []model.Reading{
		reading(10, "h0", f(10), f(50), nil, nil),
		reading(20, "h0", f(12), f(55), nil, nil),
		reading(10, "h1", f(20), f(40), nil, nil),
		reading(20, "h1", f(22), f(45), nil, nil),
	}

	out := ComputeAggregates(readings, nil, "computed")
	if len(out) != 2 {
		t.Fatalf("expected 2 aggregate rows (one per hour), got %d", len(out))
	}
	if out[0].ValidTime != "h0" || out[1].ValidTime != "h1" {
		t.Errorf("expected rows ordered h0, h1, got %v, %v", out[0].ValidTime, out[1].ValidTime)
	}
}
