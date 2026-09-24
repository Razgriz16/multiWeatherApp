package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Shape of the JSON Open-Meteo returns. Struct tags tell encoding/json
// which key maps to which field — this struct IS the parser, no
// separate parseJSON step needed.
type forecastResponse struct {
	Hourly struct {
		Time             []string  `json:"time"`
		TemperatureECMWF []float64 `json:"temperature_2m_ecmwf_ifs025"`
		TemperatureGFS   []float64 `json:"temperature_2m_gfs_seamless"`
		TemperatureICON  []float64 `json:"temperature_2m_icon_seamless"`
	} `json:"hourly"`
}

func main() {
	url := "https://api.open-meteo.com/v1/forecast?latitude=52.52&longitude=13.41&hourly=temperature_2m&models=ecmwf_ifs025,gfs_seamless,icon_seamless"

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

	for i, t := range forecast.Hourly.Time {
		fmt.Printf("%s: ecmwf=%.1f°C gfs=%.1f°C icon=%.1f°C\n",
			t,
			forecast.Hourly.TemperatureECMWF[i],
			forecast.Hourly.TemperatureGFS[i],
			forecast.Hourly.TemperatureICON[i])
	}
}
