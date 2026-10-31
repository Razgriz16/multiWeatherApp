package web

import (
	"fmt"
	"time"

	"multiweather/database"
)

// dayCard is one calendar day's rolled-up summary, shared by both the
// hero (detailed) and strip (summary) card renderings — the template
// decides which fields to show.
type dayCard struct {
	Date                     string // YYYY-MM-DD
	Label                    string // "Today", "Tomorrow", or weekday name
	Emoji                    string
	Condition                string
	MaxTemp                  string
	MinTemp                  string
	Humidity                 string // average across the day's hours
	PrecipitationMM          string // total across the day's hours
	PrecipitationProbability string // peak chance across the day's hours
	Selected                 bool
}

// rollupDaily groups hourly forecast rows by calendar date and reduces
// each day to one summary: max/min temperature, average humidity, total
// precipitation, peak precipitation chance, and the most frequent
// condition code (ties broken toward the more severe WMO code, same
// rule aggregates.consensusCondition uses across sources — here applied
// across a day's hours instead).
func rollupDaily(hours []database.ForecastHour) []dayCard {
	type dayAgg struct {
		maxTemp, minTemp, maxPrecipProb *float64
		humiditySum, precipSum          float64
		humidityCount, precipCount      int
		codeCounts                      map[int64]int
		codeDesc                        map[int64]string
	}

	var order []string
	byDate := make(map[string]*dayAgg)

	for _, h := range hours {
		date := h.ValidTime[:10]
		agg, ok := byDate[date]
		if !ok {
			agg = &dayAgg{codeCounts: map[int64]int{}, codeDesc: map[int64]string{}}
			byDate[date] = agg
			order = append(order, date)
		}

		if h.Temperature != nil {
			if agg.maxTemp == nil || *h.Temperature > *agg.maxTemp {
				agg.maxTemp = h.Temperature
			}
			if agg.minTemp == nil || *h.Temperature < *agg.minTemp {
				agg.minTemp = h.Temperature
			}
		}
		if h.Humidity != nil {
			agg.humiditySum += *h.Humidity
			agg.humidityCount++
		}
		if h.PrecipitationMM != nil {
			agg.precipSum += *h.PrecipitationMM
			agg.precipCount++
		}
		if h.PrecipitationProbability != nil {
			if agg.maxPrecipProb == nil || *h.PrecipitationProbability > *agg.maxPrecipProb {
				agg.maxPrecipProb = h.PrecipitationProbability
			}
		}
		if h.ConditionCode != nil {
			agg.codeCounts[*h.ConditionCode]++
			if h.Condition != nil {
				agg.codeDesc[*h.ConditionCode] = *h.Condition
			}
		}
	}

	days := make([]dayCard, len(order))
	for i, date := range order {
		agg := byDate[date]
		code, ok := dominantCode(agg.codeCounts)

		var avgHumidity *float64
		if agg.humidityCount > 0 {
			avg := agg.humiditySum / float64(agg.humidityCount)
			avgHumidity = &avg
		}
		var precipTotal *float64
		if agg.precipCount > 0 {
			sum := agg.precipSum
			precipTotal = &sum
		}

		days[i] = dayCard{
			Date:                     date,
			Label:                    dayLabel(i, date),
			Emoji:                    conditionEmoji(code, ok),
			Condition:                agg.codeDesc[code],
			MaxTemp:                  formatUnit(agg.maxTemp, "%.0f°C"),
			MinTemp:                  formatUnit(agg.minTemp, "%.0f°C"),
			Humidity:                 formatUnit(avgHumidity, "%.0f%%"),
			PrecipitationMM:          formatUnit(precipTotal, "%.1fmm"),
			PrecipitationProbability: formatUnit(agg.maxPrecipProb, "%.0f%%"),
		}
	}
	return days
}

// dominantCode picks the most frequent condition code, breaking ties
// toward the higher (more severe) WMO code.
func dominantCode(counts map[int64]int) (code int64, ok bool) {
	bestCount := -1
	for c, n := range counts {
		if n > bestCount || (n == bestCount && c > code) {
			code, bestCount, ok = c, n, true
		}
	}
	return code, ok
}

