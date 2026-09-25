package main

func main() {
	locations := []location{
		{Name: "Berlin", Latitude: 52.52, Longitude: 13.41},
		{Name: "Santiago", Latitude: -33.45, Longitude: -70.65},
	}

	for _, loc := range locations {
		fetchForecast(loc)
	}
}
