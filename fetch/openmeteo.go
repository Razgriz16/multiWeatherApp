package main

import (
	"encoding/json"
	"fmt"
	"multiweather/model"
	"net/http"
)

func fetchForecast(loc model.Location) {
	url := fmt.Sprintf(
		"https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&hourly=temperature_2m&models=ecmwf_ifs025,gfs_seamless,icon_seamless",
		loc.Latitude, loc.Longitude,
	)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("request failed:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("open-meteo returned status:", resp.StatusCode)
		return
	}

	var forecast forecastResponse
	if err := json.NewDecoder(resp.Body).Decode(&forecast); err != nil {
		fmt.Println("decode failed:", err)
		return
	}

	fmt.Println("---", loc.Name, "---")
	for i, t := range forecast.Hourly.Time {
		fmt.Printf("%s: ecmwf=%.1f°C gfs=%.1f°C icon=%.1f°C\n",
			t,
			forecast.Hourly.TemperatureECMWF[i],
			forecast.Hourly.TemperatureGFS[i],
			forecast.Hourly.TemperatureICON[i])
	}
}
