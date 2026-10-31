package web

import (
	"strings"
	"testing"

	"multiweather/database"
)

func tempHour(temp float64) database.ForecastHour {
	return database.ForecastHour{Temperature: new(temp)}
}

func TestBuildSparkline_EmptyWhenFewerThanTwoHours(t *testing.T) {
	if got := buildSparkline([]database.ForecastHour{tempHour(10)}); got != "" {
		t.Errorf("expected empty sparkline for 1 hour, got %q", got)
	}
}

func TestBuildSparkline_EmptyWhenNoTemperatureData(t *testing.T) {
	hours := []database.ForecastHour{{}, {}}
	if got := buildSparkline(hours); got != "" {
		t.Errorf("expected empty sparkline when all temperatures are nil, got %q", got)
	}
}

func TestBuildSparkline_TicksMatchDayMinMax(t *testing.T) {
	hours := []database.ForecastHour{tempHour(5), tempHour(20), tempHour(12)}

	svg := buildSparkline(hours)
	if !strings.Contains(svg, ">20</text>") {
		t.Errorf("expected max tick label 20 in svg, got %q", svg)
	}
	if !strings.Contains(svg, ">5</text>") {
		t.Errorf("expected min tick label 5 in svg, got %q", svg)
	}
}

func TestBuildSparkline_HasEvenlySpacedYTicks(t *testing.T) {
	hours := []database.ForecastHour{tempHour(0), tempHour(20)}

	svg := buildSparkline(hours)
	got := strings.Count(svg, `class="axis-label axis-label-y"`)
	if got != yTickCount {
		t.Errorf("expected %d y-axis ticks, got %d in %q", yTickCount, got, svg)
	}
	for _, want := range []string{">0</text>", ">5</text>", ">10</text>", ">15</text>", ">20</text>"} {
		if !strings.Contains(svg, want) {
			t.Errorf("expected tick label %q in svg, got %q", want, svg)
		}
	}
}

func TestBuildSparkline_IncludesAxisTitles(t *testing.T) {
	hours := []database.ForecastHour{tempHour(5), tempHour(20)}

	svg := buildSparkline(hours)
	if !strings.Contains(svg, "Hour of day") {
		t.Errorf("expected x-axis title in svg, got %q", svg)
	}
	if !strings.Contains(svg, "Temperature (°C)") {
		t.Errorf("expected y-axis title in svg, got %q", svg)
	}
}

func TestBuildSparkline_SkipsMissingHoursButStillDraws(t *testing.T) {
	hours := []database.ForecastHour{tempHour(10), {}, tempHour(20)}

	svg := buildSparkline(hours)
	if !strings.Contains(svg, "<path") {
		t.Errorf("expected a drawn path despite one missing hour, got %q", svg)
	}
}
