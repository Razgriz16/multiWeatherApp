package fetch

import (
	"encoding/json"
	"fmt"
	"multiweather/model"
	"net/http"
	"strings"
)

const hourlyVars = "temperature_2m,relative_humidity_2m,dew_point_2m," +
	"precipitation,precipitation_probability,wind_speed_10m,wind_direction_10m"

func FetchForecast(loc model.Location, models []string) (forecastResponse, error) {
	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&hourly=%s&models=%s",
		loc.Latitude, loc.Longitude, hourlyVars, strings.Join(models, ","),
	)

	resp, err := http.Get(url)
	if err != nil {
		return forecastResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return forecastResponse{}, fmt.Errorf("open-meteo returned status %d", resp.StatusCode)
	}

	var forecast forecastResponse
	if err := json.NewDecoder(resp.Body).Decode(&forecast); err != nil {
		return forecastResponse{}, err
	}
	return forecast, nil
}