func dayLabel(index int, date string) string {
	switch index {
	case 0:
		return "Today"
	case 1:
		return "Tomorrow"
	default:
		t, err := time.Parse("2006-01-02", date)
		if err != nil {
			return date
		}
		return t.Format("Mon")
	}
}

// conditionEmoji maps an Open-Meteo WMO weather_code to a representative
// emoji. Ranges follow schema.sql's seeded (code, description) rows.
func conditionEmoji(code int64, ok bool) string {
	if !ok {
		return "❓" // ❓
	}
	switch {
	case code == 0:
		return "☀️" // ☀️ clear sky
	case code == 1:
		return "\U0001f324️" // 🌤️ mainly clear
	case code <= 3:
		return "⛅" // ⛅ partly cloudy / overcast
	case code == 45 || code == 48:
		return "\U0001f32b️" // 🌫️ fog
	case code >= 51 && code <= 57:
		return "\U0001f326️" // 🌦️ drizzle
	case code >= 61 && code <= 67:
		return "\U0001f327️" // 🌧️ rain
	case code >= 71 && code <= 77:
		return "\U0001f328️" // 🌨️ snowfall
	case code >= 80 && code <= 82:
		return "\U0001f327️" // 🌧️ rain showers
	case code >= 85 && code <= 86:
		return "\U0001f328️" // 🌨️ snow showers
	case code >= 95:
		return "⛈️" // ⛈️ thunderstorm
	default:
		return "❓"
	}
}

// selectHero picks the day(s) for the detailed hero cards: the single
// matching day if selectedDate is set and found, otherwise today+tomorrow.
func selectHero(days []dayCard, selectedDate string) []dayCard {
	if selectedDate != "" {
		for _, d := range days {
			if d.Date == selectedDate {
				return []dayCard{d}
			}
		}
	}
	if len(days) > 2 {
		return days[:2]
	}
	return days
}

// markSelected returns a copy of days with Selected set on the one
// matching selectedDate, for highlighting in the strip.
func markSelected(days []dayCard, selectedDate string) []dayCard {
	marked := make([]dayCard, len(days))
	for i, d := range days {
		d.Selected = selectedDate != "" && d.Date == selectedDate
		marked[i] = d
	}
	return marked
}

func formatUnit(v *float64, format string) string {
	if v == nil {
		return "—" // —
	}
	return fmt.Sprintf(format, *v)
}

// hourCard is one hour's detail row for the day-detail page's expandable list.
type hourCard struct {
	Time                     string // "HH:MM"
	Emoji                    string
	Temperature              string
	Humidity                 string
	PrecipitationMM          string
	PrecipitationProbability string
	Condition                string
}

// hoursForDate filters hours down to those falling on a single calendar date.
func hoursForDate(hours []database.ForecastHour, date string) []database.ForecastHour {
	var matched []database.ForecastHour
	for _, h := range hours {
		if len(h.ValidTime) >= 10 && h.ValidTime[:10] == date {
			matched = append(matched, h)
		}
	}
	return matched
}

func buildHourCards(hours []database.ForecastHour) []hourCard {
	cards := make([]hourCard, len(hours))
	for i, h := range hours {
		var code int64
		var ok bool
		if h.ConditionCode != nil {
			code, ok = *h.ConditionCode, true
		}
		var condition string
		if h.Condition != nil {
			condition = *h.Condition
		}
		cards[i] = hourCard{
			Time:                     hourLabel(h.ValidTime),
			Emoji:                    conditionEmoji(code, ok),
			Temperature:              formatUnit(h.Temperature, "%.0f°C"),
			Humidity:                 formatUnit(h.Humidity, "%.0f%%"),
			PrecipitationMM:          formatUnit(h.PrecipitationMM, "%.1fmm"),
			PrecipitationProbability: formatUnit(h.PrecipitationProbability, "%.0f%%"),
			Condition:                condition,
		}
	}
	return cards
}

// hourLabel extracts "HH:MM" from an ISO-ish "YYYY-MM-DDTHH:MM" valid_time.
func hourLabel(validTime string) string {
	if len(validTime) < 16 {
		return validTime
	}
	return validTime[11:16]
}
