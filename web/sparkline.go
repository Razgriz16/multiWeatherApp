package web

import (
	"fmt"
	"strings"

	"multiweather/database"
)

const (
	sparkWidth  = 640
	sparkHeight = 200
)

// Plot area: the rectangle the temperature line is actually drawn in.
// Everything outside it (axis lines, tick labels, titles) is margin.
const (
	plotLeft   float64 = 56  // room for y-axis tick numbers + title
	plotRight  float64 = 620
	plotTop    float64 = 20
	plotBottom float64 = 150 // room below for x-axis tick labels + title

	xAxisTickLen float64 = 4
	xAxisLabelY  float64 = plotBottom + 16
	xAxisTitleY  float64 = plotBottom + 36

	yAxisTickLen float64 = 4
	yAxisTitleX  float64 = 14
	yTickCount   int     = 5 // min, max, and 3 evenly spaced steps between
)

// buildSparkline renders an hourly temperature curve as inline SVG:
//   - x-axis: hour of day (left = earliest hour, right = latest)
//   - y-axis: temperature in °C, with the day's own min/max as dashed
//     reference lines so the curve's highs/lows are easy to read off
//
// Returns "" when there isn't enough temperature data to draw a line.
func buildSparkline(hours []database.ForecastHour) string {
	n := len(hours)
	if n < 2 {
		return ""
	}

	var minTemp, maxTemp float64
	haveRange := false
	for _, h := range hours {
		if h.Temperature == nil {
			continue
		}
		t := *h.Temperature
		if !haveRange {
			minTemp, maxTemp, haveRange = t, t, true
			continue
		}
		if t < minTemp {
			minTemp = t
		}
		if t > maxTemp {
			maxTemp = t
		}
	}
	if !haveRange {
		return ""
	}

	spread := maxTemp - minTemp
	if spread == 0 {
		spread = 1
	}
	plotW := plotRight - plotLeft
	plotH := plotBottom - plotTop

	yFor := func(temp float64) float64 { return plotTop + (maxTemp-temp)/spread*plotH }
	xFor := func(i int) float64 { return plotLeft + float64(i)/float64(n-1)*plotW }

	var path strings.Builder
	points := 0
	for i, h := range hours {
		if h.Temperature == nil {
			continue
		}
		x, y := xFor(i), yFor(*h.Temperature)
		if points == 0 {
			fmt.Fprintf(&path, "M%.1f,%.1f", x, y)
		} else {
			fmt.Fprintf(&path, " L%.1f,%.1f", x, y)
		}
		points++
	}
	if points < 2 {
		return ""
	}

	var svg strings.Builder
	fmt.Fprintf(&svg, `<svg viewBox="0 0 %d %d" class="sparkline" xmlns="http://www.w3.org/2000/svg">`, sparkWidth, sparkHeight)

	// Axis lines (the plot's left and bottom edges).
	fmt.Fprintf(&svg, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" class="axis" />`, plotLeft, plotTop, plotLeft, plotBottom)
	fmt.Fprintf(&svg, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" class="axis" />`, plotLeft, plotBottom, plotRight, plotBottom)

	// Axis titles explaining what each axis means.
	fmt.Fprintf(&svg, `<text x="%.1f" y="%.1f" class="axis-title">Hour of day</text>`, (plotLeft+plotRight)/2, xAxisTitleY)
	fmt.Fprintf(&svg, `<text x="%.1f" y="%.1f" class="axis-title" transform="rotate(-90 %.1f %.1f)">Temperature (°C)</text>`,
		yAxisTitleX, (plotTop+plotBottom)/2, yAxisTitleX, (plotTop+plotBottom)/2)

	// X-axis ticks: a handful of evenly spaced hours, labelled "HHh".
	for _, idx := range xTickIndices(n) {
		x := xFor(idx)
		label := hourTickLabel(hours[idx].ValidTime)
		fmt.Fprintf(&svg, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" class="axis" />`, x, plotBottom, x, plotBottom+xAxisTickLen)
		fmt.Fprintf(&svg, `<text x="%.1f" y="%.1f" class="axis-label">%s</text>`, x, xAxisLabelY, label)
	}

	// Y-axis ticks: min, max, and evenly spaced steps between, each with
	// its own gridline so you can read a value off the curve without
	// guessing between just two extremes.
	for _, temp := range yTickValues(minTemp, maxTemp, yTickCount) {
		y := yFor(temp)
		fmt.Fprintf(&svg, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" class="axis" />`, plotLeft-yAxisTickLen, y, plotLeft, y)
		fmt.Fprintf(&svg, `<text x="%.1f" y="%.1f" class="axis-label axis-label-y">%.0f</text>`, plotLeft-yAxisTickLen-4, y+4, temp)
		fmt.Fprintf(&svg, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" class="ref" />`, plotLeft, y, plotRight, y)
	}

	fmt.Fprintf(&svg, `<path d="%s" class="temp-line" />`, path.String())
	svg.WriteString(`</svg>`)
	return svg.String()
}

// xTickIndices picks up to 5 evenly spaced hour indices to label on the
// x-axis, so the chart doesn't try to cram a label under every hour.
func xTickIndices(n int) []int {
	const want = 5
	if n <= want {
		idx := make([]int, n)
		for i := range idx {
			idx[i] = i
		}
		return idx
	}
	idx := make([]int, want)
	for i := range idx {
		idx[i] = i * (n - 1) / (want - 1)
	}
	return idx
}

// yTickValues returns count evenly spaced temperatures from minTemp to
// maxTemp (inclusive), e.g. count=5 gives min, +25%, +50%, +75%, max.
func yTickValues(minTemp, maxTemp float64, count int) []float64 {
	if count < 2 {
		count = 2
	}
	step := (maxTemp - minTemp) / float64(count-1)
	vals := make([]float64, count)
	for i := range vals {
		vals[i] = minTemp + float64(i)*step
	}
	return vals
}

// hourTickLabel extracts just the hour from a "YYYY-MM-DDTHH:MM" valid_time.
func hourTickLabel(validTime string) string {
	if len(validTime) < 13 {
		return ""
	}
	return validTime[11:13] + "h"
}
