package fetch

// One field per (variable, model) pair. models is fixed for v1
// (docs/PROJECT-RESEARCH.md §5.1/§7) — widening it means adding fields
// here too, it won't decode an unlisted model on its own.
type forecastResponse struct {
	Hourly struct {
		Time []string `json:"time"`

		TemperatureECMWF []*float64 `json:"temperature_2m_ecmwf_ifs025"`
		TemperatureGFS   []*float64 `json:"temperature_2m_gfs_seamless"`
		TemperatureICON  []*float64 `json:"temperature_2m_icon_seamless"`

		HumidityECMWF []*float64 `json:"relative_humidity_2m_ecmwf_ifs025"`
		HumidityGFS   []*float64 `json:"relative_humidity_2m_gfs_seamless"`
		HumidityICON  []*float64 `json:"relative_humidity_2m_icon_seamless"`

		DewPointECMWF []*float64 `json:"dew_point_2m_ecmwf_ifs025"`
		DewPointGFS   []*float64 `json:"dew_point_2m_gfs_seamless"`
		DewPointICON  []*float64 `json:"dew_point_2m_icon_seamless"`

		PrecipitationECMWF []*float64 `json:"precipitation_ecmwf_ifs025"`
		PrecipitationGFS   []*float64 `json:"precipitation_gfs_seamless"`
		PrecipitationICON  []*float64 `json:"precipitation_icon_seamless"`

		PrecipProbabilityECMWF []*float64 `json:"precipitation_probability_ecmwf_ifs025"`
		PrecipProbabilityGFS   []*float64 `json:"precipitation_probability_gfs_seamless"`
		PrecipProbabilityICON  []*float64 `json:"precipitation_probability_icon_seamless"`

		WindSpeedECMWF []*float64 `json:"wind_speed_10m_ecmwf_ifs025"`
		WindSpeedGFS   []*float64 `json:"wind_speed_10m_gfs_seamless"`
		WindSpeedICON  []*float64 `json:"wind_speed_10m_icon_seamless"`

		WindDirectionECMWF []*float64 `json:"wind_direction_10m_ecmwf_ifs025"`
		WindDirectionGFS   []*float64 `json:"wind_direction_10m_gfs_seamless"`
		WindDirectionICON  []*float64 `json:"wind_direction_10m_icon_seamless"`

		WeatherCodeECMWF []*int64 `json:"weather_code_ecmwf_ifs025"`
		WeatherCodeGFS   []*int64 `json:"weather_code_gfs_seamless"`
		WeatherCodeICON  []*int64 `json:"weather_code_icon_seamless"`
	} `json:"hourly"`
}
