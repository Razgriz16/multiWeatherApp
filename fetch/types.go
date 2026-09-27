package fetch

type forecastResponse struct {
	Hourly struct {
		Time             []string  `json:"time"`
		TemperatureECMWF []float64 `json:"temperature_2m_ecmwf_ifs025"`
		TemperatureGFS   []float64 `json:"temperature_2m_gfs_seamless"`
		TemperatureICON  []float64 `json:"temperature_2m_icon_seamless"`
	} `json:"hourly"`
}
